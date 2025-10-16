package play2

import (
	"context"
	_ "embed"

	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/application"
	"go.clever-cloud.com/terraform-provider/pkg/attributes"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
)

type Play2 struct {
	attributes.Runtime
	Play1Version  types.String `tfsdk:"play1_version"`
	SbtDeployGoal types.String `tfsdk:"sbt_deploy_goal"`
	SbtTargetBin  types.String `tfsdk:"sbt_target_bin"`
	SbtTargetDir  types.String `tfsdk:"sbt_target_dir"`
}

type Play2V0 struct {
	attributes.RuntimeV0
}

//go:embed doc.md
var play2Doc string

func (r ResourcePlay2) Schema(ctx context.Context, req resource.SchemaRequest, res *resource.SchemaResponse) {
	res.Schema = schemaPlay2
}

var schemaPlay2 = schema.Schema{
	Version:             1,
	MarkdownDescription: play2Doc,
	Attributes: attributes.WithRuntimeCommons(map[string]schema.Attribute{
		// PLAY1_VERSION
		"play1_version": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Define which play1 version to use between `1.2`, `1.3`, `1.4` and `1.5`",
		},
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

var schemaPlay2V0 = schema.Schema{
	Version:             0,
	MarkdownDescription: play2Doc,
	Attributes:          attributes.WithRuntimeCommonsV0(map[string]schema.Attribute{}),
	Blocks:              attributes.WithBlockRuntimeCommons(map[string]schema.Block{}),
}

func (plan *Play2) toEnv(ctx context.Context, diags *diag.Diagnostics) map[string]string {
	env := plan.ToEnv(ctx, diags)
	if diags.HasError() {
		return env
	}

	pkg.IfIsSetStr(plan.Play1Version, func(s string) { env[PLAY1_VERSION] = s })
	pkg.IfIsSetStr(plan.SbtDeployGoal, func(s string) { env[SBT_DEPLOY_GOAL] = s })
	pkg.IfIsSetStr(plan.SbtTargetBin, func(s string) { env[CC_SBT_TARGET_BIN] = s })
	pkg.IfIsSetStr(plan.SbtTargetDir, func(s string) { env[CC_SBT_TARGET_DIR] = s })

	return env
}

func (play2 *Play2) fromEnv(ctx context.Context, env map[string]string) {
	m := helper.NewEnvMap(env)

	play2.Play1Version = pkg.FromStr(m.Pop(PLAY1_VERSION))
	play2.SbtDeployGoal = pkg.FromStr(m.Pop(SBT_DEPLOY_GOAL))
	play2.SbtTargetBin = pkg.FromStr(m.Pop(CC_SBT_TARGET_BIN))
	play2.SbtTargetDir = pkg.FromStr(m.Pop(CC_SBT_TARGET_DIR))

	play2.FromEnvironment(ctx, m)
}

func (play2 *Play2) toDeployment(gitAuth *http.BasicAuth) *application.Deployment {
	if play2.Deployment == nil || play2.Deployment.Repository.IsNull() {
		return nil
	}

	return &application.Deployment{
		Repository:    play2.Deployment.Repository.ValueString(),
		Commit:        play2.Deployment.Commit.ValueStringPointer(),
		CleverGitAuth: gitAuth,
	}
}
