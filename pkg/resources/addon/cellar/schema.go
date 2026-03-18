package cellar

import (
	"context"
	_ "embed"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/resources/addon"
	"go.clever-cloud.com/terraform-provider/pkg/s3"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
	"go.clever-cloud.dev/client"
)

type Cellar struct {
	addon.CommonAttributes

	Host      types.String `tfsdk:"host"`
	KeyID     types.String `tfsdk:"key_id"`
	KeySecret types.String `tfsdk:"key_secret"`
}

func (c *Cellar) GetCommonPtr() *addon.CommonAttributes {
	return &c.CommonAttributes
}

func (c *Cellar) GetAddonOptions() map[string]string {
	return map[string]string{}
}

func (c *Cellar) SetFromResponse(ctx context.Context, cc *client.Client, org string, addonID string, diags *diag.Diagnostics) {
	envRes := tmp.GetAddonEnv(ctx, cc, org, addonID)
	if envRes.HasError() {
		diags.AddError("failed to get addon env vars", envRes.Error().Error())
		return
	}
	envVars := envRes.Payload()

	creds := s3.FromEnvVars(*envVars)
	c.Host = pkg.FromStr(creds.Host)
	c.KeyID = pkg.FromStr(creds.KeyID)
	c.KeySecret = pkg.FromStr(creds.KeySecret)
}

func (c *Cellar) SetDefaults() {}

//go:embed doc.md
var resourceCellarDoc string

func (r ResourceCellar) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:             0,
		MarkdownDescription: resourceCellarDoc,
		Attributes: addon.WithAddonCommons(map[string]schema.Attribute{
			"plan": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Cellar plan (single-plan addon)",
			},
			"host": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "S3 compatible Cellar endpoint",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"key_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Key ID used to authenticate",
			},
			"key_secret": schema.StringAttribute{
				Computed:            true,
				Sensitive:           true,
				MarkdownDescription: "Key secret used to authenticate",
			},
		}),
	}
}
