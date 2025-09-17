package python

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
)

type ResourcePython struct {
	helper.Configurer
}

func NewResourcePython() resource.Resource {
	return &ResourcePython{}
}

func (r *ResourcePython) Metadata(ctx context.Context, req resource.MetadataRequest, res *resource.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_python"
}
