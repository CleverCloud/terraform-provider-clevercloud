// Package statestore implements a Terraform state store for Clever Cloud.
//
// All state persistence and lock management is delegated to the Clever Cloud
// backend identified by the addon created in Initialize. This package holds
// only the configuration and the reference to that addon — it does not cache
// state or locks locally.
//
// State store support in terraform-plugin-framework is experimental and the
// API may change.
package statestore

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	fwstatestore "github.com/hashicorp/terraform-plugin-framework/statestore"
	"github.com/hashicorp/terraform-plugin-framework/statestore/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/provider"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
	"go.clever-cloud.dev/client"
)

// defaultLockTTL is the lock duration applied when the user does not configure
// lock_duration. Long-running applies should set their own value.
const defaultLockTTL = 30 * time.Minute

func New() fwstatestore.StateStore {
	return &StateStore{}
}

type StateStore struct {
	cfg     Config
	lockTTL time.Duration

	// prov is the configured Clever Cloud provider, captured during Initialize
	// so backend calls (Read/Write/Lock/...) can reach the API client.
	prov provider.Provider

	// addonID is the RealID returned by CreateAddon, identifying the backing
	// addon for every subsequent backend call.
	addonID string
}

type Config struct {
	// Key identifies which Clever Cloud addon backs this state store.
	// CRITICAL: changing this value after the state has been written will
	// orphan the existing state — Terraform will look at a different addon
	// and find nothing. Treat it as immutable for the lifetime of the
	// project.
	Key          types.String `tfsdk:"key"`
	LockDuration types.String `tfsdk:"lock_duration"`
}

var (
	_ fwstatestore.StateStore              = (*StateStore)(nil)
	_ fwstatestore.StateStoreWithConfigure = (*StateStore)(nil)
)

func (s *StateStore) Metadata(_ context.Context, req fwstatestore.MetadataRequest, resp *fwstatestore.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_statestore"
}

func (s *StateStore) Schema(_ context.Context, _ fwstatestore.SchemaRequest, resp *fwstatestore.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Clever Cloud Terraform state store (placeholder, in-memory).",
		Attributes: map[string]schema.Attribute{
			"key": schema.StringAttribute{
				Description: "Identifier of the Clever Cloud addon backing this state store. " +
					"This value is used to locate the addon every time Terraform reads or " +
					"writes its state.\n\n" +
					"**Do not change this value once the state has been written.** Terraform " +
					"determines where to look for an existing state purely from this key — " +
					"changing it will not migrate the state, it will simply make Terraform " +
					"point at a different (likely empty) location, and any previously stored " +
					"state will be effectively lost from Terraform's point of view. " +
					"If you need to move the state, copy it manually before changing the key.",
				Required: true,
			},
			"lock_duration": schema.StringAttribute{
				Description: "Maximum lifetime of a state lock (e.g. \"10m\", \"1h\"). " +
					"Once elapsed, the lock is considered stale and a new Lock call replaces it. " +
					"Defaults to \"30m\". Use \"0\" to disable expiry.",
				Optional: true,
			},
		},
	}
}

// runtimeData is the payload threaded through Initialize → Configure. The
// framework calls Initialize once (during init) and Configure before every
// subsequent RPC, on potentially fresh StateStore instances. Anything the
// other methods need to function (client, addon ID, parsed config) must
// travel through here.
type runtimeData struct {
	prov    provider.Provider
	cfg     Config
	lockTTL time.Duration
	addonID string
}

// Configure is invoked before every state store RPC. It hydrates the receiver
// from the data Initialize stashed in StateStoreData. Initialize itself does
// not populate s.* — only this method does, which keeps the two code paths
// consistent (first-RPC after Initialize and every later RPC on a fresh
// instance).
func (s *StateStore) Configure(_ context.Context, req fwstatestore.ConfigureRequest, resp *fwstatestore.ConfigureResponse) {
	if req.StateStoreData == nil {
		// Configure can be called before Initialize during config validation;
		// nothing to hydrate yet.
		return
	}
	rt, ok := req.StateStoreData.(runtimeData)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected state store data",
			"State store received an unexpected payload from Initialize. This is a bug in the provider.",
		)
		return
	}
	s.prov = rt.prov
	s.cfg = rt.cfg
	s.lockTTL = rt.lockTTL
	s.addonID = rt.addonID
}

func (s *StateStore) Initialize(ctx context.Context, req fwstatestore.InitializeRequest, resp *fwstatestore.InitializeResponse) {
	cfg := helper.From[Config](ctx, req.Config, &resp.Diagnostics)
	if resp.Diagnostics.HasError() || cfg == nil {
		return
	}

	ttl, d := parseLockTTL(cfg.LockDuration)
	resp.Diagnostics.Append(d)
	if resp.Diagnostics.HasError() {
		return
	}

	prov, ok := req.ProviderData.(provider.Provider)
	if !ok {
		resp.Diagnostics.AddError(
			"Provider data unavailable",
			"State store cannot reach the configured Clever Cloud client; the provider must be configured before the state store is initialized.",
		)
		return
	}

	addonID, addonDiag := provisionAddon(ctx, prov, cfg.Key.ValueString())
	if addonDiag != nil {
		resp.Diagnostics.Append(addonDiag)
		return
	}

	resp.StateStoreData = runtimeData{
		prov:    prov,
		cfg:     *cfg,
		lockTTL: ttl,
		addonID: addonID,
	}

	tflog.Debug(ctx, "state store initialized", map[string]any{
		"key":      cfg.Key.ValueString(),
		"addon_id": addonID,
		"lock_ttl": ttl.String(),
	})
}

// addonProviderID is the Clever Cloud addon provider used to back state stores.
const addonProviderID = "terraform-state-store"

// provisionAddon returns the addon ID for the state store identified by key.
// If an addon with this key already exists in the organization, it is reused;
// otherwise a new one is created. This makes Initialize idempotent across
// successive `terraform init` runs.
func provisionAddon(ctx context.Context, prov provider.Provider, key string) (string, diag.Diagnostic) {
	if existingID, d := findExistingAddon(ctx, prov, key); d != nil {
		return "", d
	} else if existingID != "" {
		tflog.Debug(ctx, "reusing existing state store addon", map[string]any{
			"key":      key,
			"addon_id": existingID,
		})
		return existingID, nil
	}

	providersRes := tmp.GetAddonsProviders(ctx, prov.Client())
	if providersRes.HasError() {
		return "", diag.NewErrorDiagnostic("Failed to list addon providers", providersRes.Error().Error())
	}

	addonProvider := pkg.LookupAddonProvider(*providersRes.Payload(), addonProviderID)
	if addonProvider == nil {
		return "", diag.NewErrorDiagnostic("Addon provider not found", addonProviderID+" is not available on this Clever Cloud account")
	}
	plan := addonProvider.FirstPlan()
	if plan == nil {
		return "", diag.NewErrorDiagnostic("No plan available", addonProviderID+" has no plans on this account")
	}

	createRes := tmp.CreateAddon(ctx, prov.Client(), prov.Organization(), tmp.AddonRequest{
		Name:       key,
		Plan:       plan.ID,
		Region:     "par",
		ProviderID: addonProviderID,
	})
	if createRes.HasError() {
		return "", diag.NewAttributeErrorDiagnostic(
			path.Root("key"),
			"Failed to create state store addon",
			createRes.Error().Error(),
		)
	}

	addonID := createRes.Payload().RealID
	tflog.Info(ctx, "state store addon created", map[string]any{
		"key":      key,
		"addon_id": addonID,
	})
	return addonID, nil
}

// findExistingAddon returns the RealID of an existing terraform-state-store
// addon whose Name matches key, or "" if none is found. Both the name and the
// provider ID are checked so an unrelated addon that happens to share the
// name is not picked up.
func findExistingAddon(ctx context.Context, prov provider.Provider, key string) (string, diag.Diagnostic) {
	listRes := tmp.ListAddons(ctx, prov.Client(), prov.Organization())
	if listRes.HasError() {
		return "", diag.NewErrorDiagnostic("Failed to list addons", listRes.Error().Error())
	}

	for _, a := range *listRes.Payload() {
		if a.Name == key && a.Provider.ID == addonProviderID {
			return a.RealID, nil
		}
	}
	return "", nil
}

func parseLockTTL(v types.String) (time.Duration, diag.Diagnostic) {
	if v.IsNull() || v.IsUnknown() || v.ValueString() == "" {
		return defaultLockTTL, nil
	}
	d, err := time.ParseDuration(v.ValueString())
	if err != nil {
		return 0, diag.NewAttributeErrorDiagnostic(
			path.Root("lock_duration"),
			"Invalid lock_duration",
			"Expected a Go duration (e.g. \"10m\", \"1h\"), got: "+err.Error(),
		)
	}
	if d < 0 {
		return 0, diag.NewAttributeErrorDiagnostic(
			path.Root("lock_duration"),
			"Invalid lock_duration",
			"lock_duration must be zero or positive",
		)
	}
	return d, nil
}

// getStatesResponsePayload mirrors what the backend returns from
// GET /v4/terraform/store/{addonID}/state — the list of state IDs that have
// a head pointer in latest_versions.
type getStatesResponsePayload struct {
	StateIDs []string `json:"state_ids"`
}

// GetStates returns the list of workspaces that have a current state in the
// backend addon by calling GET /v4/terraform/store/{addonID}/state.
func (s *StateStore) GetStates(ctx context.Context, _ fwstatestore.GetStatesRequest, resp *fwstatestore.GetStatesResponse) {
	path := fmt.Sprintf("/v4/terraform/store/%s/state", s.addonID)

	res := client.Get[getStatesResponsePayload](ctx, s.prov.Client(), path)
	if res.HasError() {
		resp.Diagnostics.AddError(
			"Failed to list states",
			res.Error().Error(),
		)
		return
	}

	resp.StateIDs = res.Payload().StateIDs

	tflog.Debug(ctx, "states listed", map[string]any{
		"addon_id": s.addonID,
		"count":    len(resp.StateIDs),
	})
}

// Read fetches the bytes of the current state for (addon, state) from the
// backend by calling GET /v4/terraform/store/{addonID}/state/{stateID}.
//
// A 404 means the state has never been written (or has been DeleteState'd).
// Terraform interprets an empty StateBytes as "no state", so we leave the
// response untouched without raising a diagnostic.
//
// PlainTextString is used to bypass the client's automatic JSON unmarshalling
// — the body is the state file itself and we want it as raw bytes.
func (s *StateStore) Read(ctx context.Context, req fwstatestore.ReadRequest, resp *fwstatestore.ReadResponse) {
	// "current" is a backend-side sentinel that resolves through
	// latest_versions to the head hash. The route also accepts a real sha256
	// hex for historical reads, but Terraform only ever wants the head.
	path := fmt.Sprintf("/v4/terraform/store/%s/state/%s/current",
		s.addonID, url.PathEscape(req.StateID))

	res := client.Get[client.PlainTextString](ctx, s.prov.Client(), path)
	if res.HasError() {
		if res.StatusCode() == http.StatusNotFound {
			return
		}
		resp.Diagnostics.AddError(
			"Failed to read state",
			res.Error().Error(),
		)
		return
	}

	resp.StateBytes = []byte(*res.Payload())

	tflog.Debug(ctx, "state read", map[string]any{
		"addon_id": s.addonID,
		"state_id": req.StateID,
		"bytes":    len(resp.StateBytes),
	})
}

// writeResponsePayload mirrors what the backend returns from
// PUT /state/{stateID}/versions — the content-addressed version ID and the
// row creation timestamp. We only consume it for logging.
type writeResponsePayload struct {
	VersionID string    `json:"version_id"`
	CreatedAt time.Time `json:"created_at"`
}

// Write persists the bytes of a state file to the backend addon by calling
// PUT /v4/terraform/store/{addonID}/state/{stateID}/versions with the raw
// state JSON as the request body.
//
// req.StateBytes is wrapped in json.RawMessage so the Clever Cloud client
// forwards the bytes verbatim instead of re-encoding them as a JSON string.
func (s *StateStore) Write(ctx context.Context, req fwstatestore.WriteRequest, resp *fwstatestore.WriteResponse) {
	path := fmt.Sprintf("/v4/terraform/store/%s/state/%s/versions",
		s.addonID, url.PathEscape(req.StateID))

	res := client.Put[writeResponsePayload](ctx, s.prov.Client(), path, json.RawMessage(req.StateBytes))

	if res.HasError() {
		resp.Diagnostics.AddError(
			"Failed to write state",
			res.Error().Error(),
		)
		return
	}

	tflog.Debug(ctx, "state written", map[string]any{
		"addon_id":   s.addonID,
		"state_id":   req.StateID,
		"version_id": res.Payload().VersionID,
		"bytes":      len(req.StateBytes),
	})
}

// DeleteState removes the state from Terraform's perspective by calling
// DELETE /v4/terraform/store/{addonID}/state/{stateID}/versions on the
// backend. Server-side this drops the head pointer but keeps historical
// state_versions rows for audit / recovery.
func (s *StateStore) DeleteState(ctx context.Context, req fwstatestore.DeleteStateRequest, resp *fwstatestore.DeleteStateResponse) {
	path := fmt.Sprintf("/v4/terraform/store/%s/state/%s",
		s.addonID, url.PathEscape(req.StateID))

	res := client.Delete[client.Nothing](ctx, s.prov.Client(), path)
	if res.HasError() {
		resp.Diagnostics.AddError(
			"Failed to delete state",
			res.Error().Error(),
		)
		return
	}

	tflog.Debug(ctx, "state deleted", map[string]any{
		"addon_id": s.addonID,
		"state_id": req.StateID,
	})
}

// lockRequestPayload is the body sent to the backend's PUT /lock endpoint.
// Both ends are Go, so time.Duration travels as its int64 nanosecond form.
// state_id travels in the URL path, not the body.
type lockRequestPayload struct {
	Duration  time.Duration `json:"duration"`
	Operation string        `json:"operation"`
}

type lockResponsePayload struct {
	LockID    string    `json:"lock_id"`
	ExpiresAt time.Time `json:"expires_at"`
}

// Lock acquires a lock for the given state by calling
// PUT /v4/terraform/store/{addonID}/lock on the backend addon. The configured
// lockTTL is sent so the backend can expire stale locks server-side.
//
// On HTTP 409 the backend signals that another client already holds a lock
// for this state; we report it as a workspace-already-locked diagnostic so
// Terraform aborts cleanly.
func (s *StateStore) Lock(ctx context.Context, req fwstatestore.LockRequest, resp *fwstatestore.LockResponse) {
	path := fmt.Sprintf("/v4/terraform/store/%s/state/%s/lock",
		s.addonID, url.PathEscape(req.StateID))

	res := client.Put[lockResponsePayload](ctx, s.prov.Client(), path, lockRequestPayload{
		Duration:  s.lockTTL,
		Operation: req.Operation,
	})

	if res.HasError() {
		if res.StatusCode() == http.StatusConflict {
			resp.Diagnostics.AddError(
				"Workspace Already Locked",
				fmt.Sprintf("State %q is already locked by another client.", req.StateID),
			)
			return
		}
		resp.Diagnostics.AddError("Failed to acquire state lock", res.Error().Error())
		return
	}

	resp.LockID = res.Payload().LockID

	tflog.Debug(ctx, "state lock acquired", map[string]any{
		"addon_id":   s.addonID,
		"state_id":   req.StateID,
		"lock_id":    resp.LockID,
		"expires_at": res.Payload().ExpiresAt,
	})
}

// Unlock releases the lock identified by req.LockID for the given state by
// calling DELETE /v4/terraform/store/{addonID}/lock/{lockID} on the backend.
//
// A 404 from the backend means the lock no longer exists — either it was
// already released or it expired server-side. We surface that as an error
// diagnostic so Terraform doesn't silently assume success.
func (s *StateStore) Unlock(ctx context.Context, req fwstatestore.UnlockRequest, resp *fwstatestore.UnlockResponse) {
	path := fmt.Sprintf("/v4/terraform/store/%s/state/%s/lock/%s",
		s.addonID, url.PathEscape(req.StateID), url.PathEscape(req.LockID))

	res := client.Delete[client.Nothing](ctx, s.prov.Client(), path)

	if res.HasError() {
		if res.StatusCode() == http.StatusNotFound {
			resp.Diagnostics.AddWarning(
				"State lock not found",
				fmt.Sprintf("No active lock %q for state %q on the backend; it may have already been released or expired.", req.LockID, req.StateID),
			)
			return
		}
		resp.Diagnostics.AddError("Failed to release state lock", res.Error().Error())
		return
	}

	tflog.Debug(ctx, "state lock released", map[string]any{
		"addon_id": s.addonID,
		"state_id": req.StateID,
		"lock_id":  req.LockID,
	})
}
