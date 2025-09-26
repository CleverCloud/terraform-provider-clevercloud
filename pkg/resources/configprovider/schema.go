package configprovider

import (
	"context"
	_ "embed"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/attributes"
)

type ConfigProvider struct {
	attributes.Addon
	Environment types.Map `tfsdk:"environment"`
}

//go:embed doc.md
var resourceConfigProviderDoc string

func (r ResourceConfigProvider) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:             0,
		MarkdownDescription: resourceConfigProviderDoc,
		Attributes: attributes.WithAddonCommons(map[string]schema.Attribute{

			"environment": schema.MapAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Environment variables injected into the application",
				ElementType: types.StringType,
			},
		}),
	}
}

// https://developer.hashicorp.com/terraform/plugin/framework/resources/state-upgrade#implementing-state-upgrade-support
func (r ResourceConfigProvider) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{}
}

func (appCp ConfigProvider) toEnv(ctx context.Context, diags *diag.Diagnostics) map[string]string {
	env := map[string]string{}

	// do not use the real map since ElementAs can nullish it
	// https://github.com/hashicorp/terraform-plugin-framework/issues/698
	customEnv := map[string]string{}
	diags.Append(appCp.Environment.ElementsAs(ctx, &customEnv, false)...)
	if diags.HasError() {
		return env
	}
	env = pkg.Merge(env, customEnv)

	return env
}

func (appCp *ConfigProvider) fromEnv(ctx context.Context, env map[string]string, diags *diag.Diagnostics) {
	m, d := types.MapValueFrom(ctx, types.StringType, env)
	diags.Append(d...)
	if diags.HasError() {
		return
	}
	appCp.Environment = m
}
