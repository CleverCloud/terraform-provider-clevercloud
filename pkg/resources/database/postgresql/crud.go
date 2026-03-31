package postgresql

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/resources"
	"go.clever-cloud.com/terraform-provider/pkg/resources/addon"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func (r *ResourcePostgreSQL) FetchPostgresInfos(ctx context.Context, diags *diag.Diagnostics) {
	// Skip fetching during schema validation (before provider is configured)
	if r.Provider == nil || r.Client() == nil {
		tflog.Debug(ctx, "Skipping postgres infos fetch - provider not configured yet")
		return
	}

	res := tmp.GetPostgresInfos(ctx, r.Client())
	if res.HasError() {
		tflog.Error(ctx, "failed to get postgres infos", map[string]any{"error": res.Error().Error()})
		return
	}
	r.infos = res.Payload()
	for k := range r.infos.Dedicated {
		r.dedicatedVersions = append(r.dedicatedVersions, k)
	}
}

func (r *ResourcePostgreSQL) Infos(ctx context.Context, diags *diag.Diagnostics) *tmp.PostgresInfos {
	if r.infos == nil {
		r.FetchPostgresInfos(ctx, diags)
	}

	return r.infos
}

// getLocaleFromDatabase connects to the PostgreSQL database and retrieves the LC_COLLATE setting
// which indicates the locale used when the database was created.
// Returns the locale in format "en_GB" or empty string if unable to retrieve.
// Retries connection up to 30 times with 2 second delay to handle database startup time.
func getLocaleFromDatabase(ctx context.Context, host string, port int64, database, user, password string) (string, error) {
	// Build DSN (Data Source Name) for PostgreSQL connection
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=require",
		host, port, user, password, database)

	var db *gorm.DB
	var err error

	// Retry connection up to 30 times (60 seconds total) to allow database to start
	maxRetries := 30
	for i := range maxRetries {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent), // Silent mode to avoid logs
		})
		if err == nil {
			break
		}

		if i < maxRetries-1 {
			tflog.Debug(ctx, "Database not ready, retrying...", map[string]any{
				"attempt":     i + 1,
				"max_retries": maxRetries,
				"error":       err.Error(),
			})
			time.Sleep(5 * time.Second)
		}
	}

	if err != nil {
		return "", fmt.Errorf("failed to connect to database after %d attempts: %w", maxRetries, err)
	}

	// Get the underlying *sql.DB to close it later
	sqlDB, err := db.WithContext(ctx).DB()
	if err != nil {
		return "", fmt.Errorf("failed to get database instance: %w", err)
	}
	defer func() {
		if closeErr := sqlDB.Close(); closeErr != nil {
			tflog.Warn(ctx, "failed to close database connection", map[string]any{"error": closeErr.Error()})
		}
	}()

	// Set connection timeout
	sqlDB.SetConnMaxLifetime(5 * time.Second)

	// Query LC_COLLATE setting from pg_database
	// SHOW LC_COLLATE doesn't work on all PostgreSQL versions, so we query pg_database instead
	var lcCollate string
	query := "SELECT datcollate FROM pg_database WHERE datname = current_database()"
	result := db.Raw(query).Scan(&lcCollate)
	if result.Error != nil {
		return "", fmt.Errorf("failed to query LC_COLLATE: %w", result.Error)
	}

	// LC_COLLATE returns format like "en_GB.UTF-8", we need to extract "en_GB"
	// Split by "." and take the first part
	parts := strings.Split(lcCollate, ".")
	if len(parts) > 0 && parts[0] != "" {
		return parts[0], nil
	}

	return lcCollate, nil
}

func (r *ResourcePostgreSQL) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	pg := helper.PlanFrom[PostgreSQL](ctx, req.Plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	addonsProvidersRes := tmp.GetAddonsProviders(ctx, r.Client())
	if addonsProvidersRes.HasError() {
		resp.Diagnostics.AddError("failed to get addon providers", addonsProvidersRes.Error().Error())
		return
	}
	addonsProviders := addonsProvidersRes.Payload()

	prov := pkg.LookupAddonProvider(*addonsProviders, "postgresql-addon")
	plan := pkg.LookupProviderPlan(prov, pg.Plan.ValueString())
	if plan == nil {
		resp.Diagnostics.AddError("failed to find plan", "expect: "+strings.Join(pkg.ProviderPlansAsList(prov), ", ")+", got: "+pg.Plan.String())
		return
	}

	addonReq := tmp.AddonRequest{
		Name:       pg.Name.ValueString(),
		Plan:       plan.ID,
		ProviderID: "postgresql-addon",
		Region:     pg.Region.ValueString(),
		Options:    map[string]string{},
	}

	pkg.IfIsSetStr(pg.Version, func(s string) {
		addonReq.Options["version"] = pg.Version.ValueString()
	})

	if !pg.Backup.IsNull() && !pg.Backup.IsUnknown() {
		backupValue := fmt.Sprintf("%t", pg.Backup.ValueBool())
		addonReq.Options["do-backup"] = backupValue
	} else {
		addonReq.Options["do-backup"] = "true"
	}

	if !pg.Encryption.IsNull() && !pg.Encryption.IsUnknown() {
		addonReq.Options["encryption"] = fmt.Sprintf("%t", pg.Encryption.ValueBool())
	}

	if !pg.DirectHostOnly.IsNull() && !pg.DirectHostOnly.IsUnknown() {
		addonReq.Options["direct-host-only"] = fmt.Sprintf("%t", pg.DirectHostOnly.ValueBool())
	}

	if !pg.Locale.IsNull() && !pg.Locale.IsUnknown() {
		addonReq.Options["locale"] = pg.Locale.ValueString()
	}

	res := tmp.CreateAddon(ctx, r.Client(), r.Organization(), addonReq)
	if res.HasError() {
		resp.Diagnostics.AddError("failed to create addon", res.Error().Error())
		return
	}
	createdPg := res.Payload()

	pg.ID = pkg.FromStr(createdPg.RealID)
	pg.CreationDate = pkg.FromI(createdPg.CreationDate)
	pg.Plan = pkg.FromStr(createdPg.Plan.Slug)

	resp.Diagnostics.Append(resp.State.Set(ctx, pg)...)

	pgInfoRes := tmp.GetPostgreSQL(ctx, r.Client(), createdPg.ID)
	if pgInfoRes.HasError() {
		resp.Diagnostics.AddError("failed to get postgres connection infos", pgInfoRes.Error().Error())
		return
	}
	addonPG := pgInfoRes.Payload()

	tflog.Debug(ctx, "API response", map[string]any{
		"payload": fmt.Sprintf("%+v", addonPG),
	})
	pg.Host = pkg.FromStr(addonPG.Host)
	pg.Port = pkg.FromI(int64(addonPG.Port))
	pg.Database = pkg.FromStr(addonPG.Database)
	pg.User = pkg.FromStr(addonPG.User)
	pg.Password = pkg.FromStr(addonPG.Password)
	pg.Version = pkg.FromStr(addonPG.Version)
	pg.Uri = pkg.FromStr(addonPG.Uri())

	if pg.Encryption.IsUnknown() {
		pg.Encryption = pkg.FromBool(false)
	}
	if pg.DirectHostOnly.IsUnknown() {
		pg.DirectHostOnly = pkg.FromBool(false)
	}

	for _, option := range addonPG.Features {
		switch option.Name {
		case "encryption":
			pg.Encryption = pkg.FromBool(option.Enabled)
		case "direct-host-only":
			pg.DirectHostOnly = pkg.FromBool(option.Enabled)
		}
	}

	if plan.IsDedicated() {
		locale, err := getLocaleFromDatabase(
			ctx,
			addonPG.Host,
			int64(addonPG.Port),
			addonPG.Database,
			addonPG.User,
			addonPG.Password,
		)
		if err != nil {
			resp.Diagnostics.AddAttributeError(
				path.Root("locale"),
				"Failed to retrieve locale from database",
				err.Error())
		} else {
			pg.Locale = pkg.FromStr(locale)
		}
	} else {
		pg.Locale = pkg.FromStr("en_GB")
	}

	addon.SyncNetworkGroups(
		ctx,
		r,
		createdPg.ID,
		pg.Networkgroups,
		&pg.Networkgroups,
		&resp.Diagnostics,
	)

	resp.Diagnostics.Append(resp.State.Set(ctx, pg)...)
}

func (r *ResourcePostgreSQL) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	pg := helper.StateFrom[PostgreSQL](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	if pg.ID.ValueString() == "" {
		resp.State.RemoveResource(ctx)
		return
	}

	realID, err := tmp.AddonIDToRealID(ctx, r.Client(), r.Organization(), pg.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("failed to get addon ID", err.Error())
	} else {
		pg.ID = pkg.FromStr(realID)
	}

	addonID, err := tmp.RealIDToAddonID(ctx, r.Client(), r.Organization(), pg.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("failed to get addon ID", err.Error())
		return
	}

	addonRes := tmp.GetAddon(ctx, r.Client(), r.Organization(), addonID)
	if addonRes.IsNotFoundError() {
		resp.State.RemoveResource(ctx)
		return
	} else if addonRes.HasError() {
		resp.Diagnostics.AddError("failed to get Postgres resource", addonRes.Error().Error())
	} else {
		addon := addonRes.Payload()
		pg.Name = pkg.FromStr(addon.Name)
		pg.CreationDate = pkg.FromI(addon.CreationDate)
	}

	addonPGRes := tmp.GetPostgreSQL(ctx, r.Client(), addonID)
	if addonPGRes.IsNotFoundError() {
		resp.State.RemoveResource(ctx)
		return
	} else if addonPGRes.HasError() {
		resp.Diagnostics.AddError("failed to get Postgres resource", addonPGRes.Error().Error())
	} else {
		addonPG := addonPGRes.Payload()
		if addonPG.Status == "TO_DELETE" {
			resp.State.RemoveResource(ctx)
			return
		}

		pg.Plan = pkg.FromStr(addonPG.Plan)
		pg.Region = pkg.FromStr(addonPG.Zone)
		pg.Host = pkg.FromStr(addonPG.Host)
		pg.Port = pkg.FromI(int64(addonPG.Port))
		pg.Database = pkg.FromStr(addonPG.Database)
		pg.User = pkg.FromStr(addonPG.User)
		pg.Password = pkg.FromStr(addonPG.Password)
		pg.Version = pkg.FromStr(addonPG.Version)
		pg.Uri = pkg.FromStr(addonPG.Uri())

		for _, feature := range addonPG.Features {
			switch feature.Name {
			case "do-backup":
				pg.Backup = pkg.FromBool(feature.Enabled)
			case "encryption":
				pg.Encryption = pkg.FromBool(feature.Enabled)
			case "direct-host-only":
				pg.DirectHostOnly = pkg.FromBool(feature.Enabled)
				// Note: locale feature only indicates if locale support is enabled, not the actual locale value
				// We retrieve the actual locale value by connecting to the database
			}
		}
	}

	// Retrieve the actual locale value from the database by querying LC_COLLATE
	// This is necessary because the API doesn't return the locale value
	locale, err := getLocaleFromDatabase(
		ctx,
		pg.Host.ValueString(),
		pg.Port.ValueInt64(),
		pg.Database.ValueString(),
		pg.User.ValueString(),
		pg.Password.ValueString(),
	)
	if err != nil {
		// On shared/dev plans, the user doesn't have permission to query pg_database
		// This is expected, so we just log a warning
		tflog.Warn(ctx, "Failed to retrieve locale from database, using default", map[string]any{"error": err.Error()})
	} else {
		pg.Locale = pkg.FromStr(locale)
	}

	pg.Networkgroups = resources.ReadNetworkGroups(ctx, r, addonID, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, pg)...)
}

func (r *ResourcePostgreSQL) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	plan := helper.PlanFrom[PostgreSQL](ctx, req.Plan, &resp.Diagnostics)
	state := helper.StateFrom[PostgreSQL](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.ID.ValueString() != state.ID.ValueString() { // unneeded with Identity
		resp.Diagnostics.AddError("postgresql cannot be updated", "mismatched IDs")
		return
	}

	// Only name can be edited
	if !plan.Name.Equal(state.Name) {
		addonRes := tmp.UpdateAddon(ctx, r.Client(), r.Organization(), plan.ID.ValueString(), map[string]string{
			"name": plan.Name.ValueString(),
		})
		if addonRes.HasError() {
			resp.Diagnostics.AddError("failed to update PostgreSQL", addonRes.Error().Error())
		} else {
			addon := addonRes.Payload()
			state.Name = pkg.FromStr(addon.Name)
			resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
		}
	}

	addon.SyncNetworkGroups(
		ctx,
		r,
		plan.ID.ValueString(),
		plan.Networkgroups,
		&state.Networkgroups,
		&resp.Diagnostics,
	)

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
	// Handle plan, region, or version changes via migration
	needsMigration := !plan.Plan.Equal(state.Plan) ||
		!plan.Region.Equal(state.Region) ||
		!plan.Version.Equal(state.Version)
	if needsMigration {
		r.migrate(ctx, plan, &state, &resp.Diagnostics)
		resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
	}
}

func (r *ResourcePostgreSQL) migrate(ctx context.Context, plan PostgreSQL, state *PostgreSQL, diags *diag.Diagnostics) {
	addonsProvidersRes := tmp.GetAddonsProviders(ctx, r.Client())
	if addonsProvidersRes.HasError() {
		diags.AddError("failed to get addon providers", addonsProvidersRes.Error().Error())
		return
	}
	addonsProviders := addonsProvidersRes.Payload()

	prov := pkg.LookupAddonProvider(*addonsProviders, "postgresql-addon")
	billingPlan := pkg.LookupProviderPlan(prov, plan.Plan.ValueString())
	if billingPlan == nil {
		diags.AddError("failed to find plan", "expect: "+strings.Join(pkg.ProviderPlansAsList(prov), ", ")+", got: "+plan.Plan.String())
		return
	}

	addonID, err := tmp.RealIDToAddonID(ctx, r.Client(), r.Organization(), state.ID.ValueString())
	if err != nil {
		diags.AddError("failed to get addon ID", err.Error())
		return
	}

	// Check for already running migrations
	migrationsRes := tmp.ListAddonMigrations(ctx, r.Client(), r.Organization(), addonID)
	if migrationsRes.HasError() {
		diags.AddError("failed to list migrations", migrationsRes.Error().Error())
		return
	}
	migrations := migrationsRes.Payload()

	runningMig := pkg.First(*migrations, func(mig tmp.AddonMigrationResponse) bool {
		return mig.Status == "RUNNING"
	})
	if runningMig != nil {
		diags.AddError(
			"migration already in progress",
			fmt.Sprintf("A migration (ID: %s) is already running for this addon. Please wait for it to complete before requesting a new migration.", runningMig.MigrationID),
		)
		return
	}

	// TODO:
	// migration cannot be run if instance is not ready
	// hard to know when instance is OK, because instance API return UP even if Postgres is not listening
	/*for {
		pgRes := tmp.ListInstances(ctx, r.Client(), r.Organization(), state.ID.ValueString())
		if pgRes.HasError() {
			tflog.Warn(ctx, "failed to get PG", map[string]any{"error": pgRes.Error().Error()})
			continue
		}
		instances := *pgRes.Payload()

		if len(instances) == 0 {
			continue
		}

		firstInstance := pkg.First(instances, func(instance tmp.AppInstance) bool { return true })
		state := firstInstance.State
		if state != "UP" { // BOOTING
			tflog.Info(ctx, "pg status not OK", map[string]any{"state": state})
			continue
		}

		fmt.Printf("\n\n%+v\n\n", firstInstance)
		break
	}*/

	migrationReq := tmp.AddonMigrationRequest{Region: plan.Region.ValueString(), PlanID: billingPlan.ID}
	if plan.Version.IsNull() || plan.Version.IsUnknown() {
		migrationReq.Version = state.Version.ValueStringPointer()
	} else {
		migrationReq.Version = plan.Version.ValueStringPointer()
	}
	tflog.Debug(ctx, "migration request", map[string]any{
		"version": migrationReq.Version,
		"region":  migrationReq.Region,
		"plan":    migrationReq.PlanID,
	})

	migrationRes := tmp.MigrateAddon(ctx, r.Client(), r.Organization(), addonID, migrationReq)
	if migrationRes.HasError() {
		diags.AddError("failed to migrate PostgreSQL", migrationRes.Error().Error())
		return
	}
	migration := migrationRes.Payload()

	tflog.Info(ctx, "PostgreSQL migration started", map[string]any{
		"migration_id": migration.MigrationID,
		"status":       migration.Status,
		"request_date": migration.RequestDate,
	})

	t := time.NewTicker(1 * time.Second)

	// Wait for migration to complete
	migrationID := migration.MigrationID
	for {
		// Check if context is done (timeout or cancellation)
		select {
		case <-ctx.Done():
			diags.AddError("migration timeout", "Migration did not complete within the allowed time, check DB logs")
			return
		case <-t.C:
			migrationsRes := tmp.GetAddonMigrations(ctx, r.Client(), r.Organization(), addonID, migrationID)
			if migrationsRes.HasError() {
				diags.AddWarning("failed to check migration status", migrationsRes.Error().Error())
				continue
			}
			currentMigration := migrationsRes.Payload()

			tflog.Info(ctx, "Migration status check", map[string]any{
				"addon":        state.ID.ValueString(),
				"migration_id": currentMigration.MigrationID,
				"status":       currentMigration.Status,
			})
			for _, step := range currentMigration.Steps {
				tflog.Debug(ctx, step.Name, map[string]any{
					"message": step.Message,
					"value":   step.Value,
					"status":  step.Status,
				})
			}

			if currentMigration.Status == "OK" {
				state.Plan = pkg.FromStr(billingPlan.Slug)
				state.Region = pkg.FromStr(migrationReq.Region)
				state.Version = pkg.FromStr(*migrationReq.Version)
				return
			} else if currentMigration.Status != "RUNNING" {
				diags.AddError(
					"migration failed",
					fmt.Sprintf("Migration ended with status: %s", currentMigration.Status),
				)
				return
			}
		}
	}
}

func (r *ResourcePostgreSQL) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	pg := helper.StateFrom[PostgreSQL](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	addonID, err := tmp.RealIDToAddonID(ctx, r.Client(), r.Organization(), pg.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("failed to get addon ID", err.Error())
		return
	}

	res := tmp.DeleteAddon(ctx, r.Client(), r.Organization(), addonID)
	if res.HasError() && !res.IsNotFoundError() {
		resp.Diagnostics.AddError("failed to delete addon", res.Error().Error())
	} else {
		resp.State.RemoveResource(ctx)
	}
}
