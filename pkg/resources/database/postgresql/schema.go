package postgresql

import (
	"context"
	_ "embed"
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/resources/addon"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

type PostgreSQL struct {
	addon.CommonAttributes
	Host     types.String `tfsdk:"host"`
	Port     types.Int64  `tfsdk:"port"`
	Database types.String `tfsdk:"database"`
	User     types.String `tfsdk:"user"`
	Password types.String `tfsdk:"password"`
	Version  types.String `tfsdk:"version"`
	Uri      types.String `tfsdk:"uri"`

	Backup         types.Bool   `tfsdk:"backup"`
	Encryption     types.Bool   `tfsdk:"encryption"`
	DirectHostOnly types.Bool   `tfsdk:"direct_host_only"`
	Locale         types.String `tfsdk:"locale"`
}

//go:embed doc.md
var resourcePostgresqlDoc string

func (r ResourcePostgreSQL) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:             1, // Incremented from 0 due to locale type change from Bool to String
		MarkdownDescription: resourcePostgresqlDoc,
		Attributes: addon.WithAddonCommons(map[string]schema.Attribute{
			"host":     schema.StringAttribute{Computed: true, MarkdownDescription: "Database host, used to connect to", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"port":     schema.Int64Attribute{Computed: true, MarkdownDescription: "Database port", PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()}},
			"database": schema.StringAttribute{Computed: true, MarkdownDescription: "Database name on the PostgreSQL server", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"user":     schema.StringAttribute{Computed: true, MarkdownDescription: "Login username", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"password": schema.StringAttribute{Computed: true, MarkdownDescription: "Login password", Sensitive: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"uri":      schema.StringAttribute{Computed: true, MarkdownDescription: "Database connection string", Sensitive: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"version": schema.StringAttribute{
				Computed:            true,
				Optional:            true,
				MarkdownDescription: "PostgreSQL version",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
				Validators: []validator.String{
					pkg.NewStringValidator("Match existing PostgresQL version", r.validatePGVersion),
				},
			},
			"backup": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				PlanModifiers:       []planmodifier.Bool{boolplanmodifier.RequiresReplace()},
				MarkdownDescription: "Enable or disable backups for this PostgreSQL add-on. Since backups are included in the add-on price, disabling it has no impact on your billing.",
			},
			"encryption": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Encrypt the hard drive at rest",
				PlanModifiers:       []planmodifier.Bool{boolplanmodifier.UseStateForUnknown(), boolplanmodifier.RequiresReplace()},
			},
			"direct_host_only": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Connect directly to the database host, bypassing the reverse proxy. Lower latency but no automatic failover on migration.",
				PlanModifiers:       []planmodifier.Bool{boolplanmodifier.UseStateForUnknown(), boolplanmodifier.RequiresReplace()},
			},
			"locale": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Database locale for collation and character classification. Must be in format 'language_COUNTRY' (e.g., 'en_GB', 'fr_FR'). Only available on dedicated plans. If not specified, defaults to 'en_GB'.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
				Default:             stringdefault.StaticString("en_GB"),
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[a-z]{2}_[A-Z]{2}$`),
						"must be in format 'language_COUNTRY' (e.g., 'en_GB', 'fr_FR', 'de_DE')",
					),
				},
			},
		}),
	}
}

func (r ResourcePostgreSQL) validatePGVersion(ctx context.Context, req validator.StringRequest, res *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	var pg PostgreSQL
	res.Diagnostics.Append(req.Config.Get(ctx, &pg)...)
	if res.Diagnostics.HasError() {
		return
	}

	requestVersion := pg.Version.ValueString()
	region := pg.Region.ValueString()
	plan := pg.Plan.ValueString()
	infos := r.Infos(ctx, &res.Diagnostics)
	if res.Diagnostics.HasError() {
		return
	}

	// Skip validation if infos not available (provider not configured yet)
	if infos == nil {
		return
	}

	switch plan {
	case "dev":
		cluster := pkg.First(infos.Clusters, func(cluster tmp.PostgresCluster) bool {
			return cluster.Region == region
		})
		if cluster == nil {
			res.Diagnostics.Append(diag.NewAttributeErrorDiagnostic(
				req.Path,
				"No PostgreSQL dev cluster found for this region",
				fmt.Sprintf("could not determine dev cluster on region %s", region),
			))
			return
		}

		if cluster.Version != requestVersion {
			res.Diagnostics.Append(diag.NewAttributeErrorDiagnostic(
				req.Path,
				"PostgreSQL version not available on this cluster",
				fmt.Sprintf("Cluster %s is running version %s, not version %s", cluster.Label, cluster.Version, requestVersion),
			))
		}

	default: // on dedicated plan, any available version is OK
		exists := pkg.HasSome(r.dedicatedVersions, func(v string) bool { return v == requestVersion })
		if !exists {
			res.Diagnostics.Append(diag.NewAttributeErrorDiagnostic(
				req.Path,
				"PostgreSQL version not available",
				fmt.Sprintf(
					"version %s not available, available versions: %s",
					requestVersion,
					strings.Join(r.dedicatedVersions, ", "),
				),
			))
		}

	}
}

// The API-to-state mappers for clevercloud_postgresql follow.
// See CONTRIBUTING.md § "API → state mapping".

// defaultLocale is the locale the platform uses when none is requested at creation
const defaultLocale = "en_GB"

// FromAddon maps the generic add-on view. It is the only source of name, plan,
// region and creation_date: tmp.PostgreSQL carries none of them.
func (pg *PostgreSQL) FromAddon(ctx context.Context, addon *tmp.AddonResponse, diags *diag.Diagnostics) *PostgreSQL {
	if pg == nil || addon == nil {
		return pg
	}

	pg.Name = pkg.FromStr(addon.Name)
	pg.Plan = pkg.FromStr(addon.Plan.Slug)
	pg.Region = pkg.FromStr(addon.Region)
	pg.CreationDate = pkg.FromI(addon.CreationDate)

	return pg
}

// FromPostgreSQL maps the product view: connection details, locale and the
// feature toggles.
//
// The feature list is sparse and all three toggles are Computed, so an absent
// one resolves to its fallback and never to a null — a null is what left an
// imported add-on with a non-empty plan (#404). backup carries a schema
// default, which its fallback has to match.
func (pg *PostgreSQL) FromPostgreSQL(ctx context.Context, api *tmp.PostgreSQL, diags *diag.Diagnostics) *PostgreSQL {
	if pg == nil || api == nil {
		return pg
	}

	pg.Host = pkg.FromStr(api.Host)
	pg.Port = pkg.FromI(int64(api.Port))
	pg.Database = pkg.FromStr(api.Database)
	pg.User = pkg.FromStr(api.User)
	pg.Password = pkg.FromStr(api.Password)
	pg.Version = pkg.FromStr(api.Version)
	pg.Uri = pkg.FromStr(api.Uri())

	// The locale is fixed at creation time (LC_COLLATE cannot change afterwards)
	// and exposed by the API. When the API does not return it, keep the value
	// already in state (config or default) and only fall back on import.
	switch {
	case api.Locale != "":
		pg.Locale = pkg.FromStr(api.Locale)
	case pg.Locale.IsNull() || pg.Locale.IsUnknown():
		tflog.Warn(ctx, "API did not return the PostgreSQL locale, using default", map[string]any{"locale": defaultLocale})
		pg.Locale = pkg.FromStr(defaultLocale)
	}

	features := pkg.FeaturesOf(api.Features, func(f tmp.PostgreSQLFeature) (string, bool) {
		return f.Name, f.Enabled
	})
	features.Or(&pg.Backup, "do-backup", true) // schema default
	features.Or(&pg.Encryption, "encryption", false)
	features.Or(&pg.DirectHostOnly, "direct-host-only", false)

	return pg
}
