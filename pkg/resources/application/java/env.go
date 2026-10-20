package java

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/resources/application"
	"go.clever-cloud.dev/sdk/app"
)

// trueToEnv writes "true" when the attribute is true, nothing otherwise.
func trueToEnv(b types.Bool) string {
	if b.ValueBool() {
		return "true"
	}
	return ""
}

var (
	JavaVersionEnv = pkg.NewStringEnv[Java](app.CC_JAVA_VERSION,
		func(j *Java) types.String { return j.JavaVersion },
		func(j *Java, v types.String) { j.JavaVersion = v })
	ExtraJavaArgsEnv = pkg.NewStringEnv[Java](app.CC_EXTRA_JAVA_ARGS,
		func(j *Java) types.String { return j.ExtraJavaArgs },
		func(j *Java, v types.String) { j.ExtraJavaArgs = v })
	JarPathEnv = pkg.NewStringEnv[Java](app.CC_JAR_PATH,
		func(j *Java) types.String { return j.JarPath },
		func(j *Java, v types.String) { j.JarPath = v })
	JarArgsEnv = pkg.NewStringEnv[Java](app.CC_JAR_ARGS,
		func(j *Java) types.String { return j.JarArgs },
		func(j *Java, v types.String) { j.JarArgs = v })
	MavenBuildGoalEnv = pkg.NewStringEnv[Java](app.CC_MAVEN_BUILD_GOAL,
		func(j *Java) types.String { return j.MavenBuildGoal },
		func(j *Java, v types.String) { j.MavenBuildGoal = v })
	MavenDeployGoalEnv = pkg.NewStringEnv[Java](app.CC_MAVEN_DEPLOY_GOAL,
		func(j *Java) types.String { return j.MavenDeployGoal },
		func(j *Java, v types.String) { j.MavenDeployGoal = v })
	MavenProfilesEnv = pkg.NewStringEnv[Java](app.CC_MAVEN_PROFILES,
		func(j *Java) types.String { return j.MavenProfiles },
		func(j *Java, v types.String) { j.MavenProfiles = v })
	GradleBuildGoalEnv = pkg.NewStringEnv[Java](app.CC_GRADLE_BUILD_GOAL,
		func(j *Java) types.String { return j.GradleBuildGoal },
		func(j *Java, v types.String) { j.GradleBuildGoal = v })
	GradleDeployGoalEnv = pkg.NewStringEnv[Java](app.CC_GRADLE_DEPLOY_GOAL,
		func(j *Java) types.String { return j.GradleDeployGoal },
		func(j *Java, v types.String) { j.GradleDeployGoal = v })
	Play1VersionEnv = pkg.NewStringEnv[Java](app.CC_PLAY1_VERSION,
		func(j *Java) types.String { return j.Play1Version },
		func(j *Java, v types.String) { j.Play1Version = v })
	DisableMaxMetaspaceEnv = pkg.NewBoolEnv[Java](app.CC_DISABLE_MAX_METASPACE,
		func(j *Java) string { return trueToEnv(j.DisableMaxMetaspace) },
		func(j *Java, value string) {
			// an absent variable leaves the attribute as it is
			if value == "true" {
				j.DisableMaxMetaspace = types.BoolValue(true)
			}
		})
	APMAgentEnabledEnv = pkg.NewBoolEnv[Java](app.CC_JAVA_APM_AGENT_ENABLE,
		func(j *Java) string { return trueToEnv(j.APMAgentEnabled) },
		func(j *Java, value string) {
			if value == "true" {
				j.APMAgentEnabled = types.BoolValue(true)
			}
		})
	APMAgentContinuousEnv = pkg.NewBoolEnv[Java](app.CC_JAVA_APM_AGENT_CONTINUOUS,
		func(j *Java) string { return trueToEnv(j.APMAgentContinuous) },
		func(j *Java, value string) {
			if value == "true" {
				j.APMAgentContinuous = types.BoolValue(true)
			}
		})
)

// envVars lists every java attribute backed by an environment variable.
var envVars = []pkg.EnvVarMapping[*Java]{
	application.OnRuntime[*Java](application.AppFolderEnv),
	JavaVersionEnv,
	ExtraJavaArgsEnv,
	JarPathEnv,
	JarArgsEnv,
	MavenBuildGoalEnv,
	MavenDeployGoalEnv,
	MavenProfilesEnv,
	GradleBuildGoalEnv,
	GradleDeployGoalEnv,
	Play1VersionEnv,
	DisableMaxMetaspaceEnv,
	APMAgentEnabledEnv,
	APMAgentContinuousEnv,
}
