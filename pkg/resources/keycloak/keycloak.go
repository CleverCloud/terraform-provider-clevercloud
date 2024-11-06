package keycloak

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"go.clever-cloud.dev/client"
)

type ResourceKeycloak struct {
	cc  *client.Client
	org string
}

func NewResourceKeycloak() resource.Resource {
	return &ResourceKeycloak{}
}

func (r *ResourceKeycloak) Metadata(ctx context.Context, req resource.MetadataRequest, res *resource.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_keycloak"
}
