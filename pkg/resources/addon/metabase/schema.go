package metabase

import (
	"context"
	_ "embed"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/resources/addon"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
	"go.clever-cloud.dev/client"
)

type Metabase struct {
	addon.CommonAttributes
	Host types.String `tfsdk:"host"`
}

func (m *Metabase) GetCommonPtr() *addon.CommonAttributes {
	return &m.CommonAttributes
}

func (m *Metabase) GetAddonOptions() map[string]string {
	return map[string]string{}
}

func (m *Metabase) SetFromResponse(ctx context.Context, cc *client.Client, org string, addonID string, diags *diag.Diagnostics) {
	metabaseRes := tmp.GetMetabase(ctx, cc, m.ID.ValueString())
	if metabaseRes.HasError() {
		diags.AddError("failed to get Metabase", metabaseRes.Error().Error())
		return
	}
	metabase := metabaseRes.Payload()
	m.Host = pkg.FromStr(metabase.AccessURL)
}

func (m *Metabase) SetDefaults() {
	// Metabase has no optional fields requiring defaults
}

//go:embed doc.md
var resourceMetabaseDoc string

func (r ResourceMetabase) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:             0,
		MarkdownDescription: resourceMetabaseDoc,
		Attributes: addon.WithAddonCommons(map[string]schema.Attribute{
			// Single-plan addon: plan is computed, not user-specified
			"plan": schema.StringAttribute{Computed: true},
			"host": schema.StringAttribute{Computed: true, MarkdownDescription: "Metabase host, used to connect to"},
		}),
	}
}
