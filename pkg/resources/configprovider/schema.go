package configprovider

import (
	"context"
	_ "embed"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

type ConfigProvider struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Environment types.Map    `tfsdk:"environment"`
}

//go:embed doc.md
var resourceConfigProviderDoc string

func (r ResourceConfigProvider) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:             0,
		MarkdownDescription: resourceConfigProviderDoc,
		Attributes: map[string]schema.Attribute{
			"environment": schema.MapAttribute{
				Required:    true,
				Sensitive:   !pkg.BypassSensitiveImports(),
				Description: "Environment variables injected into the application",
				ElementType: types.StringType,
				Validators:  []validator.Map{pkg.NoNullMapValuesValidator()},
			},
			"id":   schema.StringAttribute{Computed: true, MarkdownDescription: "Generated unique identifier", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"name": schema.StringAttribute{Required: true, MarkdownDescription: "Name of the service"},
		},
	}
}

func (appCp ConfigProvider) toEnv(ctx context.Context, diags *diag.Diagnostics) map[string]string {
	env := map[string]string{}
	diags.Append(appCp.Environment.ElementsAs(ctx, &env, false)...)
	return env
}

// The API-to-state mappers for clevercloud_configprovider follow.
// See CONTRIBUTING.md § "API → state mapping".

// FromAddon maps the generic add-on view.
//
// The config-provider env endpoint carries no name, so this is the only source
// of it. Without it an import left the required attribute null and the plan
// failed (#447).
func (cp *ConfigProvider) FromAddon(ctx context.Context, addon *tmp.AddonResponse, diags *diag.Diagnostics) {
	if cp == nil || addon == nil {
		return
	}

	cp.Name = pkg.FromStr(addon.Name)
}

// FromEnv maps the configured variables, which are this resource's whole point.
// Non-destructive — see CONTRIBUTING.md.
func (cp *ConfigProvider) FromEnv(ctx context.Context, env tmp.EnvVars, diags *diag.Diagnostics) {
	if cp == nil || env == nil {
		return
	}

	environment, d := types.MapValueFrom(ctx, types.StringType, env.Map())
	diags.Append(d...)
	if d.HasError() {
		return
	}
	cp.Environment = environment
}
