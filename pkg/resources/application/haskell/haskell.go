package haskell

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"go.clever-cloud.com/terraform-provider/pkg/resources/application"
)

type ResourceHaskell struct {
	application.Configurer[*Haskell]
}

func NewResourceHaskell() resource.Resource {
	return &ResourceHaskell{}
}

func (r *ResourceHaskell) Metadata(ctx context.Context, req resource.MetadataRequest, res *resource.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_haskell"
}

func (r *ResourceHaskell) GetVariantSlug() string {
	return "haskell"
}
