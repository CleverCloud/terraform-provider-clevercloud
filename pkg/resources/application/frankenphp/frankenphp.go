package frankenphp

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"go.clever-cloud.com/terraform-provider/pkg/resources/application"
)

type ResourceFrankenPHP struct {
	application.Configurer[*FrankenPHP]
}

func NewResourceFrankenPHP() resource.Resource {
	return &ResourceFrankenPHP{}
}

func (r *ResourceFrankenPHP) Metadata(ctx context.Context, req resource.MetadataRequest, res *resource.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_frankenphp"
}

func (r *ResourceFrankenPHP) GetVariantSlug() string {
	return "frankenphp"
}
