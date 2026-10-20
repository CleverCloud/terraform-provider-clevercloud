package scala

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
	JavaVersionEnv = pkg.NewStringEnv[Scala](app.CC_JAVA_VERSION,
		func(p *Scala) types.String { return p.JavaVersion },
		func(p *Scala, v types.String) { p.JavaVersion = v })
	SbtBuildGoalEnv = pkg.NewStringEnv[Scala](app.CC_SBT_BUILD_GOAL,
		func(p *Scala) types.String { return p.SbtBuildGoal },
		func(p *Scala, v types.String) { p.SbtBuildGoal = v })
	SbtDeployArgsEnv = pkg.NewStringEnv[Scala](app.CC_SBT_DEPLOY_ARGS,
		func(p *Scala) types.String { return p.SbtDeployArgs },
		func(p *Scala, v types.String) { p.SbtDeployArgs = v })
	SbtTargetBinEnv = pkg.NewStringEnv[Scala](app.CC_SBT_TARGET_BIN,
		func(p *Scala) types.String { return p.SbtTargetBin },
		func(p *Scala, v types.String) { p.SbtTargetBin = v })
	SbtTargetDirEnv = pkg.NewStringEnv[Scala](app.CC_SBT_TARGET_DIR,
		func(p *Scala) types.String { return p.SbtTargetDir },
		func(p *Scala, v types.String) { p.SbtTargetDir = v })
	DisableMaxMetaspaceEnv = pkg.NewBoolEnv[Scala](app.CC_DISABLE_MAX_METASPACE,
		func(p *Scala) string { return trueToEnv(p.DisableMaxMetaspace) },
		func(p *Scala, value string) {
			// an absent variable leaves the attribute as it is
			if value == "true" {
				p.DisableMaxMetaspace = types.BoolValue(true)
			}
		})
	APMAgentEnabledEnv = pkg.NewBoolEnv[Scala](app.CC_JAVA_APM_AGENT_ENABLE,
		func(p *Scala) string { return trueToEnv(p.APMAgentEnabled) },
		func(p *Scala, value string) {
			if value == "true" {
				p.APMAgentEnabled = types.BoolValue(true)
			}
		})
	APMAgentContinuousEnv = pkg.NewBoolEnv[Scala](app.CC_JAVA_APM_AGENT_CONTINUOUS,
		func(p *Scala) string { return trueToEnv(p.APMAgentContinuous) },
		func(p *Scala, value string) {
			if value == "true" {
				p.APMAgentContinuous = types.BoolValue(true)
			}
		})
)

// envVars lists every scala attribute backed by an environment variable.
var envVars = []pkg.EnvVarMapping[*Scala]{
	application.OnRuntime[*Scala](application.AppFolderEnv),
	JavaVersionEnv,
	SbtBuildGoalEnv,
	SbtDeployArgsEnv,
	SbtTargetBinEnv,
	SbtTargetDirEnv,
	DisableMaxMetaspaceEnv,
	APMAgentEnabledEnv,
	APMAgentContinuousEnv,
}
