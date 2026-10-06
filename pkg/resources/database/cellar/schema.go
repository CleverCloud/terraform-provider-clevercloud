package cellar

import (
	"context"
	_ "embed"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/s3"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

type Cellar struct {
	ID types.String `tfsdk:"id"`

	Name   types.String `tfsdk:"name"`
	Region types.String `tfsdk:"region"`

	Host      types.String `tfsdk:"host"`
	KeyID     types.String `tfsdk:"key_id"`
	KeySecret types.String `tfsdk:"key_secret"`
}

//go:embed doc.md
var resourceCellarDoc string

func (r ResourceCellar) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:             0,
		MarkdownDescription: resourceCellarDoc,
		Attributes: map[string]schema.Attribute{
			// customer provided
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Name of the Cellar",
			},
			"region": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Geographical region where the data will be stored",
				Default:             stringdefault.StaticString("par"),
			},

			// provider
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Generated unique identifier",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"host": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "S3 compatible Cellar endpoint",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"key_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Key ID used to authenticate",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"key_secret": schema.StringAttribute{
				Computed:            true,
				Sensitive:           true,
				MarkdownDescription: "Key secret used to authenticate",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		},
	}
}

// The API-to-state mappers for clevercloud_cellar follow.
// See CONTRIBUTING.md § "API → state mapping".

// FromAddon maps the generic add-on view, the only source of name and region
// for this resource.
func (c *Cellar) FromAddon(ctx context.Context, addon *tmp.AddonResponse, diags *diag.Diagnostics) {
	if c == nil || addon == nil {
		return
	}

	c.Name = pkg.FromStr(addon.Name)
	c.Region = pkg.FromStr(addon.Region)
}

// FromEnv maps the S3 credentials, which Cellar only exposes as environment
// variables. Non-destructive — see CONTRIBUTING.md.
func (c *Cellar) FromEnv(ctx context.Context, env tmp.EnvVars, diags *diag.Diagnostics) {
	if c == nil || env == nil {
		return
	}

	credentials := s3.FromEnvVars(env)
	c.Host = pkg.FromStr(credentials.Host)
	c.KeyID = pkg.FromStr(credentials.KeyID)
	c.KeySecret = pkg.FromStr(credentials.KeySecret)
}
