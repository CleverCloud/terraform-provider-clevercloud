package postgresqlbackup

import (
	"context"
	_ "embed"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type PostgreSQLBackup struct {
	PostgresqlID types.String `tfsdk:"postgresql_id"`
	Selector     types.String `tfsdk:"selector"`

	// Computed attributes
	ID           types.String `tfsdk:"id"`
	DownloadURL  types.String `tfsdk:"download_url"`
	CreationDate types.String `tfsdk:"creation_date"`
	DeletionDate types.String `tfsdk:"deletion_date"`
}

//go:embed doc.md
var postgresqlBackupDoc string

func (d *DataSourcePostgreSQLBackup) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Retrieves information about a specific PostgreSQL backup from Clever Cloud.",
		MarkdownDescription: postgresqlBackupDoc,
		Attributes: map[string]schema.Attribute{
			"postgresql_id": schema.StringAttribute{
				Required:    true,
				Description: "The PostgreSQL ID (format: postgresql_xxx)",
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^postgresql_[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`),
						"must be a valid PostgreSQL addon ID (format: postgresql_xxx)",
					),
				},
			},
			"selector": schema.StringAttribute{
				Required: true,
				Description: "One of three options: " +
					"(1) A backup UUID for exact match (e.g., '550e8400-e29b-41d4-a716-446655440000'), " +
					"(2) An ISO8601 date string to find the most recent backup before this date (e.g., '2025-12-23T10:00:00Z'), " +
					"(3) The string 'latest' to get the most recent backup",
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The backup UUID",
			},
			"download_url": schema.StringAttribute{
				Computed: true,
				Description: "The signed S3 URL to download the backup archive. " +
					"Note: This URL expires after 2 hours. Retrieve fresh datasource before downloading.",
			},
			"creation_date": schema.StringAttribute{
				Computed:    true,
				Description: "The ISO8601 timestamp when the backup was created",
			},
			"deletion_date": schema.StringAttribute{
				Computed:    true,
				Description: "The ISO8601 timestamp when the backup will be deleted",
			},
		},
	}
}
