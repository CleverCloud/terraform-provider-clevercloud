package static

import (
	"context"
	_ "embed"

	"go.clever-cloud.com/terraform-provider/pkg/attributes"
	"go.clever-cloud.com/terraform-provider/pkg/resources/application"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/miton18/helper/maps"
	"go.clever-cloud.com/terraform-provider/pkg"
)

type Static struct {
	application.Runtime
	Webroot           types.String `tfsdk:"webroot"`
	BuildCommand      types.String `tfsdk:"build_command"`
	HugoVersion       types.String `tfsdk:"hugo_version"`
	AutobuildOutdir   types.String `tfsdk:"autobuild_outdir"`
	StaticServer      types.String `tfsdk:"static_server"`
	StaticServerFlags types.String `tfsdk:"static_server_flags"`
	Caddyfile         types.String `tfsdk:"caddyfile"`
}

type StaticV0 struct {
	application.RuntimeV0
	// Static related
}

//go:embed doc.md
var staticDoc string

func (r ResourceStatic) Schema(ctx context.Context, req resource.SchemaRequest, res *resource.SchemaResponse) {
	res.Schema = schemaStatic
}

var schemaStatic = schema.Schema{
	Version:             1,
	MarkdownDescription: staticDoc,
	Attributes: application.WithRuntimeCommons(map[string]schema.Attribute{
		"webroot": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Folder to serve, relative to the application root (`CC_WEBROOT`)",
		},
		"build_command": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Command building the site during the deployment, instead of the platform's autobuild (`CC_BUILD_COMMAND`)",
		},
		"hugo_version": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Hugo version used to build a Hugo site (`CC_HUGO_VERSION`)",
		},
		"autobuild_outdir": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Folder the autobuild generates the site into (`CC_STATIC_AUTOBUILD_OUTDIR`)",
		},
		"static_server": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Server serving the files, `static-web-server` or `caddy` (`CC_STATIC_SERVER`)",
			Validators: []validator.String{
				stringvalidator.OneOf("static-web-server", "caddy"),
			},
		},
		"static_server_flags": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Extra flags passed to the static file server (`CC_STATIC_FLAGS`)",
		},
		"caddyfile": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Path of the Caddyfile used by the `caddy` server, `./Caddyfile` by default (`CC_STATIC_CADDYFILE`)",
		},
	}),
	Blocks: attributes.WithBlockRuntimeCommons(map[string]schema.Block{}),
}

var schemaStaticV0 = schema.Schema{
	Version:             0,
	MarkdownDescription: staticDoc,
	Attributes:          application.WithRuntimeCommonsV0(map[string]schema.Attribute{}),
	Blocks:              attributes.WithBlockRuntimeCommons(map[string]schema.Block{}),
}

func (plan *Static) ToEnv(ctx context.Context, diags *diag.Diagnostics) map[string]string {
	env := map[string]string{}

	// do not use the real map since ElementAs can nullish it
	// https://github.com/hashicorp/terraform-plugin-framework/issues/698
	customEnv := map[string]string{}
	diags.Append(plan.Environment.ElementsAs(ctx, &customEnv, false)...)
	if diags.HasError() {
		return env
	}
	for k, v := range customEnv {
		env[k] = v
	}

	pkg.WriteEnv(env, plan, envVars...)
	env = pkg.Merge(env, plan.Hooks.ToEnv())
	env = pkg.Merge(env, plan.Integrations.ToEnv(ctx, diags))
	return env
}

func (s *Static) FromEnv(ctx context.Context, env *maps.Map[string, string], diags *diag.Diagnostics) {
	if s == nil || env == nil {
		return
	}

	pkg.ReadEnv(env, s, envVars...)

	s.Integrations = attributes.FromEnvIntegrations(ctx, env, s.Integrations, diags)
}
