package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/provider"
)

func (p *Provider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "clevercloud"
	resp.Version = p.version
}
