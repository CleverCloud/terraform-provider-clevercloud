package materiakv

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"go.clever-cloud.dev/client"
)

type ResourceMateriaKV struct {
	cc  *client.Client
	org string
}

func NewResourceMateriaKV() resource.Resource {
	return &ResourceMateriaKV{}
}

func (r *ResourceMateriaKV) Metadata(ctx context.Context, req resource.MetadataRequest, res *resource.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_materiadb_kv"
}
