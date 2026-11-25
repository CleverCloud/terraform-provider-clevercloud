package postgresql

import (
	"context"
	_ "embed"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
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
	"go.clever-cloud.com/terraform-provider/pkg/resources/addon"
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
				MarkdownDescription: "PostgreSQL version. Changing it migrates the add-on; the version an existing add-on already runs stays accepted once the platform stops offering it to new add-ons.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
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
