package keycloak

import (
	"context"
	_ "embed"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
	"go.clever-cloud.dev/sdk/models"
)

type Keycloak struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	Region        types.String `tfsdk:"region"`
	Version       types.String `tfsdk:"version"`
	AccessDomain  types.String `tfsdk:"access_domain"`
	Host          types.String `tfsdk:"host"`
	AdminUsername types.String `tfsdk:"admin_username"`
	AdminPassword types.String `tfsdk:"admin_password"`
	FSBucketID    types.String `tfsdk:"fsbucket_id"`
}

//go:embed doc.md
var resourceKeycloakDoc string

func (r ResourceKeycloak) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:             0,
		MarkdownDescription: resourceKeycloakDoc,
		Attributes: map[string]schema.Attribute{
			"id":   schema.StringAttribute{Computed: true, MarkdownDescription: "Generated unique identifier", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"name": schema.StringAttribute{Required: true, MarkdownDescription: "Name of the service"},
			"region": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("par"),
				MarkdownDescription: "Geographical region where the data will be stored",
			},
			"access_domain":  schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "Main domaine to access the instance"},
			"version":        schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "Keycloak official version"},
			"host":           schema.StringAttribute{Computed: true, MarkdownDescription: "URL to access Keycloak"},
			"admin_username": schema.StringAttribute{Computed: true, MarkdownDescription: "Initial admin username for Keycloak"},
			"admin_password": schema.StringAttribute{Computed: true, Sensitive: true, MarkdownDescription: "Initial admin password for Keycloak"},
			"fsbucket_id":    schema.StringAttribute{Computed: true, MarkdownDescription: "ID of the fsbucket subresource"},
		},
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

	// Only validate a version the user is actually asking for. An add-on that has
	// been running for a while may sit on a version the provider no longer
	// offers; rejecting it would make it impossible to import or even to plan
	// against — precisely the add-ons that most need to come under Terraform.
	if !req.State.Raw.IsNull() {
		state := helper.From[Keycloak](ctx, req.State, &res.Diagnostics)
		if res.Diagnostics.HasError() {
			return
		}

		if state.Version.Equal(plan.Version) {
			return
		}
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

// The API-to-state mappers for clevercloud_keycloak follow.
// See CONTRIBUTING.md § "API → state mapping".

// FromAddon maps the generic add-on view, the only source of name and region
// for this resource.
func (kc *Keycloak) FromAddon(ctx context.Context, addon *tmp.AddonResponse, diags *diag.Diagnostics) *Keycloak {
	if kc == nil || addon == nil {
		return kc
	}

	kc.Name = pkg.FromStr(addon.Name)
	kc.Region = pkg.FromStr(addon.Region)

	return kc
}

// FromKeycloak maps the product view: the access URL, the initial credentials,
// the version and the bucket backing it.
//
// access_domain lives in the payload's own env map rather than in a field of its
// own, which is why this resource has no FromEnv: there is no add-on env call to
// make, the variable arrives inside the product view.
func (kc *Keycloak) FromKeycloak(ctx context.Context, api *models.Keycloak, diags *diag.Diagnostics) *Keycloak {
	if kc == nil || api == nil {
		return kc
	}

	kc.Host = pkg.FromStr(api.AccessURL)
	kc.AdminUsername = pkg.FromStr(api.InitialCredentials.User)
	kc.AdminPassword = pkg.FromStr(api.InitialCredentials.Password)
	kc.Version = pkg.FromStr(api.Version)
	kc.AccessDomain = pkg.FromStr(api.EnvVars["CC_KEYCLOAK_HOSTNAME"])
	kc.FSBucketID = types.StringPointerValue(api.Resources.FsbucketID)

	return kc
}
