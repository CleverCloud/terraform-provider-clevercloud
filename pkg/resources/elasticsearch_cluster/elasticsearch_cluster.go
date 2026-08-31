package elasticsearch_cluster

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
)

type ResourceElasticsearchCluster struct {
	helper.Configurer
}

func NewResourceElasticsearchCluster() resource.Resource {
	return &ResourceElasticsearchCluster{}
}

func (r *ResourceElasticsearchCluster) Metadata(ctx context.Context, req resource.MetadataRequest, res *resource.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_elasticsearch_cluster"
}
