package configprovider

import (
	"context"
	_ "embed"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
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
				Sensitive:   true,
				Description: "Environment variables injected into the application",
				ElementType: types.StringType,
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

func (appCp *ConfigProvider) fromEnv(ctx context.Context, env map[string]string, diags *diag.Diagnostics) {
	m, d := types.MapValueFrom(ctx, types.StringType, env)
	diags.Append(d...)
	if diags.HasError() {
		return
	}
	appCp.Environment = m
}
