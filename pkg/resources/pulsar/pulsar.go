package pulsar

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
)

type ResourcePulsar struct {
	helper.Configurer
}

func NewResourcePulsar() resource.Resource {
	return &ResourcePulsar{}
}

func (r *ResourcePulsar) Metadata(ctx context.Context, req resource.MetadataRequest, res *resource.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_pulsar"
}
