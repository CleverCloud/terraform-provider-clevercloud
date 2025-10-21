package matomo

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
)

type ResourceMatomo struct {
	helper.Configurer
}

func NewResourceMatomo() resource.Resource {
	return &ResourceMatomo{}
}

func (r *ResourceMatomo) Metadata(ctx context.Context, req resource.MetadataRequest, res *resource.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_matomo"
}
