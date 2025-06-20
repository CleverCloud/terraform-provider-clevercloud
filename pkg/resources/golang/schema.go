package golang

import (
	"context"
	_ "embed"

	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/application"
	"go.clever-cloud.com/terraform-provider/pkg/attributes"
)

type Go struct {
	attributes.Runtime
}

//go:embed doc.md
var goDoc string

func (r ResourceGo) Schema(ctx context.Context, req resource.SchemaRequest, res *resource.SchemaResponse) {

	res.Schema = schema.Schema{
		Version:             0,
		MarkdownDescription: goDoc,
		Attributes:          attributes.WithRuntimeCommons(map[string]schema.Attribute{}),
		Blocks:              attributes.WithBlockRuntimeCommons(map[string]schema.Block{}),
	}
}

// https://developer.hashicorp.com/terraform/plugin/framework/resources/state-upgrade#implementing-state-upgrade-support
func (r ResourceGo) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{}
}

func (g Go) toEnv(ctx context.Context, diags diag.Diagnostics) map[string]string {
	env := map[string]string{}

	// do not use the real map since ElementAs can nullish it
	// https://github.com/hashicorp/terraform-plugin-framework/issues/698
	customEnv := map[string]string{}
	diags.Append(g.Environment.ElementsAs(ctx, &customEnv, false)...)
	if diags.HasError() {
		return env
	}
	env = pkg.Merge(env, customEnv)

	pkg.IfIsSet(g.AppFolder, func(s string) { env["APP_FOLDER"] = s })
	env = pkg.Merge(env, g.Hooks.ToEnv())

	return env
}

func (g Go) toDeployment(gitAuth *http.BasicAuth) *application.Deployment {
	if g.Deployment == nil || g.Deployment.Repository.IsNull() {
		return nil
	}

	return &application.Deployment{
		Repository:    g.Deployment.Repository.ValueString(),
		Commit:        g.Deployment.Commit.ValueStringPointer(),
		CleverGitAuth: gitAuth,
	}
}
