package kubernetes

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"go.clever-cloud.com/terraform-provider/pkg/resources"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

// runCreate runs Create against a fake API whose cluster answers with get once
// the creation request went through
func runCreate(t *testing.T, get http.HandlerFunc) (resource.CreateResponse, Kubernetes) {
	t.Helper()
	ctx := t.Context()

	cc := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(clusterPayload("DEPLOYING", "1.36"))
			return
		}
		get(w, r)
	})
	res := &ResourceKubernetes{}
	res.Provider = clusterReadTestProvider{client: cc}

	var s resource.SchemaResponse
	res.Schema(ctx, resource.SchemaRequest{}, &s)
	var is resource.IdentitySchemaResponse
	res.IdentitySchema(ctx, resource.IdentitySchemaRequest{}, &is)
	null := tftypes.NewValue(s.Schema.Type().TerraformType(ctx), nil)

	plan := tfsdk.Plan{Schema: s.Schema, Raw: null.Copy()}
	if d := plan.Set(ctx, Kubernetes{
		Name:                 types.StringValue("c"),
		NodeAutoprovisioning: types.BoolValue(false),
		Tags:                 resources.NullTags,
		Topology:             types.ObjectNull(topologyAttrTypes),
	}); d.HasError() {
		t.Fatalf("set plan: %v", d)
	}

	resp := resource.CreateResponse{
		State:    tfsdk.State{Schema: s.Schema, Raw: null.Copy()},
		Identity: &tfsdk.ResourceIdentity{Schema: is.IdentitySchema, Raw: tftypes.NewValue(is.IdentitySchema.Type().TerraformType(ctx), nil)},
	}
	res.Create(ctx, resource.CreateRequest{Plan: plan, Config: tfsdk.Config{Schema: s.Schema, Raw: plan.Raw}}, &resp)

	var state Kubernetes
	if !resp.State.Raw.IsNull() {
		if d := resp.State.Get(ctx, &state); d.HasError() {
			t.Fatalf("get state: %v", d)
		}
	}

	return resp, state
}

// TestCreateSavesAClusterThatFails pins that a cluster the API accepted stays in
// state: it still exists, and without it the next apply creates a second one
func TestCreateSavesAClusterThatFails(t *testing.T) {
	tests := []struct {
		name   string
		get    http.HandlerFunc
		detail string
	}{
		{
			name: "the deployment failed",
			get: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(clusterPayload("FAILED", "1.36"))
			},
			detail: "has to be resumed",
		},
		{
			// deleted from the console while it was deploying, this never ended
			name: "the cluster is being deleted",
			get: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(clusterPayload("DELETING", "1.36"))
			},
			detail: "cannot be resumed",
		},
		{
			// polling errors were ignored, this never ended either
			name:   "the cluster is gone",
			get:    func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNotFound) },
			detail: "does not exist any more",
		},
		{
			name:   "the credentials stopped working",
			get:    func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusForbidden) },
			detail: "credentials were refused",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, state := runCreate(t, tt.get)

			if resp.Diagnostics.ErrorsCount() != 1 {
				t.Fatalf("got %d errors, want 1: %v", resp.Diagnostics.ErrorsCount(), resp.Diagnostics)
			}
			if detail := resp.Diagnostics.Errors()[0].Detail(); !strings.Contains(detail, tt.detail) {
				t.Errorf("error %q, want it to contain %q", detail, tt.detail)
			}
			if state.ID.ValueString() != "cluster" {
				t.Errorf("the cluster is not in state: %+v", state)
			}
			if resp.Identity.Raw.IsNull() {
				t.Error("the cluster has no identity")
			}
		})
	}
}

func TestCreateWaitsForAnActiveCluster(t *testing.T) {
	resp, state := runCreate(t, func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/kubeconfig.yaml") {
			w.Header().Set("Content-Type", "text/plain")
			_, _ = w.Write([]byte("kubeconfig"))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(clusterPayload("ACTIVE", "1.36"))
	})

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %v", resp.Diagnostics)
	}
	if state.ID.ValueString() != "cluster" || state.KubeConfig.ValueString() != "kubeconfig" {
		t.Errorf("unexpected state: %+v", state)
	}
}

func TestWaitForClusterCreationIsBounded(t *testing.T) {
	cc := testClient(t, func(w http.ResponseWriter, _ *http.Request) {
		view := clusterPayload("DEPLOYING", "1.36")
		view.Waiting = &tmp.KubernetesWaiting{Attempts: 3, Dependency: "OTEL_API", Since: "2026-01-01T00:00:00Z", Step: "OTEL_CONFIG"}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(view)
	})

	var diags diag.Diagnostics
	if waitForClusterCreation(t.Context(), cc, "orga", "cluster", time.Millisecond, 50*time.Millisecond, &diags) {
		t.Fatal("expected the wait to give up")
	}

	// the status and the dependency are what tells a stuck deployment apart
	detail := diags.Errors()[0].Detail()
	for _, want := range []string{"did not become ACTIVE", "last seen DEPLOYING", "waiting for OTEL_API", "3 attempts", "step OTEL_CONFIG"} {
		if !strings.Contains(detail, want) {
			t.Errorf("error %q, want it to contain %q", detail, want)
		}
	}
}

func TestLastSeenDetail(t *testing.T) {
	got := lastSeenDetail(&tmp.ClusterView{Status: "RECONCILING"})
	if got != "the cluster was last seen RECONCILING" {
		t.Errorf("lastSeenDetail() = %q", got)
	}
}
