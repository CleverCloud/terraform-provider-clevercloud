package ruby

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
)

type ResourceRuby struct {
	helper.Configurer
}

func NewResourceRuby() resource.Resource {
	return &ResourceRuby{}
}

func (r *ResourceRuby) Metadata(ctx context.Context, req resource.MetadataRequest, res *resource.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_ruby"
}
