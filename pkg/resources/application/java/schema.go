package java

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

type Java struct {
	application.Runtime
	JavaVersion         types.String `tfsdk:"java_version"`
	ExtraJavaArgs       types.String `tfsdk:"extra_java_args"`
	JarPath             types.String `tfsdk:"jar_path"`
	JarArgs             types.String `tfsdk:"jar_args"`
	MavenBuildGoal      types.String `tfsdk:"maven_build_goal"`
	MavenDeployGoal     types.String `tfsdk:"maven_deploy_goal"`
	MavenProfiles       types.String `tfsdk:"maven_profiles"`
	GradleBuildGoal     types.String `tfsdk:"gradle_build_goal"`
	GradleDeployGoal    types.String `tfsdk:"gradle_deploy_goal"`
	Play1Version        types.String `tfsdk:"play1_version"`
	DisableMaxMetaspace types.Bool   `tfsdk:"disable_max_metaspace"`
	APMAgentEnabled     types.Bool   `tfsdk:"apm_agent_enabled"`
	APMAgentContinuous  types.Bool   `tfsdk:"apm_agent_continuous"`
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
		"java_version": schema.StringAttribute{
			Optional:    true,
			Description: "Choose the JVM version between 7 to 24 for OpenJDK or graalvm-ce for GraalVM 21.0.0.2 (based on OpenJDK 11.0) (`CC_JAVA_VERSION`).",
		},
		"extra_java_args": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Extra arguments passed to the `java` command running a JAR application (`CC_EXTRA_JAVA_ARGS`)",
		},
		"jar_path": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Path of the JAR to run, relative to the application root: the application runs as a JAR whatever its profile (`CC_JAR_PATH`)",
		},
		"jar_args": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Arguments passed to the JAR application (`CC_JAR_ARGS`)",
		},
		"maven_build_goal": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Maven goal building the application, `package` by default (`CC_MAVEN_BUILD_GOAL`)",
		},
		"maven_deploy_goal": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Maven goal(s) running the application; setting it builds the application with Maven (`CC_MAVEN_DEPLOY_GOAL`)",
		},
		"maven_profiles": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Maven profile(s) to activate, as passed to `-P` (`CC_MAVEN_PROFILES`)",
		},
		"gradle_build_goal": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Gradle task building the application, `assemble` by default; setting it builds the application with Gradle (`CC_GRADLE_BUILD_GOAL`)",
		},
		"gradle_deploy_goal": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Gradle task(s) running the application (`CC_GRADLE_DEPLOY_GOAL`)",
		},
		"play1_version": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Play 1 version, for example `1.7`; setting it builds the application with Play 1 (`CC_PLAY1_VERSION`)",
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

func (plan *Java) ToEnv(ctx context.Context, diags *diag.Diagnostics) map[string]string {
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

func (plan *Java) FromEnv(ctx context.Context, env *helperMaps.Map[string, string], diags *diag.Diagnostics) {
	if plan == nil || env == nil {
		return
	}

	pkg.ReadEnv(env, plan, envVars...)

	plan.Integrations = attributes.FromEnvIntegrations(ctx, env, plan.Integrations, diags)
}
