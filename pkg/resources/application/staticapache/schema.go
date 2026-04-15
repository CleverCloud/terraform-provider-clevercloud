package staticapache

import (
	"context"
	_ "embed"
	"strings"

	"go.clever-cloud.com/terraform-provider/pkg/attributes"
	"go.clever-cloud.com/terraform-provider/pkg/resources/application"

	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/miton18/helper/maps"
	"go.clever-cloud.com/terraform-provider/pkg"
)

type StaticApache struct {
	application.Runtime
	// StaticApache related
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
	Attributes:          application.WithRuntimeCommons(map[string]schema.Attribute{}),
	Blocks:              attributes.WithBlockRuntimeCommons(map[string]schema.Block{}),
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

	pkg.IfIsSetStr(plan.AppFolder, func(s string) { env["APP_FOLDER"] = s })
	env = pkg.Merge(env, plan.Hooks.ToEnv())
	env = pkg.Merge(env, plan.Integrations.ToEnv(ctx, diags))
	return env
}

func (s *StaticApache) FromEnv(ctx context.Context, env *maps.Map[string, string], diags *diag.Diagnostics) {
	s.AppFolder = pkg.FromStrPtr(env.PopPtr("APP_FOLDER"))

	s.Integrations = attributes.FromEnvIntegrations(ctx, env, s.Integrations, diags)
}

func (s *StaticApache) ToDeployment(gitAuth *http.BasicAuth) *application.Deployment {
	if s.Deployment == nil || s.Deployment.Repository.IsNull() {
		return nil
	}

	d := &application.Deployment{
		Repository:    s.Deployment.Repository.ValueString(),
		Commit:        s.Deployment.Commit.ValueStringPointer(),
		CleverGitAuth: gitAuth,
	}

	if !s.Deployment.BasicAuthentication.IsNull() && !s.Deployment.BasicAuthentication.IsUnknown() {
		// Expect validation to be done in the schema valisation step
		userPass := s.Deployment.BasicAuthentication.ValueString()
		splits := strings.SplitN(userPass, ":", 2)
		d.Username = &splits[0]
		d.Password = &splits[1]
	}

	return d
}
