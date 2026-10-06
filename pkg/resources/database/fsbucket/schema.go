package fsbucket

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
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

type FSBucket struct {
	ID types.String `tfsdk:"id"`

	Name   types.String `tfsdk:"name"`
	Region types.String `tfsdk:"region"`

	Host        types.String `tfsdk:"host"`
	FTPUsername types.String `tfsdk:"ftp_username"`
	FTPPassword types.String `tfsdk:"ftp_password"`
}

//go:embed doc.md
var resourceFSBucketDoc string

func (r ResourceFSBucket) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:             0,
		MarkdownDescription: resourceFSBucketDoc,
		Attributes: map[string]schema.Attribute{
			// customer provided
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Name of the FS Bucket",
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
				MarkdownDescription: "FSBucket FTP endpoint",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"ftp_username": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "FTP username used to authenticate"},
			"ftp_password": schema.StringAttribute{
				Computed:            true,
				Sensitive:           true,
				MarkdownDescription: "FTP password used to authenticate"},
		},
	}
}

// The API-to-state mappers for clevercloud_fsbucket follow.
// See CONTRIBUTING.md § "API → state mapping".

// FromAddon maps the generic add-on view, the only source of name and region
// for this resource.
func (fs *FSBucket) FromAddon(ctx context.Context, addon *tmp.AddonResponse, diags *diag.Diagnostics) *FSBucket {
	if fs == nil || addon == nil {
		return fs
	}

	fs.Name = pkg.FromStr(addon.Name)
	fs.Region = pkg.FromStr(addon.Region)

	return fs
}

// FromEnv maps the FTP credentials, which the bucket only exposes as
// environment variables. Non-destructive — see CONTRIBUTING.md.
func (fs *FSBucket) FromEnv(ctx context.Context, env tmp.EnvVars, diags *diag.Diagnostics) *FSBucket {
	if fs == nil || env == nil {
		return fs
	}

	vars := env.Map()
	fs.Host = pkg.FromStr(vars["BUCKET_HOST"])
	fs.FTPUsername = pkg.FromStr(vars["BUCKET_FTP_USERNAME"])
	fs.FTPPassword = pkg.FromStr(vars["BUCKET_FTP_PASSWORD"])

	return fs
}
