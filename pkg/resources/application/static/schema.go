package static

import (
	"context"
	_ "embed"

	"go.clever-cloud.com/terraform-provider/pkg/attributes"
	"go.clever-cloud.com/terraform-provider/pkg/resources/application"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/miton18/helper/maps"
	"go.clever-cloud.com/terraform-provider/pkg"
)

type Static struct {
	application.Runtime
	// Static related
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
	Attributes:          application.WithRuntimeCommons(map[string]schema.Attribute{}),
	Blocks:              attributes.WithBlockRuntimeCommons(map[string]schema.Block{}),
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

	pkg.IfIsSetStr(plan.AppFolder, func(s string) { env["APP_FOLDER"] = s })
	env = pkg.Merge(env, plan.Hooks.ToEnv())
	env = pkg.Merge(env, plan.Integrations.ToEnv(ctx, diags))
	return env
}

func (s *Static) FromEnv(ctx context.Context, env *maps.Map[string, string], diags *diag.Diagnostics) {
	s.AppFolder = pkg.FromStrPtr(env.PopPtr("APP_FOLDER"))

	s.Integrations = attributes.FromEnvIntegrations(ctx, env, s.Integrations, diags)
}
