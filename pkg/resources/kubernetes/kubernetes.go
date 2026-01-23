package kubernetes

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
)

type ResourceKubernetes struct {
	helper.Configurer
}

func NewResourceKubernetes() resource.Resource {
	return &ResourceKubernetes{}
}

func (r *ResourceKubernetes) Metadata(ctx context.Context, req resource.MetadataRequest, res *resource.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_kubernetes"
}
