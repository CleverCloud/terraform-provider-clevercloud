package scala

import (
	"context"
	_ "embed"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/application"
	"go.clever-cloud.com/terraform-provider/pkg/attributes"
)

type Scala struct {
	attributes.Runtime
	// Scala related
}

//go:embed doc.md
var scalaDoc string

func (r ResourceScala) Schema(ctx context.Context, req resource.SchemaRequest, res *resource.SchemaResponse) {
	res.Schema = schema.Schema{
		Version:             0,
		MarkdownDescription: scalaDoc,
		Attributes:          attributes.WithRuntimeCommons(map[string]schema.Attribute{}),
		Blocks:              attributes.WithBlockRuntimeCommons(map[string]schema.Block{}),
	}
}

// https://developer.hashicorp.com/terraform/plugin/framework/resources/state-upgrade#implementing-state-upgrade-support
func (plan *Scala) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{}
}

func (plan *Scala) toEnv(ctx context.Context, diags diag.Diagnostics) map[string]string {
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

	pkg.IfIsSet(plan.AppFolder, func(s string) { env["APP_FOLDER"] = s })
	return env
}

func (java *Scala) toDeployment() *application.Deployment {
	if java.Deployment == nil || java.Deployment.Repository.IsNull() {
		return nil
	}

	return &application.Deployment{
		Repository: java.Deployment.Repository.ValueString(),
		Commit:     java.Deployment.Commit.ValueStringPointer(),
	}
}
