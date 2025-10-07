package frankenphp

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
)

type ResourceFrankenPHP struct {
	helper.Configurer
}

const CC_PHP_DEV_DEPENDENCIES = "CC_PHP_DEV_DEPENDENCIES"

func NewResourceFrankenPHP() resource.Resource {
	return &ResourceFrankenPHP{}
}

func (r *ResourceFrankenPHP) Metadata(ctx context.Context, req resource.MetadataRequest, res *resource.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_frankenphp"
}
