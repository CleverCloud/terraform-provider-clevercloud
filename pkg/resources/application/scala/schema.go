package scala

import (
	"context"
	_ "embed"

	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/resources/application"
	"go.clever-cloud.com/terraform-provider/pkg/attributes"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
)

type Scala struct {
	application.Runtime
	SbtDeployGoal types.String `tfsdk:"sbt_deploy_goal"`
	SbtTargetBin  types.String `tfsdk:"sbt_target_bin"`
	SbtTargetDir  types.String `tfsdk:"sbt_target_dir"`
}

type ScalaV0 struct {
	application.RuntimeV0
}

//go:embed doc.md
var scalaDoc string

func (r ResourceScala) Schema(ctx context.Context, req resource.SchemaRequest, res *resource.SchemaResponse) {
	res.Schema = schemaScala
}

var schemaScala = schema.Schema{
	Version:             1,
	MarkdownDescription: scalaDoc,
	Attributes: application.WithRuntimeCommons(map[string]schema.Attribute{
		// SBT_DEPLOY_GOAL
		"sbt_deploy_goal": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Define which SBT goals to run during build (default: `stage`)",
		},
		// CC_SBT_TARGET_BIN
		"sbt_target_bin": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Define the bin to pick in the `CC_SBT_TARGET_DIR`",
		},
		// CC_SBT_TARGET_DIR
		"sbt_target_dir": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Define the folder the `target` dir is in (default: `.`)",
		},
	}),
	Blocks: attributes.WithBlockRuntimeCommons(map[string]schema.Block{}),
}

var schemaScalaV0 = schema.Schema{
	Version:             0,
	MarkdownDescription: scalaDoc,
	Attributes:          application.WithRuntimeCommonsV0(map[string]schema.Attribute{}),
	Blocks:              attributes.WithBlockRuntimeCommons(map[string]schema.Block{}),
}

func (plan *Scala) toEnv(ctx context.Context, diags *diag.Diagnostics) map[string]string {
	env := plan.ToEnv(ctx, diags)
	if diags.HasError() {
		return env
	}

	pkg.IfIsSetStr(plan.SbtDeployGoal, func(s string) { env[SBT_DEPLOY_GOAL] = s })
	pkg.IfIsSetStr(plan.SbtTargetBin, func(s string) { env[CC_SBT_TARGET_BIN] = s })
	pkg.IfIsSetStr(plan.SbtTargetDir, func(s string) { env[CC_SBT_TARGET_DIR] = s })

	return env
}

func (scala *Scala) fromEnv(ctx context.Context, env map[string]string) {
	m := helper.NewEnvMap(env)

	scala.SbtDeployGoal = pkg.FromStr(m.Pop(SBT_DEPLOY_GOAL))
	scala.SbtTargetBin = pkg.FromStr(m.Pop(CC_SBT_TARGET_BIN))
	scala.SbtTargetDir = pkg.FromStr(m.Pop(CC_SBT_TARGET_DIR))

	scala.FromEnvironment(ctx, m)
}

func (java *Scala) toDeployment(gitAuth *http.BasicAuth) *application.Deployment {
	if java.Deployment == nil || java.Deployment.Repository.IsNull() {
		return nil
	}

	return &application.Deployment{
		Repository:    java.Deployment.Repository.ValueString(),
		Commit:        java.Deployment.Commit.ValueStringPointer(),
		CleverGitAuth: gitAuth,
	}
}
