package rust

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
)

type ResourceRust struct {
	helper.Configurer
}

const (
	CC_RUSTUP_CHANNEL = "CC_RUSTUP_CHANNEL"
	CC_RUST_BIN       = "CC_RUST_BIN"
	CC_RUST_FEATURES  = "CC_RUST_FEATURES"
	CC_RUN_COMMAND    = "CC_RUN_COMMAND"
)

func NewResourceRust() resource.Resource {
	return &ResourceRust{}
}

func (r *ResourceRust) Metadata(ctx context.Context, req resource.MetadataRequest, res *resource.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_rust"
}
