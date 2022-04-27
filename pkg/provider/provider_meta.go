package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func (p *Provider) GetMetaSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	tflog.Info(ctx, "TEST GetMetaSchema()", map[string]interface{}{
		"Configured": p.cc != nil,
	})
	return tfsdk.Schema{}, nil
}
