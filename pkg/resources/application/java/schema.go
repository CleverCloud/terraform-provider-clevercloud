package java

import (
	"context"
	_ "embed"
	"strconv"

	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/application"
	"go.clever-cloud.com/terraform-provider/pkg/attributes"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
)

type Java struct {
	application.Runtime
	DisableMaxMetaspace types.Bool   `tfsdk:"disable_max_metaspace"`
	ExtraJavaArgs       types.String `tfsdk:"extra_java_args"`
	GradleDeployGoal    types.String `tfsdk:"gradle_deploy_goal"`
	JarArgs             types.String `tfsdk:"jar_args"`
	JarPath             types.String `tfsdk:"jar_path"`
	JavaVersion         types.String `tfsdk:"java_version"`
	MavenDeployGoal     types.String `tfsdk:"maven_deploy_goal"`
	MavenProfiles       types.String `tfsdk:"maven_profiles"`
	NudgeAppId          types.String `tfsdk:"nudge_app_id"`
	Play1Version        types.String `tfsdk:"play1_version"`
	RunCommand          types.String `tfsdk:"run_command"`
	SbtDeployGoal       types.String `tfsdk:"sbt_deploy_goal"`
	SbtTargetBin        types.String `tfsdk:"sbt_target_bin"`
	SbtTargetDir        types.String `tfsdk:"sbt_target_dir"`
}

type JavaV0 struct {
	application.RuntimeV0
	JavaVersion types.String `tfsdk:"java_version"`
}

//go:embed doc.md
var javaDoc string

func (r ResourceJava) Schema(ctx context.Context, req resource.SchemaRequest, res *resource.SchemaResponse) {
	res.Schema = schemaJava
}

var schemaJava = schema.Schema{
	Version:             1,
	MarkdownDescription: javaDoc,
	Attributes: application.WithRuntimeCommons(map[string]schema.Attribute{
		// CC_DISABLE_MAX_METASPACE
		"disable_max_metaspace": schema.BoolAttribute{
			Optional:            true,
			Computed:            true,
			Default:             booldefault.StaticBool(false),
			MarkdownDescription: "Allows to disable the Java option `-XX:MaxMetaspaceSize`",
		},
		// CC_EXTRA_JAVA_ARGS
		"extra_java_args": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Define extra arguments to pass to `java` for JAR",
		},
		// GRADLE_DEPLOY_GOAL
		"gradle_deploy_goal": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Define which Gradle goals to run during build",
		},
		// CC_JAR_ARGS
		"jar_args": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Define arguments to pass to the launched JAR",
		},
		// CC_JAR_PATH
		"jar_path": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Define the path to your JAR",
		},
		// CC_JAVA_VERSION
		"java_version": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Choose the JVM version between 7 to 24 for OpenJDK or `graalvm-ce` for GraalVM (default: 21)",
		},
		// MAVEN_DEPLOY_GOAL
		"maven_deploy_goal": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Define which Maven goals to run during build",
		},
		// CC_MAVEN_PROFILES
		"maven_profiles": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Define which Maven profile to use during default build",
		},
		// NUDGE_APPID
		"nudge_app_id": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Nudge application ID",
		},
		// PLAY1_VERSION
		"play1_version": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Define which play1 version to use between `1.2`, `1.3`, `1.4` and `1.5`",
		},
		// CC_RUN_COMMAND
		"run_command": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Custom command to run your application. Replaces the default behavior",
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

var schemaJavaV0 = schema.Schema{
	Version:             0,
	MarkdownDescription: javaDoc,
	Attributes: application.WithRuntimeCommonsV0(map[string]schema.Attribute{
		"java_version": schema.StringAttribute{
			Optional:    true,
			Description: "Choose the JVM version between 7 to 24 for OpenJDK or graalvm-ce for GraalVM 21.0.0.2 (based on OpenJDK 11.0).",
		},
	}),
	Blocks: attributes.WithBlockRuntimeCommons(map[string]schema.Block{}),
}

func (plan *Java) toEnv(ctx context.Context, diags *diag.Diagnostics) map[string]string {
	env := plan.ToEnv(ctx, diags)
	if diags.HasError() {
		return env
	}

	pkg.IfIsSetB(plan.DisableMaxMetaspace, func(b bool) { env[CC_DISABLE_MAX_METASPACE] = strconv.FormatBool(b) })
	pkg.IfIsSetStr(plan.ExtraJavaArgs, func(s string) { env[CC_EXTRA_JAVA_ARGS] = s })
	pkg.IfIsSetStr(plan.GradleDeployGoal, func(s string) { env[GRADLE_DEPLOY_GOAL] = s })
	pkg.IfIsSetStr(plan.JarArgs, func(s string) { env[CC_JAR_ARGS] = s })
	pkg.IfIsSetStr(plan.JarPath, func(s string) { env[CC_JAR_PATH] = s })
	pkg.IfIsSetStr(plan.JavaVersion, func(s string) { env[CC_JAVA_VERSION] = s })
	pkg.IfIsSetStr(plan.MavenDeployGoal, func(s string) { env[MAVEN_DEPLOY_GOAL] = s })
	pkg.IfIsSetStr(plan.MavenProfiles, func(s string) { env[CC_MAVEN_PROFILES] = s })
	pkg.IfIsSetStr(plan.NudgeAppId, func(s string) { env[NUDGE_APPID] = s })
	pkg.IfIsSetStr(plan.Play1Version, func(s string) { env[PLAY1_VERSION] = s })
	pkg.IfIsSetStr(plan.RunCommand, func(s string) { env[CC_RUN_COMMAND] = s })
	pkg.IfIsSetStr(plan.SbtDeployGoal, func(s string) { env[SBT_DEPLOY_GOAL] = s })
	pkg.IfIsSetStr(plan.SbtTargetBin, func(s string) { env[CC_SBT_TARGET_BIN] = s })
	pkg.IfIsSetStr(plan.SbtTargetDir, func(s string) { env[CC_SBT_TARGET_DIR] = s })

	return env
}

func (java *Java) fromEnv(ctx context.Context, env map[string]string) {
	m := helper.NewEnvMap(env)

	if disable, err := strconv.ParseBool(m.Pop(CC_DISABLE_MAX_METASPACE)); err == nil {
		java.DisableMaxMetaspace = pkg.FromBool(disable)
	}
	java.ExtraJavaArgs = pkg.FromStr(m.Pop(CC_EXTRA_JAVA_ARGS))
	java.GradleDeployGoal = pkg.FromStr(m.Pop(GRADLE_DEPLOY_GOAL))
	java.JarArgs = pkg.FromStr(m.Pop(CC_JAR_ARGS))
	java.JarPath = pkg.FromStr(m.Pop(CC_JAR_PATH))
	java.JavaVersion = pkg.FromStr(m.Pop(CC_JAVA_VERSION))
	java.MavenDeployGoal = pkg.FromStr(m.Pop(MAVEN_DEPLOY_GOAL))
	java.MavenProfiles = pkg.FromStr(m.Pop(CC_MAVEN_PROFILES))
	java.NudgeAppId = pkg.FromStr(m.Pop(NUDGE_APPID))
	java.Play1Version = pkg.FromStr(m.Pop(PLAY1_VERSION))
	java.RunCommand = pkg.FromStr(m.Pop(CC_RUN_COMMAND))
	java.SbtDeployGoal = pkg.FromStr(m.Pop(SBT_DEPLOY_GOAL))
	java.SbtTargetBin = pkg.FromStr(m.Pop(CC_SBT_TARGET_BIN))
	java.SbtTargetDir = pkg.FromStr(m.Pop(CC_SBT_TARGET_DIR))

	java.FromEnvironment(ctx, m)
}

func (java *Java) toDeployment(gitAuth *http.BasicAuth) *application.Deployment {
	if java.Deployment == nil || java.Deployment.Repository.IsNull() {
		return nil
	}

	return &application.Deployment{
		Repository:    java.Deployment.Repository.ValueString(),
		Commit:        java.Deployment.Commit.ValueStringPointer(),
		CleverGitAuth: gitAuth,
	}
}
