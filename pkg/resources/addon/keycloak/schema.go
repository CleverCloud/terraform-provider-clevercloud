package keycloak

import (
	"context"
	_ "embed"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/resources/addon"
	"go.clever-cloud.dev/client"
	"go.clever-cloud.dev/sdk"
)

type Keycloak struct {
	addon.CommonAttributes
	Version       types.String `tfsdk:"version"`
	AccessDomain  types.String `tfsdk:"access_domain"`
	Host          types.String `tfsdk:"host"`
	AdminUsername types.String `tfsdk:"admin_username"`
	AdminPassword types.String `tfsdk:"admin_password"`
	FSBucketID    types.String `tfsdk:"fsbucket_id"`
}

func (kc *Keycloak) GetCommonPtr() *addon.CommonAttributes {
	return &kc.CommonAttributes
}

func (kc *Keycloak) GetAddonOptions() map[string]string {
	opts := map[string]string{}

	if !kc.AccessDomain.IsNull() && !kc.AccessDomain.IsUnknown() {
		opts["access-domain"] = kc.AccessDomain.ValueString()
	}
	if !kc.Version.IsNull() && !kc.Version.IsUnknown() {
		opts["version"] = kc.Version.ValueString()
	}

	return opts
}

func (kc *Keycloak) SetFromResponse(ctx context.Context, cc *client.Client, org string, addonID string, diags *diag.Diagnostics) {
	s := sdk.NewSDK(sdk.WithClient(cc))
	keycloakRes := s.V4().Keycloaks().Organisations().Ownerid(org).Keycloaks().Addonkeycloakid(kc.ID.ValueString()).Getkeycloak(ctx)
	if keycloakRes.HasError() {
		diags.AddError("failed to get Keycloak", keycloakRes.Error().Error())
		return
	}

	keycloak := keycloakRes.Payload()
	kc.Name = pkg.FromStr(keycloak.Name)
	kc.Host = pkg.FromStr(keycloak.AccessURL)
	kc.AdminUsername = pkg.FromStr(keycloak.InitialCredentials.User)
	kc.AdminPassword = pkg.FromStr(keycloak.InitialCredentials.Password)
	kc.Version = pkg.FromStr(keycloak.Version)
	kc.AccessDomain = pkg.FromStr(keycloak.EnvVars["CC_KEYCLOAK_HOSTNAME"])
	kc.FSBucketID = types.StringPointerValue(keycloak.Resources.FsbucketID)
}

func (kc *Keycloak) SetDefaults() {
	// Keycloak has no optional fields requiring defaults
}

//go:embed doc.md
var resourceKeycloakDoc string

func (r ResourceKeycloak) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:             0,
		MarkdownDescription: resourceKeycloakDoc,
		Attributes: addon.WithAddonCommons(map[string]schema.Attribute{
			// Single-plan addon: plan is computed, not user-specified
			"plan":             schema.StringAttribute{Computed: true},
			"access_domain":    schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "Main domaine to access the instance"},
			"version":          schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "Keycloak official version"},
			"host":             schema.StringAttribute{Computed: true, MarkdownDescription: "URL to access Keycloak"},
			"admin_username":   schema.StringAttribute{Computed: true, MarkdownDescription: "Initial admin username for Keycloak"},
			"admin_password":   schema.StringAttribute{Computed: true, Sensitive: true, MarkdownDescription: "Initial admin password for Keycloak"},
			"fsbucket_id":      schema.StringAttribute{Computed: true, MarkdownDescription: "ID of the fsbucket subresource"},
		}),
	}
}

func (r ResourceKeycloak) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, res *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() { // plan is null when calling Delete() methode
		return
	}
	plan := helper.From[Keycloak](ctx, req.Plan, &res.Diagnostics)
	if res.Diagnostics.HasError() {
		return
	}

	// Skip validation if version is not specified
	if !plan.Version.IsNull() && !plan.Version.IsUnknown() {
		infosRes := r.SDK.V4().AddonProviders().Keycloak().Getkeycloakproviderinformation(ctx)
		if infosRes.HasError() {
			res.Diagnostics.AddError("failed to get provider infos", infosRes.Error().Error())
		} else {
			infos := infosRes.Payload()

			versions := make([]string, 0, len(infos.Dedicated))
			for k := range infos.Dedicated {
				versions = append(versions, k)
			}

			_, ok := infos.Dedicated[plan.Version.ValueString()]
			if !ok {
				res.Diagnostics.AddError(
					"unavailable version",
					fmt.Sprintf("available versions are: %s", strings.Join(versions, ", ")),
				)
			}
		}
	}
}
