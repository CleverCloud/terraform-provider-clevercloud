package play2

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
	JavaVersionEnv = pkg.NewStringEnv[Play2](app.CC_JAVA_VERSION,
		func(p *Play2) types.String { return p.JavaVersion },
		func(p *Play2, v types.String) { p.JavaVersion = v })
	SbtBuildGoalEnv = pkg.NewStringEnv[Play2](app.CC_SBT_BUILD_GOAL,
		func(p *Play2) types.String { return p.SbtBuildGoal },
		func(p *Play2, v types.String) { p.SbtBuildGoal = v })
	SbtDeployArgsEnv = pkg.NewStringEnv[Play2](app.CC_SBT_DEPLOY_ARGS,
		func(p *Play2) types.String { return p.SbtDeployArgs },
		func(p *Play2, v types.String) { p.SbtDeployArgs = v })
	SbtTargetBinEnv = pkg.NewStringEnv[Play2](app.CC_SBT_TARGET_BIN,
		func(p *Play2) types.String { return p.SbtTargetBin },
		func(p *Play2, v types.String) { p.SbtTargetBin = v })
	SbtTargetDirEnv = pkg.NewStringEnv[Play2](app.CC_SBT_TARGET_DIR,
		func(p *Play2) types.String { return p.SbtTargetDir },
		func(p *Play2, v types.String) { p.SbtTargetDir = v })
	DisableMaxMetaspaceEnv = pkg.NewBoolEnv[Play2](app.CC_DISABLE_MAX_METASPACE,
		func(p *Play2) string { return trueToEnv(p.DisableMaxMetaspace) },
		func(p *Play2, value string) {
			// an absent variable leaves the attribute as it is
			if value == "true" {
				p.DisableMaxMetaspace = types.BoolValue(true)
			}
		})
	APMAgentEnabledEnv = pkg.NewBoolEnv[Play2](app.CC_JAVA_APM_AGENT_ENABLE,
		func(p *Play2) string { return trueToEnv(p.APMAgentEnabled) },
		func(p *Play2, value string) {
			if value == "true" {
				p.APMAgentEnabled = types.BoolValue(true)
			}
		})
	APMAgentContinuousEnv = pkg.NewBoolEnv[Play2](app.CC_JAVA_APM_AGENT_CONTINUOUS,
		func(p *Play2) string { return trueToEnv(p.APMAgentContinuous) },
		func(p *Play2, value string) {
			if value == "true" {
				p.APMAgentContinuous = types.BoolValue(true)
			}
		})
)

// envVars lists every play2 attribute backed by an environment variable.
var envVars = []pkg.EnvVarMapping[*Play2]{
	application.OnRuntime[*Play2](application.AppFolderEnv),
	JavaVersionEnv,
	SbtBuildGoalEnv,
	SbtDeployArgsEnv,
	SbtTargetBinEnv,
	SbtTargetDirEnv,
	DisableMaxMetaspaceEnv,
	APMAgentEnabledEnv,
	APMAgentContinuousEnv,
}
