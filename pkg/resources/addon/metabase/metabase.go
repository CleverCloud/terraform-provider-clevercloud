package metabase

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
)

type ResourceMetabase struct {
	helper.Configurer
}

func NewResourceMetabase() resource.Resource {
	return &ResourceMetabase{}
}

func (r *ResourceMetabase) Metadata(ctx context.Context, req resource.MetadataRequest, res *resource.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_metabase"
}
