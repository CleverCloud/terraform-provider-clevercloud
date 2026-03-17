package fsbucket

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
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
	"go.clever-cloud.dev/client"
)

type FSBucket struct {
	addon.CommonAttributes

	Host        types.String `tfsdk:"host"`
	FTPUsername types.String `tfsdk:"ftp_username"`
	FTPPassword types.String `tfsdk:"ftp_password"`
}

func (f *FSBucket) GetCommonPtr() *addon.CommonAttributes {
	return &f.CommonAttributes
}

func (f *FSBucket) GetAddonOptions() map[string]string {
	return map[string]string{}
}

func (f *FSBucket) SetFromResponse(ctx context.Context, cc *client.Client, org string, addonID string, diags *diag.Diagnostics) {
	envRes := tmp.GetAddonEnv(ctx, cc, org, addonID)
	if envRes.HasError() {
		diags.AddError("failed to get addon env vars", envRes.Error().Error())
		return
	}
	envVars := envRes.Payload()

	envMap := pkg.Reduce(*envVars, map[string]string{}, func(m map[string]string, v tmp.EnvVar) map[string]string {
		m[v.Name] = v.Value
		return m
	})

	f.Host = pkg.FromStr(envMap["BUCKET_HOST"])
	f.FTPUsername = pkg.FromStr(envMap["BUCKET_FTP_USERNAME"])
	f.FTPPassword = pkg.FromStr(envMap["BUCKET_FTP_PASSWORD"])
}

func (f *FSBucket) SetDefaults() {}

//go:embed doc.md
var resourceFSBucketDoc string

func (r ResourceFSBucket) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:             0,
		MarkdownDescription: resourceFSBucketDoc,
		Attributes: addon.WithAddonCommons(map[string]schema.Attribute{
			"plan": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "FS Bucket plan (single-plan addon)",
			},
			"host": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "FSBucket FTP endpoint",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"ftp_username": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "FTP username used to authenticate",
			},
			"ftp_password": schema.StringAttribute{
				Computed:            true,
				Sensitive:           true,
				MarkdownDescription: "FTP password used to authenticate",
			},
		}),
	}
}
