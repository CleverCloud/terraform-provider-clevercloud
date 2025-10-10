package frankenphp

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
)

type ResourceFrankenPHP struct {
	helper.Configurer
}

const (
	CC_FRANKENPHP_PORT      = "CC_FRENKENPHP_PORT"
	CC_FRANKENPHP_WORKER    = "CC_FRENKENPHP_WORKER"
	CC_PHP_COMPOSER_FLAGS   = "CC_PHP_COMPOSER_FLAGS"
	CC_PHP_DEV_DEPENDENCIES = "CC_PHP_DEV_DEPENDENCIES"
	CC_WEBROOT              = "CC_WEBROOT"
)

func NewResourceFrankenPHP() resource.Resource {
	return &ResourceFrankenPHP{}
}

func (r *ResourceFrankenPHP) Metadata(ctx context.Context, req resource.MetadataRequest, res *resource.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_frankenphp"
}
