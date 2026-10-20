package python

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/resources/application"
	"go.clever-cloud.dev/sdk/app"
)

// pipRequirementsFile has no constant in the SDK.
const pipRequirementsFile = "CC_PIP_REQUIREMENTS_FILE"

// trueToEnv writes "true" when the attribute is true, nothing otherwise.
func trueToEnv(b types.Bool) string {
	if b.ValueBool() {
		return "true"
	}
	return ""
}

var (
	PythonVersionEnv = pkg.NewStringEnv[Python](app.CC_PYTHON_VERSION,
		func(p *Python) types.String { return p.PythonVersion },
		func(p *Python, v types.String) { p.PythonVersion = v })
	PipRequirementsEnv = pkg.NewStringEnv[Python](pipRequirementsFile,
		func(p *Python) types.String { return p.PipRequirements },
		func(p *Python, v types.String) { p.PipRequirements = v })
	BackendEnv = pkg.NewStringEnv[Python](app.CC_PYTHON_BACKEND,
		func(p *Python) types.String { return p.Backend },
		func(p *Python, v types.String) { p.Backend = v })
	ModuleEnv = pkg.NewStringEnv[Python](app.CC_PYTHON_MODULE,
		func(p *Python) types.String { return p.Module },
		func(p *Python, v types.String) { p.Module = v })
	ManageTasksEnv = pkg.NewStringEnv[Python](app.CC_PYTHON_MANAGE_TASKS,
		func(p *Python) types.String { return p.ManageTasks },
		func(p *Python, v types.String) { p.ManageTasks = v })
	GunicornWorkerClassEnv = pkg.NewStringEnv[Python](app.CC_GUNICORN_WORKER_CLASS,
		func(p *Python) types.String { return p.GunicornWorkerClass },
		func(p *Python, v types.String) { p.GunicornWorkerClass = v })
	UwsgiExtraConfigEnv = pkg.NewStringEnv[Python](app.CC_UWSGI_EXTRA_CONFIG,
		func(p *Python) types.String { return p.UwsgiExtraConfig },
		func(p *Python, v types.String) { p.UwsgiExtraConfig = v })
	CeleryModuleEnv = pkg.NewStringEnv[Python](app.CC_PYTHON_CELERY_MODULE,
		func(p *Python) types.String { return p.CeleryModule },
		func(p *Python, v types.String) { p.CeleryModule = v })
	CeleryLogfileEnv = pkg.NewStringEnv[Python](app.CC_PYTHON_CELERY_LOGFILE,
		func(p *Python) types.String { return p.CeleryLogfile },
		func(p *Python, v types.String) { p.CeleryLogfile = v })
	UvRunCommandEnv = pkg.NewStringEnv[Python](app.CC_PYTHON_UV_RUN_COMMAND,
		func(p *Python) types.String { return p.UvRunCommand },
		func(p *Python, v types.String) { p.UvRunCommand = v })
	EnvironmentModeEnv = pkg.NewStringEnv[Python](app.ENVIRONMENT,
		func(p *Python) types.String { return p.EnvironmentMode },
		func(p *Python, v types.String) { p.EnvironmentMode = v })
	CeleryUseBeatEnv = pkg.NewBoolEnv[Python](app.CC_PYTHON_CELERY_USE_BEAT,
		func(p *Python) string { return trueToEnv(p.CeleryUseBeat) },
		func(p *Python, value string) {
			// an absent variable leaves the attribute as it is
			if value == "true" {
				p.CeleryUseBeat = types.BoolValue(true)
			}
		})
	UseGeventEnv = pkg.NewBoolEnv[Python](app.CC_PYTHON_USE_GEVENT,
		func(p *Python) string { return trueToEnv(p.UseGevent) },
		func(p *Python, value string) {
			if value == "true" {
				p.UseGevent = types.BoolValue(true)
			}
		})
	UwsgiInterceptErrorsEnv = pkg.NewBoolEnv[Python](app.UWSGI_INTERCEPT_ERRORS,
		func(p *Python) string { return trueToEnv(p.UwsgiInterceptErrors) },
		func(p *Python, value string) {
			if value == "true" {
				p.UwsgiInterceptErrors = types.BoolValue(true)
			}
		})
)

// envVars lists every python attribute backed by an environment variable.
var envVars = []pkg.EnvVarMapping[*Python]{
	application.OnRuntime[*Python](application.AppFolderEnv),
	PythonVersionEnv,
	PipRequirementsEnv,
	BackendEnv,
	ModuleEnv,
	ManageTasksEnv,
	GunicornWorkerClassEnv,
	UwsgiExtraConfigEnv,
	CeleryModuleEnv,
	CeleryLogfileEnv,
	UvRunCommandEnv,
	EnvironmentModeEnv,
	CeleryUseBeatEnv,
	UseGeventEnv,
	UwsgiInterceptErrorsEnv,
}
