package nodegroup

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
)

type ResourceKubernetesNodegroup struct {
	helper.Configurer
}

func NewResourceKubernetesNodegroup() resource.Resource {
	return &ResourceKubernetesNodegroup{}
}

func (r *ResourceKubernetesNodegroup) Metadata(ctx context.Context, req resource.MetadataRequest, res *resource.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_kubernetes_nodegroup"
}
