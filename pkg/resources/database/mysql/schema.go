package mysql

import (
	"context"
	_ "embed"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/resources/addon"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

type MySQL struct {
	addon.CommonAttributes
	Host     types.String `tfsdk:"host"`
	Port     types.Int64  `tfsdk:"port"`
	Database types.String `tfsdk:"database"`
	User     types.String `tfsdk:"user"`
	Password types.String `tfsdk:"password"`
	Version  types.String `tfsdk:"version"`
	Uri      types.String `tfsdk:"uri"`

	Backup        types.Bool `tfsdk:"backup"`
	ReadOnlyUsers types.List `tfsdk:"read_only_users"`
}

//go:embed doc.md
var resourceMysqlDoc string

func (r ResourceMySQL) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:             0,
		MarkdownDescription: resourceMysqlDoc,
		Attributes: addon.WithAddonCommons(map[string]schema.Attribute{
			"host":     schema.StringAttribute{Computed: true, MarkdownDescription: "Database host, used to connect to"},
			"port":     schema.Int64Attribute{Computed: true, MarkdownDescription: "Database port"},
			"database": schema.StringAttribute{Computed: true, MarkdownDescription: "Database name on the MySQL server"},
			"user":     schema.StringAttribute{Computed: true, MarkdownDescription: "Login username"},
			"password": schema.StringAttribute{Computed: true, MarkdownDescription: "Login password", Sensitive: true},
			"uri":      schema.StringAttribute{Computed: true, MarkdownDescription: "Database connection string", Sensitive: true},
			"version": schema.StringAttribute{
				Computed:            true,
				Optional:            true,
				MarkdownDescription: "MySQL version",
				Validators: []validator.String{
					pkg.NewStringValidator("Match existing MySQL version", r.validateMyVersion),
				},
			},
			"backup": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				MarkdownDescription: "Enable or disable backups for this MySQL add-on. Since backups are included in the add-on price, disabling it has no impact on your billing.",
			},
			"read_only_users": schema.ListNestedAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "MySQL users with read-only access",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"user": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "Username for read-only access",
						},
						"password": schema.StringAttribute{
							Required:            true,
							Sensitive:           true,
							MarkdownDescription: "Password for read-only user",
						},
					},
				},
			},
		}),
	}
}

func (r ResourceMySQL) validateMyVersion(ctx context.Context, req validator.StringRequest, res *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	var my MySQL
	res.Diagnostics.Append(req.Config.Get(ctx, &my)...)
	if res.Diagnostics.HasError() {
		return
	}

	requestVersion := my.Version.ValueString()
	region := my.Region.ValueString()
	plan := my.Plan.ValueString()
	infos := r.Infos(ctx, &res.Diagnostics)
	if res.Diagnostics.HasError() {
		return
	}

	switch plan {
	case "dev":
		cluster := pkg.First(infos.Clusters, func(cluster tmp.MysqlCluster) bool {
			return cluster.Region == region
		})
		if cluster == nil {
			res.Diagnostics.Append(diag.NewAttributeErrorDiagnostic(
				req.Path,
				"No MySQL dev cluster found for this region",
				fmt.Sprintf("could not determine dev cluster on region %s", region),
			))
			return
		}

		if cluster.Version != requestVersion {
			res.Diagnostics.Append(diag.NewAttributeErrorDiagnostic(
				req.Path,
				"MySQL version not available on this cluster",
				fmt.Sprintf("Cluster %s is running version %s, not version %s", cluster.Label, cluster.Version, requestVersion),
			))
		}

	default: // on dedicated plan, any available version is OK
		exists := pkg.HasSome(r.dedicatedVersions, func(v string) bool { return v == requestVersion })
		if !exists {
			res.Diagnostics.Append(diag.NewAttributeErrorDiagnostic(
				req.Path,
				"MySQL version not available",
				fmt.Sprintf(
					"version %s not available, available versions: %s",
					requestVersion,
					strings.Join(r.dedicatedVersions, ", "),
				),
			))
		}

	}
}
