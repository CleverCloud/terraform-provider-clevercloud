package keycloak

import (
	"context"
	_ "embed"
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
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
	Realms        types.Set    `tfsdk:"realms"`
}

//go:embed doc.md
var resourceKeycloakDoc string

// GetRealms extracts realms from the Set as a string slice
func (k *Keycloak) GetRealms(ctx context.Context) []string {
	if k.Realms.IsNull() || k.Realms.IsUnknown() {
		return []string{}
	}

	var realms []string
	k.Realms.ElementsAs(ctx, &realms, false)
	return realms
}

// SetRealms converts a string slice to a Set and assigns it to Realms
func (k *Keycloak) SetRealms(ctx context.Context, realms []string, diags *diag.Diagnostics) {
	if len(realms) == 0 {
		k.Realms = types.SetNull(types.StringType)
		return
	}

	realmsSet, d := types.SetValueFrom(ctx, types.StringType, realms)
	diags.Append(d...)
	k.Realms = realmsSet
}

// GetRealmsCommaSeparated returns realms as comma-separated string for API
func (k *Keycloak) GetRealmsCommaSeparated(ctx context.Context) string {
	realms := k.GetRealms(ctx)
	if len(realms) == 0 {
		return ""
	}
	return strings.Join(realms, ",")
}

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
			"realms": schema.SetAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "List of realms to create. Note: realms can only be added, not removed once created.",
			},
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

	// Validate realms names (no special characters)
	if !plan.Realms.IsNull() && !plan.Realms.IsUnknown() {
		realms := plan.GetRealms(ctx)
		realmNamePattern := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

		for _, realm := range realms {
			if !realmNamePattern.MatchString(realm) {
				res.Diagnostics.AddError(
					"invalid realm name",
					fmt.Sprintf("realm '%s' contains invalid characters. Only alphanumeric, underscore and hyphen are allowed", realm),
				)
			}
		}
	}
}
