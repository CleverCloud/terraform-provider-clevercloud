package python

import (
	"context"
	_ "embed"
	"fmt"
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

type Python struct {
	application.Runtime
	HttpBasicAuth         types.String `tfsdk:"http_basic_auth"`
	NginxProxyBuffers     types.String `tfsdk:"nginx_proxy_buffers"`
	NginxProxyBufferSize  types.String `tfsdk:"nginx_proxy_buffer_size"`
	PipRequirements       types.String `tfsdk:"pip_requirements"`
	PythonBackend         types.String `tfsdk:"python_backend"`
	CeleryLogfile         types.String `tfsdk:"celery_logfile"`
	CeleryModule          types.String `tfsdk:"celery_module"`
	CeleryUseBeat         types.Bool   `tfsdk:"celery_use_beat"`
	ManageTasks           types.String `tfsdk:"manage_tasks"`
	PythonModule          types.String `tfsdk:"python_module"`
	UseGevent             types.Bool   `tfsdk:"use_gevent"`
	PythonVersion         types.String `tfsdk:"python_version"`
	GunicornTimeout       types.Int64  `tfsdk:"gunicorn_timeout"`
	EnableGzipCompression types.Bool   `tfsdk:"enable_gzip_compression"`
	GzipTypes             types.String `tfsdk:"gzip_types"`
	GunicornWorkerClass   types.String `tfsdk:"gunicorn_worker_class"`
	Harakiri              types.Int64  `tfsdk:"harakiri"`
	NginxReadTimeout      types.Int64  `tfsdk:"nginx_read_timeout"`
	SetupPyGoal           types.String `tfsdk:"setup_py_goal"`
	StaticFilesPath       types.String `tfsdk:"static_files_path"`
	StaticURLPrefix       types.String `tfsdk:"static_url_prefix"`
	StaticWebroot         types.String `tfsdk:"static_webroot"`
	UwsgiAsync            types.Int64  `tfsdk:"uwsgi_async"`
	UwsgiAsyncEngine      types.String `tfsdk:"uwsgi_async_engine"`
	UwsgiInterceptErrors  types.Bool   `tfsdk:"uwsgi_intercept_errors"`
	WsgiBufferSize        types.Int64  `tfsdk:"wsgi_buffer_size"`
	WsgiPostBuffering     types.Int64  `tfsdk:"wsgi_post_buffering"`
	WsgiThreads           types.Int64  `tfsdk:"wsgi_threads"`
	WsgiWorkers           types.Int64  `tfsdk:"wsgi_workers"`
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
		"http_basic_auth": schema.StringAttribute{
			Optional:            true,
			Sensitive:           true,
			MarkdownDescription: "Restrict HTTP access to your application. Example: `login:password`. Multiple credentials can be defined using `CC_HTTP_BASIC_AUTH_n`",
		},
		"nginx_proxy_buffers": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Configures the number and size of buffers for reading responses from the proxied server",
		},
		"nginx_proxy_buffer_size": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Sets the size of the buffer for the initial part of the response from the proxied server",
		},
		"pip_requirements": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Specifies a custom requirements.txt file for package installation. Default is `requirements.txt`",
		},
		"python_backend": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Selects the Python backend. Options include `daphne`, `gunicorn`, `uvicorn`, and `uwsgi`. Default is `uwsgi`",
		},
		"celery_logfile": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Sets the relative path to the Celery log file (e.g., `/path/to/logdir`)",
		},
		"celery_module": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Specifies the Celery module to start",
		},
		"celery_use_beat": schema.BoolAttribute{
			Optional:            true,
			Computed:            true,
			Default:             booldefault.StaticBool(false),
			MarkdownDescription: "Set to `true` to enable Celery Beat support",
		},
		"manage_tasks": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "A comma-separated list of Django `manage.py` tasks to execute",
		},
		"python_module": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Defines the Python module to start with, including the path to the application object. Example: `app.server:app` for a `server.py` file in an `/app` folder",
		},
		"use_gevent": schema.BoolAttribute{
			Optional:            true,
			Computed:            true,
			Default:             booldefault.StaticBool(false),
			MarkdownDescription: "Set to `true` to enable Gevent support",
		},
		"python_version": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Selects the Python version. Refer to supported versions documentation",
		},
		"gunicorn_timeout": schema.Int64Attribute{
			Optional:            true,
			MarkdownDescription: "Timeout for Gunicorn workers. Default is `180`",
		},
		"enable_gzip_compression": schema.BoolAttribute{
			Optional:            true,
			Computed:            true,
			Default:             booldefault.StaticBool(false),
			MarkdownDescription: "Set to `true` to enable Gzip compression via Nginx",
		},
		"gzip_types": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Defines the MIME types to be compressed by Gzip. Default is `text/* application/json application/xml application/javascript image/svg+xml`",
		},
		"gunicorn_worker_class": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Gunicorn worker class (e.g., `gevent`, `sync`)",
		},
		"harakiri": schema.Int64Attribute{
			Optional:            true,
			MarkdownDescription: "Timeout in seconds after which an unresponsive process is killed. Default is `180`",
		},
		"nginx_read_timeout": schema.Int64Attribute{
			Optional:            true,
			MarkdownDescription: "Read timeout in seconds for Nginx. Default is `300`",
		},
		"setup_py_goal": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "A custom goal to execute after `requirements.txt` installation",
		},
		"static_files_path": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "The relative path to the directory containing static files (e.g., `path/to/static`)",
		},
		"static_url_prefix": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "The URL path prefix for serving static files. Commonly set to `/public`",
		},
		"static_webroot": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Specifies the web root for static files",
		},
		"uwsgi_async": schema.Int64Attribute{
			Optional:            true,
			MarkdownDescription: "Configures the number of cores for uWSGI asynchronous/non-blocking modes",
		},
		"uwsgi_async_engine": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Selects the asynchronous engine for uWSGI (optional)",
		},
		"uwsgi_intercept_errors": schema.BoolAttribute{
			Optional:            true,
			Computed:            true,
			Default:             booldefault.StaticBool(false),
			MarkdownDescription: "Enables or disables error interception in uWSGI",
		},
		"wsgi_buffer_size": schema.Int64Attribute{
			Optional:            true,
			MarkdownDescription: "Buffer size in bytes for uploads. Default is `4096`",
		},
		"wsgi_post_buffering": schema.Int64Attribute{
			Optional:            true,
			MarkdownDescription: "Maximum size in bytes for request headers. Default is `4096`",
		},
		"wsgi_threads": schema.Int64Attribute{
			Optional:            true,
			MarkdownDescription: "Number of threads per worker. Defaults to automatic setup based on scaler size",
		},
		"wsgi_workers": schema.Int64Attribute{
			Optional:            true,
			MarkdownDescription: "Number of workers. Defaults to automatic setup based on scaler size",
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

func (py Python) toEnv(ctx context.Context, diags *diag.Diagnostics) map[string]string {
	// Start with common runtime environment variables (APP_FOLDER, Hooks, Environment)
	env := py.ToEnv(ctx, diags)
	if diags.HasError() {
		return env
	}

	// Add Python-specific environment variables
	pkg.IfIsSetStr(py.HttpBasicAuth, func(s string) { env[CC_HTTP_BASIC_AUTH] = s })
	pkg.IfIsSetStr(py.NginxProxyBuffers, func(s string) { env[CC_NGINX_PROXY_BUFFERS] = s })
	pkg.IfIsSetStr(py.NginxProxyBufferSize, func(s string) { env[CC_NGINX_PROXY_BUFFER_SIZE] = s })
	pkg.IfIsSetStr(py.PipRequirements, func(s string) { env[CC_PIP_REQUIREMENTS_FILE] = s })
	pkg.IfIsSetStr(py.PythonBackend, func(s string) { env[CC_PYTHON_BACKEND] = s })
	pkg.IfIsSetStr(py.CeleryLogfile, func(s string) { env[CC_PYTHON_CELERY_LOGFILE] = s })
	pkg.IfIsSetStr(py.CeleryModule, func(s string) { env[CC_PYTHON_CELERY_MODULE] = s })
	pkg.IfIsSetB(py.CeleryUseBeat, func(b bool) { env[CC_PYTHON_CELERY_USE_BEAT] = strconv.FormatBool(b) })
	pkg.IfIsSetStr(py.ManageTasks, func(s string) { env[CC_PYTHON_MANAGE_TASKS] = s })
	pkg.IfIsSetStr(py.PythonModule, func(s string) { env[CC_PYTHON_MODULE] = s })
	pkg.IfIsSetB(py.UseGevent, func(b bool) { env[CC_PYTHON_USE_GEVENT] = strconv.FormatBool(b) })
	pkg.IfIsSetStr(py.PythonVersion, func(s string) { env[CC_PYTHON_VERSION] = s })
	pkg.IfIsSetI(py.GunicornTimeout, func(i int64) { env[CC_GUNICORN_TIMEOUT] = fmt.Sprintf("%d", i) })
	pkg.IfIsSetB(py.EnableGzipCompression, func(b bool) { env[ENABLE_GZIP_COMPRESSION] = strconv.FormatBool(b) })
	pkg.IfIsSetStr(py.GzipTypes, func(s string) { env[GZIP_TYPES] = s })
	pkg.IfIsSetStr(py.GunicornWorkerClass, func(s string) { env[GUNICORN_WORKER_CLASS] = s })
	pkg.IfIsSetI(py.Harakiri, func(i int64) { env[HARAKIRI] = fmt.Sprintf("%d", i) })
	pkg.IfIsSetI(py.NginxReadTimeout, func(i int64) { env[NGINX_READ_TIMEOUT] = fmt.Sprintf("%d", i) })
	pkg.IfIsSetStr(py.SetupPyGoal, func(s string) { env[PYTHON_SETUP_PY_GOAL] = s })
	pkg.IfIsSetStr(py.StaticFilesPath, func(s string) { env[STATIC_FILES_PATH] = s })
	pkg.IfIsSetStr(py.StaticURLPrefix, func(s string) { env[STATIC_URL_PREFIX] = s })
	pkg.IfIsSetStr(py.StaticWebroot, func(s string) { env[STATIC_WEBROOT] = s })
	pkg.IfIsSetI(py.UwsgiAsync, func(i int64) { env[UWSGI_ASYNC] = fmt.Sprintf("%d", i) })
	pkg.IfIsSetStr(py.UwsgiAsyncEngine, func(s string) { env[UWSGI_ASYNC_ENGINE] = s })
	pkg.IfIsSetB(py.UwsgiInterceptErrors, func(b bool) { env[UWSGI_INTERCEPT_ERRORS] = strconv.FormatBool(b) })
	pkg.IfIsSetI(py.WsgiBufferSize, func(i int64) { env[WSGI_BUFFER_SIZE] = fmt.Sprintf("%d", i) })
	pkg.IfIsSetI(py.WsgiPostBuffering, func(i int64) { env[WSGI_POST_BUFFERING] = fmt.Sprintf("%d", i) })
	pkg.IfIsSetI(py.WsgiThreads, func(i int64) { env[WSGI_THREADS] = fmt.Sprintf("%d", i) })
	pkg.IfIsSetI(py.WsgiWorkers, func(i int64) { env[WSGI_WORKERS] = fmt.Sprintf("%d", i) })

	return env
}

// fromEnv iter on environment set on the clever application and
// handle language specific env vars
// put the others on Environment field
func (py *Python) fromEnv(ctx context.Context, env map[string]string) {
	m := helper.NewEnvMap(env)

	// Parse Python-specific environment variables
	py.HttpBasicAuth = pkg.FromStr(m.Pop(CC_HTTP_BASIC_AUTH))
	py.NginxProxyBuffers = pkg.FromStr(m.Pop(CC_NGINX_PROXY_BUFFERS))
	py.NginxProxyBufferSize = pkg.FromStr(m.Pop(CC_NGINX_PROXY_BUFFER_SIZE))
	py.PipRequirements = pkg.FromStr(m.Pop(CC_PIP_REQUIREMENTS_FILE))
	py.PythonBackend = pkg.FromStr(m.Pop(CC_PYTHON_BACKEND))
	py.CeleryLogfile = pkg.FromStr(m.Pop(CC_PYTHON_CELERY_LOGFILE))
	py.CeleryModule = pkg.FromStr(m.Pop(CC_PYTHON_CELERY_MODULE))
	py.CeleryUseBeat = pkg.FromBool(m.Pop(CC_PYTHON_CELERY_USE_BEAT) == "true")
	py.ManageTasks = pkg.FromStr(m.Pop(CC_PYTHON_MANAGE_TASKS))
	py.PythonModule = pkg.FromStr(m.Pop(CC_PYTHON_MODULE))
	py.UseGevent = pkg.FromBool(m.Pop(CC_PYTHON_USE_GEVENT) == "true")
	py.PythonVersion = pkg.FromStr(m.Pop(CC_PYTHON_VERSION))

	if timeout, err := strconv.ParseInt(m.Pop(CC_GUNICORN_TIMEOUT), 10, 64); err == nil {
		py.GunicornTimeout = pkg.FromI(timeout)
	}

	py.EnableGzipCompression = pkg.FromBool(m.Pop(ENABLE_GZIP_COMPRESSION) == "true")
	py.GzipTypes = pkg.FromStr(m.Pop(GZIP_TYPES))
	py.GunicornWorkerClass = pkg.FromStr(m.Pop(GUNICORN_WORKER_CLASS))

	if harakiri, err := strconv.ParseInt(m.Pop(HARAKIRI), 10, 64); err == nil {
		py.Harakiri = pkg.FromI(harakiri)
	}

	if nginxTimeout, err := strconv.ParseInt(m.Pop(NGINX_READ_TIMEOUT), 10, 64); err == nil {
		py.NginxReadTimeout = pkg.FromI(nginxTimeout)
	}

	py.SetupPyGoal = pkg.FromStr(m.Pop(PYTHON_SETUP_PY_GOAL))
	py.StaticFilesPath = pkg.FromStr(m.Pop(STATIC_FILES_PATH))
	py.StaticURLPrefix = pkg.FromStr(m.Pop(STATIC_URL_PREFIX))
	py.StaticWebroot = pkg.FromStr(m.Pop(STATIC_WEBROOT))

	if async, err := strconv.ParseInt(m.Pop(UWSGI_ASYNC), 10, 64); err == nil {
		py.UwsgiAsync = pkg.FromI(async)
	}

	py.UwsgiAsyncEngine = pkg.FromStr(m.Pop(UWSGI_ASYNC_ENGINE))
	py.UwsgiInterceptErrors = pkg.FromBool(m.Pop(UWSGI_INTERCEPT_ERRORS) == "true")

	if bufferSize, err := strconv.ParseInt(m.Pop(WSGI_BUFFER_SIZE), 10, 64); err == nil {
		py.WsgiBufferSize = pkg.FromI(bufferSize)
	}

	if postBuffering, err := strconv.ParseInt(m.Pop(WSGI_POST_BUFFERING), 10, 64); err == nil {
		py.WsgiPostBuffering = pkg.FromI(postBuffering)
	}

	if threads, err := strconv.ParseInt(m.Pop(WSGI_THREADS), 10, 64); err == nil {
		py.WsgiThreads = pkg.FromI(threads)
	}

	if workers, err := strconv.ParseInt(m.Pop(WSGI_WORKERS), 10, 64); err == nil {
		py.WsgiWorkers = pkg.FromI(workers)
	}

	// Handle common runtime variables (APP_FOLDER, Hooks, remaining Environment)
	py.FromEnvironment(ctx, m)
}

func (py Python) toDeployment(gitAuth *http.BasicAuth) *application.Deployment {
	if py.Deployment == nil || py.Deployment.Repository.IsNull() {
		return nil
	}

	return &application.Deployment{
		Repository:    py.Deployment.Repository.ValueString(),
		Commit:        py.Deployment.Commit.ValueStringPointer(),
		CleverGitAuth: gitAuth,
	}
}
