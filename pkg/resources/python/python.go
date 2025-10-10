package python

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
)

type ResourcePython struct {
	helper.Configurer
}

const (
	CC_HTTP_BASIC_AUTH           = "CC_HTTP_BASIC_AUTH"
	CC_NGINX_PROXY_BUFFERS       = "CC_NGINX_PROXY_BUFFERS"
	CC_NGINX_PROXY_BUFFER_SIZE   = "CC_NGINX_PROXY_BUFFER_SIZE"
	CC_PIP_REQUIREMENTS_FILE     = "CC_PIP_REQUIREMENTS_FILE"
	CC_PYTHON_BACKEND            = "CC_PYTHON_BACKEND"
	CC_PYTHON_CELERY_LOGFILE     = "CC_PYTHON_CELERY_LOGFILE"
	CC_PYTHON_CELERY_MODULE      = "CC_PYTHON_CELERY_MODULE"
	CC_PYTHON_CELERY_USE_BEAT    = "CC_PYTHON_CELERY_USE_BEAT"
	CC_PYTHON_MANAGE_TASKS       = "CC_PYTHON_MANAGE_TASKS"
	CC_PYTHON_MODULE             = "CC_PYTHON_MODULE"
	CC_PYTHON_USE_GEVENT         = "CC_PYTHON_USE_GEVENT"
	CC_PYTHON_VERSION            = "CC_PYTHON_VERSION"
	CC_GUNICORN_TIMEOUT          = "CC_GUNICORN_TIMEOUT"
	ENABLE_GZIP_COMPRESSION      = "ENABLE_GZIP_COMPRESSION"
	GZIP_TYPES                   = "GZIP_TYPES"
	GUNICORN_WORKER_CLASS        = "GUNICORN_WORKER_CLASS"
	HARAKIRI                     = "HARAKIRI"
	NGINX_READ_TIMEOUT           = "NGINX_READ_TIMEOUT"
	PYTHON_SETUP_PY_GOAL         = "PYTHON_SETUP_PY_GOAL"
	STATIC_FILES_PATH            = "STATIC_FILES_PATH"
	STATIC_URL_PREFIX            = "STATIC_URL_PREFIX"
	STATIC_WEBROOT               = "STATIC_WEBROOT"
	UWSGI_ASYNC                  = "UWSGI_ASYNC"
	UWSGI_ASYNC_ENGINE           = "UWSGI_ASYNC_ENGINE"
	UWSGI_INTERCEPT_ERRORS       = "UWSGI_INTERCEPT_ERRORS"
	WSGI_BUFFER_SIZE             = "WSGI_BUFFER_SIZE"
	WSGI_POST_BUFFERING          = "WSGI_POST_BUFFERING"
	WSGI_THREADS                 = "WSGI_THREADS"
	WSGI_WORKERS                 = "WSGI_WORKERS"
)

func NewResourcePython() resource.Resource {
	return &ResourcePython{}
}

func (r *ResourcePython) Metadata(ctx context.Context, req resource.MetadataRequest, res *resource.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_python"
}
