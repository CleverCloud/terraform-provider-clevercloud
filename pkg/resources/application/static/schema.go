package static

import (
	"go.clever-cloud.com/terraform-provider/pkg/attributes"
	"go.clever-cloud.com/terraform-provider/pkg/resources/application/common"
	"context"
	_ "embed"

	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"go.clever-cloud.com/terraform-provider/pkg"
)

type Static struct {
	common.Runtime
	// Static related
}

type StaticV0 struct {
	common.RuntimeV0
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
	Attributes:          common.WithRuntimeCommons(map[string]schema.Attribute{}),
	Blocks:              attributes.WithBlockRuntimeCommons(map[string]schema.Block{}),
}

var schemaStaticV0 = schema.Schema{
	Version:             0,
	MarkdownDescription: staticDoc,
	Attributes:          common.WithRuntimeCommonsV0(map[string]schema.Attribute{}),
	Blocks:              attributes.WithBlockRuntimeCommons(map[string]schema.Block{}),
}

func (plan *Static) toEnv(ctx context.Context, diags *diag.Diagnostics) map[string]string {
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
	return env
}

func (java *Static) toDeployment(gitAuth *http.BasicAuth) *common.Deployment {
	if java.Deployment == nil || java.Deployment.Repository.IsNull() {
		return nil
	}

	return &common.Deployment{
		Repository:    java.Deployment.Repository.ValueString(),
		Commit:        java.Deployment.Commit.ValueStringPointer(),
		CleverGitAuth: gitAuth,
	}
}
