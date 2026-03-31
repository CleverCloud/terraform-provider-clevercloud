package oauth_consumer

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
)

// Ensure ResourceOAuthConsumer implements required interfaces
var _ resource.Resource = &ResourceOAuthConsumer{}

type ResourceOAuthConsumer struct {
	helper.Configurer
}

func NewResourceOAuthConsumer() resource.Resource {
	return &ResourceOAuthConsumer{}
}

func (r *ResourceOAuthConsumer) Metadata(ctx context.Context, req resource.MetadataRequest, res *resource.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_oauth_consumer"
}
