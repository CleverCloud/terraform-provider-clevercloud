package fsbucket

import (
	"context"
	_ "embed"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type FSBucket struct {
	ID types.String `tfsdk:"id"`

	Name    types.String `tfsdk:"name"`
	Region  types.String `tfsdk:"region"`
	Tags    types.Set    `tfsdk:"tags"`
	TagsAll types.Set    `tfsdk:"tags_all"`

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
			"tags": schema.SetAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "Tags of the addon",
			},
			"tags_all": schema.SetAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				MarkdownDescription: "All tags applied to the addon: the resource `tags` merged with the provider-level `default_tags`",
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
