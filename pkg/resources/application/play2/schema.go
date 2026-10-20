package play2

import (
	"context"
	_ "embed"

	"go.clever-cloud.com/terraform-provider/pkg/attributes"
	"go.clever-cloud.com/terraform-provider/pkg/resources/application"

	"maps"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	helperMaps "github.com/miton18/helper/maps"
	"go.clever-cloud.com/terraform-provider/pkg"
)

type Play2 struct {
	application.Runtime
	JavaVersion         types.String `tfsdk:"java_version"`
	SbtBuildGoal        types.String `tfsdk:"sbt_build_goal"`
	SbtDeployArgs       types.String `tfsdk:"sbt_deploy_args"`
	SbtTargetBin        types.String `tfsdk:"sbt_target_bin"`
	SbtTargetDir        types.String `tfsdk:"sbt_target_dir"`
	DisableMaxMetaspace types.Bool   `tfsdk:"disable_max_metaspace"`
	APMAgentEnabled     types.Bool   `tfsdk:"apm_agent_enabled"`
	APMAgentContinuous  types.Bool   `tfsdk:"apm_agent_continuous"`
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
	Attributes: application.WithRuntimeCommons(map[string]schema.Attribute{
		"java_version": schema.StringAttribute{
			Optional:    true,
			Description: "Choose the JVM version between 7 to 24 for OpenJDK or graalvm-ce for GraalVM 21.0.0.2 (based on OpenJDK 11.0) (`CC_JAVA_VERSION`).",
		},
		"sbt_build_goal": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "SBT task(s) building the application, `stage` by default (`CC_SBT_BUILD_GOAL`)",
		},
		"sbt_deploy_args": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Arguments passed to the executable produced by the SBT build (`CC_SBT_DEPLOY_ARGS`)",
		},
		"sbt_target_bin": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Name of the executable to run from `target/universal/stage/bin`, when the build produces several (`CC_SBT_TARGET_BIN`)",
		},
		"sbt_target_dir": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Project subdirectory whose `target/universal/stage/bin` holds the executable to run, for multi-project builds (`CC_SBT_TARGET_DIR`)",
		},
		"disable_max_metaspace": schema.BoolAttribute{
			Optional:            true,
			MarkdownDescription: "Do not cap the JVM metaspace, which the platform limits to 15% of the memory otherwise (`CC_DISABLE_MAX_METASPACE`)",
		},
		"apm_agent_enabled": schema.BoolAttribute{
			Optional:            true,
			MarkdownDescription: "Attach the Elastic APM Java agent to the application (`CC_JAVA_APM_AGENT_ENABLE`)",
		},
		"apm_agent_continuous": schema.BoolAttribute{
			Optional:            true,
			MarkdownDescription: "Keep attaching the Elastic APM Java agent to the JVMs started later, instead of once at startup (`CC_JAVA_APM_AGENT_CONTINUOUS`)",
		},
	}),
	Blocks: attributes.WithBlockRuntimeCommons(map[string]schema.Block{}),
}

var schemaPlay2V0 = schema.Schema{
	Version:             0,
	MarkdownDescription: play2Doc,
	Attributes:          application.WithRuntimeCommonsV0(map[string]schema.Attribute{}),
	Blocks:              attributes.WithBlockRuntimeCommons(map[string]schema.Block{}),
}

func (plan *Play2) ToEnv(ctx context.Context, diags *diag.Diagnostics) map[string]string {
	env := map[string]string{}

	// do not use the real map since ElementAs can nullish it
	// https://github.com/hashicorp/terraform-plugin-framework/issues/698
	customEnv := map[string]string{}
	diags.Append(plan.Environment.ElementsAs(ctx, &customEnv, false)...)
	if diags.HasError() {
		return env
	}
	maps.Copy(env, customEnv)

	pkg.WriteEnv(env, plan, envVars...)
	env = pkg.Merge(env, plan.Hooks.ToEnv())
	env = pkg.Merge(env, plan.Integrations.ToEnv(ctx, diags))
	return env
}

func (play2 *Play2) FromEnv(ctx context.Context, env *helperMaps.Map[string, string], diags *diag.Diagnostics) {
	if play2 == nil || env == nil {
		return
	}

	pkg.ReadEnv(env, play2, envVars...)
	play2.Integrations = attributes.FromEnvIntegrations(ctx, env, play2.Integrations, diags)
}
