package python

import (
	"context"
	_ "embed"

	"go.clever-cloud.com/terraform-provider/pkg/attributes"
	"go.clever-cloud.com/terraform-provider/pkg/resources/application"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/miton18/helper/maps"
	"go.clever-cloud.com/terraform-provider/pkg"
)

type Python struct {
	application.Runtime
	PythonVersion        types.String `tfsdk:"python_version"`
	PipRequirements      types.String `tfsdk:"pip_requirements"`
	Backend              types.String `tfsdk:"backend"`
	Module               types.String `tfsdk:"module"`
	ManageTasks          types.String `tfsdk:"manage_tasks"`
	GunicornWorkerClass  types.String `tfsdk:"gunicorn_worker_class"`
	UwsgiExtraConfig     types.String `tfsdk:"uwsgi_extra_config"`
	CeleryModule         types.String `tfsdk:"celery_module"`
	CeleryLogfile        types.String `tfsdk:"celery_logfile"`
	UvRunCommand         types.String `tfsdk:"uv_run_command"`
	EnvironmentMode      types.String `tfsdk:"environment_mode"`
	CeleryUseBeat        types.Bool   `tfsdk:"celery_use_beat"`
	UseGevent            types.Bool   `tfsdk:"use_gevent"`
	UwsgiInterceptErrors types.Bool   `tfsdk:"uwsgi_intercept_errors"`
}

type PythonV0 struct {
	application.RuntimeV0
	PythonVersion   types.String `tfsdk:"python_version"`
	PipRequirements types.String `tfsdk:"pip_requirements"`
}

//go:embed doc.md
var pythonDoc string

func (r ResourcePython) Schema(ctx context.Context, req resource.SchemaRequest, res *resource.SchemaResponse) {
	res.Schema = schemaPythonV1
}

var schemaPythonV1 = schema.Schema{
	Version:             1,
	MarkdownDescription: pythonDoc,
	Attributes: application.WithRuntimeCommons(map[string]schema.Attribute{
		"python_version": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Python version >= 2.7 (`CC_PYTHON_VERSION`)",
		},
		"pip_requirements": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Define a custom requirements.txt file (default: requirements.txt) (`CC_PIP_REQUIREMENTS_FILE`)",
		},
		"backend": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "WSGI/ASGI server running the application behind Nginx: `uwsgi` (default), `gunicorn`, `daphne` or `uvicorn` (`CC_PYTHON_BACKEND`)",
		},
		"module": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Python module to run, for example `app.server:app` (`CC_PYTHON_MODULE`)",
		},
		"manage_tasks": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Comma-separated list of Django `manage.py` tasks to run during the build (`CC_PYTHON_MANAGE_TASKS`)",
		},
		"gunicorn_worker_class": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Gunicorn worker class: `sync` (default), `eventlet`, `gevent`, `tornado`, `gthread` or `gaiohttp` (`CC_GUNICORN_WORKER_CLASS`)",
		},
		"uwsgi_extra_config": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Additional uWSGI configuration (`CC_UWSGI_EXTRA_CONFIG`)",
		},
		"celery_module": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Celery module to run alongside the application (`CC_PYTHON_CELERY_MODULE`)",
		},
		"celery_logfile": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Log file path for Celery (`CC_PYTHON_CELERY_LOGFILE`)",
		},
		"uv_run_command": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Command starting the application with uv (`CC_PYTHON_UV_RUN_COMMAND`)",
		},
		"environment_mode": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Application environment mode used by uv builds, `production` by default (`ENVIRONMENT`)",
		},
		"celery_use_beat": schema.BoolAttribute{
			Optional:            true,
			MarkdownDescription: "Run Celery Beat with the Celery workers (`CC_PYTHON_CELERY_USE_BEAT`)",
		},
		"use_gevent": schema.BoolAttribute{
			Optional:            true,
			MarkdownDescription: "Run the application with gevent (`CC_PYTHON_USE_GEVENT`)",
		},
		"uwsgi_intercept_errors": schema.BoolAttribute{
			Optional:            true,
			MarkdownDescription: "Let Nginx intercept uWSGI error responses (`UWSGI_INTERCEPT_ERRORS`)",
		},
	}),
	Blocks: attributes.WithBlockRuntimeCommons(map[string]schema.Block{}),
}

var schemaPythonV0 = schema.Schema{
	Version:             0,
	MarkdownDescription: pythonDoc,
	Attributes: application.WithRuntimeCommonsV0(map[string]schema.Attribute{
		// CC_PYTHON_VERSION
		"python_version": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Python version >= 2.7",
		},
		// CC_PIP_REQUIREMENTS_FILE
		"pip_requirements": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Define a custom requirements.txt file (default: requirements.txt)",
		},
	}),
	Blocks: attributes.WithBlockRuntimeCommons(map[string]schema.Block{}),
}

func (py Python) ToEnv(ctx context.Context, diags *diag.Diagnostics) map[string]string {
	env := map[string]string{}

	// do not use the real map since ElementAs can nullish it
	// https://github.com/hashicorp/terraform-plugin-framework/issues/698
	customEnv := map[string]string{}
	diags.Append(py.Environment.ElementsAs(ctx, &customEnv, false)...)
	if diags.HasError() {
		return env
	}
	env = pkg.Merge(env, customEnv)

	pkg.WriteEnv(env, &py, envVars...)

	env = pkg.Merge(env, py.Hooks.ToEnv())
	env = pkg.Merge(env, py.Integrations.ToEnv(ctx, diags))
	return env
}

func (py *Python) FromEnv(ctx context.Context, env *maps.Map[string, string], diags *diag.Diagnostics) {
	if py == nil || env == nil {
		return
	}

	pkg.ReadEnv(env, py, envVars...)

	py.Integrations = attributes.FromEnvIntegrations(ctx, env, py.Integrations, diags)
}
