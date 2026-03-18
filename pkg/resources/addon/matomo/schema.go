package matomo

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

type Matomo struct {
	addon.CommonAttributes
	Host    types.String `tfsdk:"host"`
	Version types.String `tfsdk:"version"`
}

func (m *Matomo) GetCommonPtr() *addon.CommonAttributes {
	return &m.CommonAttributes
}

func (m *Matomo) GetAddonOptions() map[string]string {
	return map[string]string{}
}

func (m *Matomo) SetFromResponse(ctx context.Context, cc *client.Client, org string, addonID string, diags *diag.Diagnostics) {
	matomoRes := tmp.GetMatomo(ctx, cc, m.ID.ValueString())
	if matomoRes.HasError() {
		diags.AddError("cannot get matomo", matomoRes.Error().Error())
		return
	}
	matomo := matomoRes.Payload()
	m.Host = pkg.FromStr(matomo.AccessURL)
	m.Version = pkg.FromStr(matomo.Version)
}

func (m *Matomo) SetDefaults() {
	// Matomo has no optional fields requiring defaults
}

//go:embed doc.md
var resourceMatomoDoc string

func (r ResourceMatomo) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:             0,
		MarkdownDescription: resourceMatomoDoc,
		Attributes: addon.WithAddonCommons(map[string]schema.Attribute{
			// Single-plan addon: plan is computed, not user-specified
			"plan":    schema.StringAttribute{Computed: true},
			"host":    schema.StringAttribute{Computed: true, MarkdownDescription: "URL to access Matomo"},
			"version": schema.StringAttribute{Computed: true, MarkdownDescription: "Current version of Matomo"},
		}),
	}
}
