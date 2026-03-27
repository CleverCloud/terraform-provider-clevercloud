package elasticsearch

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
)

type ResourceElasticsearch struct {
	helper.Configurer
}

func NewResourceElasticsearch() resource.Resource {
	return &ResourceElasticsearch{}
}

func (r *ResourceElasticsearch) Metadata(ctx context.Context, req resource.MetadataRequest, res *resource.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_elasticsearch"
}
