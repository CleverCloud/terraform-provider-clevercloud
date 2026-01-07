package dotnet

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"go.clever-cloud.com/terraform-provider/pkg/resources/application"
)

type ResourceDotnet struct {
	application.Configurer[*Dotnet]
}

func NewResourceDotnet() resource.Resource {
	return &ResourceDotnet{}
}

func (r *ResourceDotnet) Metadata(ctx context.Context, req resource.MetadataRequest, res *resource.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_dotnet"
}

func (r *ResourceDotnet) GetVariantSlug() string {
	return "dotnet"
}
