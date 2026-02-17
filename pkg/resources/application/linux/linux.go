package linux

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"go.clever-cloud.com/terraform-provider/pkg/resources/application"
)

type ResourceLinux struct {
	application.Configurer[*Linux]
}

func NewResourceLinux() resource.Resource {
	return &ResourceLinux{}
}

func (r *ResourceLinux) Metadata(ctx context.Context, req resource.MetadataRequest, res *resource.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_linux"
}

func (r *ResourceLinux) GetVariantSlug() string {
	return "linux"
}
