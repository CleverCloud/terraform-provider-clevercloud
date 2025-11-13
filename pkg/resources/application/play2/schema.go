package play2

import (
	"context"
	_ "embed"
	"maps"
	"strings"

	"go.clever-cloud.com/terraform-provider/pkg/attributes"
	"go.clever-cloud.com/terraform-provider/pkg/resources/application"

	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"go.clever-cloud.com/terraform-provider/pkg"
)

type Play2 struct {
	application.Runtime
}

type Play2V0 struct {
	application.RuntimeV0
}

//go:embed doc.md
var play2Doc string

func (r ResourcePlay2) Schema(ctx context.Context, req resource.SchemaRequest, res *resource.SchemaResponse) {
	res.Schema = schemaPlay2
}

var schemaPlay2 = schema.Schema{
	Version:             1,
	MarkdownDescription: play2Doc,
	Attributes:          application.WithRuntimeCommons(map[string]schema.Attribute{}),
	Blocks:              attributes.WithBlockRuntimeCommons(map[string]schema.Block{}),
}

var schemaPlay2V0 = schema.Schema{
	Version:             0,
	MarkdownDescription: play2Doc,
	Attributes:          application.WithRuntimeCommonsV0(map[string]schema.Attribute{}),
	Blocks:              attributes.WithBlockRuntimeCommons(map[string]schema.Block{}),
}

func (plan *Play2) toEnv(ctx context.Context, diags *diag.Diagnostics) map[string]string {
	env := map[string]string{}

	// do not use the real map since ElementAs can nullish it
	// https://github.com/hashicorp/terraform-plugin-framework/issues/698
	customEnv := map[string]string{}
	diags.Append(plan.Environment.ElementsAs(ctx, &customEnv, false)...)
	if diags.HasError() {
		return env
	}
	maps.Copy(env, customEnv)

	pkg.IfIsSetStr(plan.AppFolder, func(s string) { env["APP_FOLDER"] = s })
	return env
}

func (play2 *Play2) toDeployment(gitAuth *http.BasicAuth) *application.Deployment {
	if play2.Deployment == nil || play2.Deployment.Repository.IsNull() {
		return nil
	}

	d := &application.Deployment{
		Repository:    play2.Deployment.Repository.ValueString(),
		Commit:        play2.Deployment.Commit.ValueStringPointer(),
		CleverGitAuth: gitAuth,
	}

	if !play2.Deployment.BasicAuthentication.IsNull() && !play2.Deployment.BasicAuthentication.IsUnknown() {
		// Expect validation to be done in the schema valisation step
		userPass := play2.Deployment.BasicAuthentication.ValueString()
		splits := strings.SplitN(userPass, ":", 2)
		d.Username = &splits[0]
		d.Password = &splits[1]
	}

	return d
}
