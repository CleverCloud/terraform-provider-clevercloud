package keycloak

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
)

type ResourceKeycloak struct {
	helper.Configurer
}

func NewResourceKeycloak() resource.Resource {
	return &ResourceKeycloak{}
}

func (r *ResourceKeycloak) Metadata(ctx context.Context, req resource.MetadataRequest, res *resource.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_keycloak"
}
