package java

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"go.clever-cloud.dev/client"
)

type ResourceJava struct {
	cc  *client.Client
	org string
}

func NewResourceJava() resource.Resource {
	return &ResourceJava{}
}

func (r *ResourceJava) Metadata(ctx context.Context, req resource.MetadataRequest, res *resource.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_java"
}
