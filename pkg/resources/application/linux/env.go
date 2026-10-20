package linux

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/resources/application"
	"go.clever-cloud.dev/sdk/app"
)

// ccMakefile names the Makefile variable, go.clever-cloud.dev/sdk/app has no
// constant for it.
const ccMakefile = "CC_MAKEFILE"

var (
	RunCommandEnv = pkg.NewStringEnv[Linux](app.CC_RUN_COMMAND,
		func(l *Linux) types.String { return l.RunCommand },
		func(l *Linux, v types.String) { l.RunCommand = v })
	BuildCommandEnv = pkg.NewStringEnv[Linux](app.CC_BUILD_COMMAND,
		func(l *Linux) types.String { return l.BuildCommand },
		func(l *Linux, v types.String) { l.BuildCommand = v })
	MakefileEnv = pkg.NewStringEnv[Linux](ccMakefile,
		func(l *Linux) types.String { return l.Makefile },
		func(l *Linux, v types.String) { l.Makefile = v })
	MiseFilePathEnv = pkg.NewStringEnv[Linux](app.CC_MISE_FILE_PATH,
		func(l *Linux) types.String { return l.MiseFilePath },
		func(l *Linux, v types.String) { l.MiseFilePath = v })
	DisableMiseEnv = pkg.NewBoolEnv[Linux](app.CC_DISABLE_MISE,
		func(l *Linux) string {
			if l.DisableMise.ValueBool() {
				return "true"
			}
			return ""
		},
		func(l *Linux, value string) {
			// the attribute defaults to false: an absent variable means false
			l.DisableMise = types.BoolValue(value == "true")
		})
)

// envVars lists every linux attribute backed by an environment variable.
var envVars = []pkg.EnvVarMapping[*Linux]{
	application.OnRuntime[*Linux](application.AppFolderEnv),
	RunCommandEnv,
	BuildCommandEnv,
	MakefileEnv,
	MiseFilePathEnv,
	DisableMiseEnv,
}
