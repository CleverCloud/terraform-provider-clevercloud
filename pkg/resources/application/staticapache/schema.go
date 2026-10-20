package staticapache

import (
	"context"
	_ "embed"

	"go.clever-cloud.com/terraform-provider/pkg/attributes"
	"go.clever-cloud.com/terraform-provider/pkg/resources/application"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/miton18/helper/maps"
	"go.clever-cloud.com/terraform-provider/pkg"
)

type StaticApache struct {
	application.Runtime
	Webroot types.String `tfsdk:"webroot"`
}

type StaticApacheV0 struct {
	application.RuntimeV0
	// StaticApache related
}

//go:embed doc.md
var staticApacheDoc string

func (r ResourceStaticApache) Schema(ctx context.Context, req resource.SchemaRequest, res *resource.SchemaResponse) {
	res.Schema = schemaStaticApache
}

var schemaStaticApache = schema.Schema{
	Version:             1,
	MarkdownDescription: staticApacheDoc,
	Attributes: application.WithRuntimeCommons(map[string]schema.Attribute{
		"webroot": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Apache DocumentRoot, relative to the application root (`CC_WEBROOT`)",
		},
	}),
	Blocks: attributes.WithBlockRuntimeCommons(map[string]schema.Block{}),
}

var schemaStaticApacheV0 = schema.Schema{
	Version:             0,
	MarkdownDescription: staticApacheDoc,
	Attributes:          application.WithRuntimeCommonsV0(map[string]schema.Attribute{}),
	Blocks:              attributes.WithBlockRuntimeCommons(map[string]schema.Block{}),
}

func (plan *StaticApache) ToEnv(ctx context.Context, diags *diag.Diagnostics) map[string]string {
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

func (s *StaticApache) FromEnv(ctx context.Context, env *maps.Map[string, string], diags *diag.Diagnostics) {
	if s == nil || env == nil {
		return
	}

	pkg.ReadEnv(env, s, envVars...)

	s.Integrations = attributes.FromEnvIntegrations(ctx, env, s.Integrations, diags)
}
