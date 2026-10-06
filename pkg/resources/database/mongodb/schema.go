package mongodb

import (
	"context"
	_ "embed"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/resources/addon"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
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
			"uri":      schema.StringAttribute{Computed: true, MarkdownDescription: "Database connection string", Sensitive: true},
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

// The API-to-state mappers for clevercloud_mongodb follow.
// See CONTRIBUTING.md § "API → state mapping".

// FromAddon maps the generic add-on view. It is the only source of name, plan,
// region and creation_date: tmp.MongoDB carries none of them.
func (mg *MongoDB) FromAddon(ctx context.Context, addon *tmp.AddonResponse, diags *diag.Diagnostics) *MongoDB {
	if mg == nil || addon == nil {
		return mg
	}

	mg.Name = pkg.FromStr(addon.Name)
	mg.Plan = pkg.FromStr(addon.Plan.Slug)
	mg.Region = pkg.FromStr(addon.Region)
	mg.CreationDate = pkg.FromI(addon.CreationDate)

	return mg
}

// FromMongoDB maps the product view: connection details and the feature toggles.
//
// The feature list is sparse and both toggles are Computed, so an absent one
// resolves to its fallback and never to a null — a null is what left an
// imported add-on with a non-empty plan (#404).
func (mg *MongoDB) FromMongoDB(ctx context.Context, api *tmp.MongoDB, diags *diag.Diagnostics) *MongoDB {
	if mg == nil || api == nil {
		return mg
	}

	mg.Host = pkg.FromStr(api.Host)
	mg.Port = pkg.FromI(api.Port)
	mg.User = pkg.FromStr(api.User)
	mg.Password = pkg.FromStr(api.Password)
	mg.Database = pkg.FromStr(api.Database)
	mg.Uri = pkg.FromStr(api.Uri())

	features := pkg.FeaturesOf(api.Features, func(f tmp.MongoDBFeature) (string, bool) {
		return f.Name, f.Enabled
	})
	features.Or(&mg.Encryption, "encryption", false)
	features.Or(&mg.DirectHostOnly, "direct-host-only", false)

	return mg
}
