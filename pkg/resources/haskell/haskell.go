package haskell

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
)

type ResourceHaskell struct {
	helper.Configurer
}

func NewResourceHaskell() resource.Resource {
	return &ResourceHaskell{}
}

func (r *ResourceHaskell) Metadata(ctx context.Context, req resource.MetadataRequest, res *resource.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_haskell"
}
