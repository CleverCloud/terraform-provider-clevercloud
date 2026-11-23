package postgresqlbackup

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.dev/client"
	"go.clever-cloud.dev/sdk/models"
)

// ListBackups returns the backups of a database add-on, by its real ID. The route answers one
// list whose entries are either database archives or Elasticsearch snapshots; a PostgreSQL
// add-on only has archives.
func ListBackups(ctx context.Context, cc *client.Client, organisationID, realID string) ([]models.AddonBackup, error) {
	res := pkg.SDK(cc).V2().Backups().Ownerid(organisationID).Ref(realID).Getbackups(ctx)
	if res.HasError() {
		return nil, res.Error()
	}

	entries := *res.Payload()
	backups := make([]models.AddonBackup, 0, len(entries))
	for _, entry := range entries {
		var backup models.AddonBackup
		if err := entry.Decode(&backup); err != nil {
			return nil, fmt.Errorf("failed to decode a backup: %w", err)
		}
		backups = append(backups, backup)
	}
	return backups, nil
}

func (d *DataSourcePostgreSQLBackup) Read(ctx context.Context, req datasource.ReadRequest, res *datasource.ReadResponse) {
	config := helper.From[PostgreSQLBackup](ctx, req.Config, &res.Diagnostics)
	if res.Diagnostics.HasError() {
		return
	}

	postgresqlID := config.PostgresqlID.ValueString()
	selector := config.Selector.ValueString()

	tflog.Debug(ctx, "Reading PostgreSQL backup", map[string]any{
		"postgresql_id": postgresqlID,
		"selector":      selector,
	})

	backups, err := ListBackups(ctx, d.Client(), d.Organization(), postgresqlID)
	if err != nil {
		res.Diagnostics.AddError("Failed to get PostgreSQL backups", err.Error())
		return
	}

	if len(backups) == 0 {
		res.Diagnostics.AddError(
			"No backups found",
			fmt.Sprintf("No backups exist for PostgreSQL addon %s. Backups are created automatically 24 hours after addon creation.", postgresqlID),
		)
		return
	}

	tflog.Debug(ctx, "Retrieved backups", map[string]any{"count": len(backups)})

	var selectedBackup *models.AddonBackup

	if selector == "latest" {
		selectedBackup = findLatestBackup(backups, &res.Diagnostics)
	} else if isUUID(selector) {
		selectedBackup = findBackupByUUID(backups, selector, &res.Diagnostics)
	} else {
		selectedBackup = findBackupByDate(backups, selector, &res.Diagnostics)
	}
	if res.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Found backup by date", map[string]any{
		"backup_id":     selectedBackup.BackupID,
		"creation_date": selectedBackup.CreationDate,
	})

	// Map to Terraform state
	config.ID = types.StringValue(selectedBackup.BackupID)
	config.DownloadURL = types.StringValue(selectedBackup.DownloadURL)
	config.CreationDate = types.StringValue(selectedBackup.CreationDate.Format(time.RFC3339))
	config.DeletionDate = types.StringValue(selectedBackup.DeleteAt.Format(time.RFC3339))

	res.Diagnostics.Append(res.State.Set(ctx, config)...)
}

func isUUID(s string) bool {
	return uuid.Validate(s) == nil
}

func findBackupByUUID(backups []models.AddonBackup, backupUUID string, diags *diag.Diagnostics) *models.AddonBackup {
	for i := range backups {
		if backups[i].BackupID == backupUUID {
			return &backups[i]
		}
	}

	diags.AddError("Backup not found", fmt.Sprintf("No backup found with UUID %s ", backupUUID))
	return nil
}

func findLatestBackup(backups []models.AddonBackup, diags *diag.Diagnostics) *models.AddonBackup {
	if len(backups) == 0 {
		diags.AddError("No backup found", "there is no backup available")
		return nil
	}

	sort.Slice(backups, func(i, j int) bool {
		return backups[i].CreationDate.After(backups[j].CreationDate)
	})

	return &backups[0]
}

func findBackupByDate(backups []models.AddonBackup, dateStr string, diags *diag.Diagnostics) *models.AddonBackup {
	targetDate, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		diags.AddError(
			"Invalid date format",
			fmt.Sprintf("Expected ISO8601/RFC3339 format (e.g., '2025-12-23T10:00:00Z'), got: %s. Error: %s", dateStr, err.Error()),
		)
		return nil
	}

	validBackups := pkg.Filter(backups, func(backup models.AddonBackup) bool {
		creationTime := backup.CreationDate
		return creationTime.Before(targetDate) || creationTime.Equal(targetDate)
	})

	if len(validBackups) == 0 {
		diags.AddError("No backup found", "No backup before the given date")
		return nil
	}

	sort.Slice(validBackups, func(i, j int) bool {
		return validBackups[i].CreationDate.After(validBackups[j].CreationDate)
	})

	return &validBackups[0]
}
