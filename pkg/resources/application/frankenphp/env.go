package frankenphp

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/resources/application"
	"go.clever-cloud.dev/sdk/app"
)

var (
	DevDependenciesEnv = pkg.NewBoolEnv[FrankenPHP](app.CC_PHP_DEV_DEPENDENCIES,
		func(fp *FrankenPHP) string {
			if fp.DevDependencies.ValueBool() {
				return "install"
			}
			return ""
		},
		func(fp *FrankenPHP, value string) {
			// the attribute defaults to false: an absent variable means false
			fp.DevDependencies = types.BoolValue(value == "install")
		})
	PHPVersionEnv = pkg.NewStringEnv[FrankenPHP](app.CC_PHP_VERSION,
		func(fp *FrankenPHP) types.String { return fp.PHPVersion },
		func(fp *FrankenPHP, v types.String) { fp.PHPVersion = v })
	WebRootEnv = pkg.NewStringEnv[FrankenPHP](app.CC_WEBROOT,
		func(fp *FrankenPHP) types.String { return fp.WebRoot },
		func(fp *FrankenPHP, v types.String) { fp.WebRoot = v })
	ComposerVersionEnv = pkg.NewStringEnv[FrankenPHP](app.CC_COMPOSER_VERSION,
		func(fp *FrankenPHP) types.String { return fp.ComposerVersion },
		func(fp *FrankenPHP, v types.String) { fp.ComposerVersion = v })
	ComposerFlagsEnv = pkg.NewStringEnv[FrankenPHP](app.CC_PHP_COMPOSER_FLAGS,
		func(fp *FrankenPHP) types.String { return fp.ComposerFlags },
		func(fp *FrankenPHP, v types.String) { fp.ComposerFlags = v })
	PortEnv = pkg.NewInt64Env[FrankenPHP](app.CC_FRANKENPHP_PORT,
		func(fp *FrankenPHP) types.Int64 { return fp.Port },
		func(fp *FrankenPHP, v types.Int64) { fp.Port = v })
	WorkerScriptEnv = pkg.NewStringEnv[FrankenPHP](app.CC_FRANKENPHP_WORKER,
		func(fp *FrankenPHP) types.String { return fp.WorkerScript },
		func(fp *FrankenPHP, v types.String) { fp.WorkerScript = v })
)

// envVars lists every frankenphp attribute backed by an environment variable.
var envVars = []pkg.EnvVarMapping[*FrankenPHP]{
	application.OnRuntime[*FrankenPHP](application.AppFolderEnv),
	DevDependenciesEnv,
	PHPVersionEnv,
	WebRootEnv,
	ComposerVersionEnv,
	ComposerFlagsEnv,
	PortEnv,
	WorkerScriptEnv,
}
