package route

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/resources"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

type ResourceOtoroshiRoute struct {
	helper.Configurer
}

func NewResourceOtoroshiRoute() resource.Resource {
	return &ResourceOtoroshiRoute{}
}

func (r *ResourceOtoroshiRoute) Metadata(ctx context.Context, req resource.MetadataRequest, res *resource.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_otoroshi_route"
}

// ImportState overrides the shared passthrough on "id": a route is only reachable
// through the admin API of the Otoroshi hosting it, and a route ID alone does not
// say which one that is. The import ID therefore carries both, as
// <otoroshi_id>/<route_id>
func (r *ResourceOtoroshiRoute) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	otoroshiID, routeID, found := strings.Cut(req.ID, "/")
	if !found || otoroshiID == "" || routeID == "" {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			fmt.Sprintf("expected \"<otoroshi_id>/<route_id>\", got %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("otoroshi_id"), otoroshiID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), routeID)...)
}

// otoroshiContext gathers everything a route needs from the add-on: where to call
// the admin API, which networkgroup backends live in, and the default public domain
type otoroshiContext struct {
	client        *otoroshiClient
	networkgroup  string
	defaultDomain string
}

func (r *ResourceOtoroshiRoute) otoroshiContext(ctx context.Context, otoroshiID string, diags *diag.Diagnostics) *otoroshiContext {
	res := tmp.GetOtoroshi(ctx, r.Client(), otoroshiID)
	if res.HasError() {
		diags.AddError("failed to get Otoroshi add-on", res.Error().Error())
		return nil
	}
	otoroshi := res.Payload()

	if otoroshi.API == nil {
		diags.AddError(
			"Otoroshi API is not available",
			fmt.Sprintf("add-on %s exposes no API credentials, it may still be provisioning", otoroshiID),
		)
		return nil
	}

	if otoroshi.Features.NetworkGroup == nil {
		diags.AddError(
			"Otoroshi has no networkgroup",
			"a route forwards to applications through the add-on networkgroup, create the Otoroshi with `networkgroup = true`",
		)
		return nil
	}

	return &otoroshiContext{
		client:        newOtoroshiClient(otoroshi.API.URL, otoroshi.API.User, otoroshi.API.Secret),
		networkgroup:  otoroshi.Features.NetworkGroup.NetworkGroupID,
		defaultDomain: cleverappsDomain(otoroshi.Resources.Entrypoint),
	}
}

// cleverappsDomain derives the default public domain from an application ID.
// Clever Cloud exposes app_<uuid> under app-<uuid>.cleverapps.io
func cleverappsDomain(applicationID string) string {
	return strings.Replace(applicationID, "app_", "app-", 1) + ".cleverapps.io"
}

// frontendDomain assembles the "host/path" form Otoroshi expects
func frontendDomain(domain, path string) string {
	if path == "" {
		return domain
	}

	return domain + "/" + strings.TrimPrefix(path, "/")
}

// splitFrontendDomain is the reverse operation, used when reading state back
func splitFrontendDomain(frontend string) (domain, path string) {
	domain, path, found := strings.Cut(frontend, "/")
	if !found || path == "" {
		return domain, ""
	}

	return domain, "/" + path
}

// backendHostname is the networkgroup FQDN of the target application, built the
// same way the provider names networkgroup members
func backendHostname(applicationID, networkgroupID string) string {
	return resources.CanonicalFQDN(applicationID, networkgroupID)
}

// applicationIDFromHostname recovers the target application from a networkgroup
// FQDN. Needed on import, where the configuration is not available and the
// backend hostname is the only trace of which application the route serves
func applicationIDFromHostname(hostname string) string {
	applicationID, _, found := strings.Cut(hostname, ".m.")
	if !found {
		return ""
	}

	return applicationID
}
