package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

type resourceNodejsType struct{}

const NodejsTypeName = "clevercloud_nodejs"

func init() {
	// Register resource
	AddResource(NodejsTypeName, resourceNodejsType{})
}

func (r resourceNodejsType) NewResource(ctx context.Context, provider tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	p := (provider.(*Provider))
	return ResourceNodeJS{
		cc:  p.cc,
		org: p.Organisation,
	}, nil
}
