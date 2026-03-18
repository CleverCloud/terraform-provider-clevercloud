package mongodb

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/resources/addon"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
	"go.clever-cloud.dev/client"
)

type MongoDB struct {
	addon.CommonAttributes
	Host           types.String `tfsdk:"host"`
	Port           types.Int64  `tfsdk:"port"`
	User           types.String `tfsdk:"user"`
	Password       types.String `tfsdk:"password"`
	Database       types.String `tfsdk:"database"`
	Uri            types.String `tfsdk:"uri"`
	Encryption     types.Bool   `tfsdk:"encryption"`
	DirectHostOnly types.Bool   `tfsdk:"direct_host_only"`
}

func (mg *MongoDB) GetCommonPtr() *addon.CommonAttributes {
	return &mg.CommonAttributes
}

func (mg *MongoDB) GetAddonOptions() map[string]string {
	opts := map[string]string{}
	if !mg.Encryption.IsNull() && !mg.Encryption.IsUnknown() {
		opts["encryption"] = fmt.Sprintf("%t", mg.Encryption.ValueBool())
	}
	if !mg.DirectHostOnly.IsNull() && !mg.DirectHostOnly.IsUnknown() {
		opts["direct-host-only"] = fmt.Sprintf("%t", mg.DirectHostOnly.ValueBool())
	}
	return opts
}

func (mg *MongoDB) SetFromResponse(ctx context.Context, cc *client.Client, _ string, addonID string, diags *diag.Diagnostics) {
	mgInfoRes := tmp.GetMongoDB(ctx, cc, addonID)
	if mgInfoRes.HasError() {
		diags.AddError("failed to get MongoDB connection infos", mgInfoRes.Error().Error())
		return
	}

	addonMG := mgInfoRes.Payload()
	tflog.Debug(ctx, "API response", map[string]any{
		"payload": fmt.Sprintf("%+v", addonMG),
	})

	if addonMG.Status == "TO_DELETE" {
		diags.AddError("addon is being deleted", "MongoDB addon is marked for deletion")
		return
	}

	mg.Host = pkg.FromStr(addonMG.Host)
	mg.Port = pkg.FromI(int64(addonMG.Port))
	mg.User = pkg.FromStr(addonMG.User)
	mg.Password = pkg.FromStr(addonMG.Password)
	mg.Database = pkg.FromStr(addonMG.Database)
	mg.Uri = pkg.FromStr(addonMG.Uri())

	for _, feature := range addonMG.Features {
		switch feature.Name {
		case "encryption":
			mg.Encryption = pkg.FromBool(feature.Enabled)
		case "direct-host-only":
			mg.DirectHostOnly = pkg.FromBool(feature.Enabled)
		}
	}
}

func (mg *MongoDB) SetDefaults() {
	if mg.Encryption.IsUnknown() {
		mg.Encryption = pkg.FromBool(false)
	}
	if mg.DirectHostOnly.IsUnknown() {
		mg.DirectHostOnly = pkg.FromBool(false)
	}
}

//go:embed doc.md
var resourceMongoDBDoc string

func (r ResourceMongoDB) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:             0,
		MarkdownDescription: resourceMongoDBDoc,
		Attributes: addon.WithAddonCommons(map[string]schema.Attribute{
			// customer provided
			"host":     schema.StringAttribute{Computed: true, MarkdownDescription: "Database host, used to connect to"},
			"port":     schema.Int64Attribute{Computed: true, MarkdownDescription: "Database port"},
			"user":     schema.StringAttribute{Computed: true, MarkdownDescription: "Login username"},
			"password": schema.StringAttribute{Computed: true, MarkdownDescription: "Login password", Sensitive: true},
			"database": schema.StringAttribute{Computed: true, MarkdownDescription: "Database name"},
			"uri": schema.StringAttribute{Computed: true, MarkdownDescription: "Database connection string", Sensitive: true},
			"encryption": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Encrypt the hard drive at rest",
				PlanModifiers:       []planmodifier.Bool{boolplanmodifier.RequiresReplace()},
			},
			"direct_host_only": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Connect directly to the database host, bypassing the reverse proxy. Lower latency but no automatic failover on migration.",
				PlanModifiers:       []planmodifier.Bool{boolplanmodifier.RequiresReplace()},
			},
		}),
	}
}
