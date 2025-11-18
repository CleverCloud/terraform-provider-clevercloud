package rust

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
)

type ResourceRust struct {
	helper.Configurer
}

func NewResourceRust() resource.Resource {
	return &ResourceRust{}
}

func (r *ResourceRust) Metadata(ctx context.Context, req resource.MetadataRequest, res *resource.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_rust"
}

const CC_RUST_FEATURES = "CC_RUST_FEATURES"

func (r *ResourceRust) GetVariantSlug() string {
	return "rust"
}
