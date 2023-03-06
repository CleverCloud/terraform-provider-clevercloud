package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func (p *Provider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	tflog.Info(ctx, "TEST Metadata()", map[string]interface{}{
		"Configured": p.cc != nil,
	})

	resp.TypeName = "clevercloud"
	resp.Version = p.version
}
