package materiakv

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
)

type ResourceMateriaKV struct {
	helper.Configurer
}

func NewResourceMateriaKV() resource.Resource {
	return &ResourceMateriaKV{}
}

func (r *ResourceMateriaKV) Metadata(ctx context.Context, req resource.MetadataRequest, res *resource.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_materia_kv"
}
