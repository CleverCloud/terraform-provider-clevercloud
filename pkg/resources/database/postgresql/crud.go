package postgresql

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/resources"
	"go.clever-cloud.com/terraform-provider/pkg/resources/addon"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
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
	r.readFromAddon(&pg, *createdPg)

	resp.Diagnostics.Append(resp.State.Set(ctx, pg)...)

	pgInfoRes := tmp.GetPostgreSQL(ctx, r.Client(), createdPg.ID)
	if pgInfoRes.HasError() {
		resp.Diagnostics.AddError("failed to get postgres connection infos", pgInfoRes.Error().Error())
		return
	}

	// readFromAPI syncs the locale from the API response; the plan value
	// (defaulted to en_GB by the schema) is kept when the API does not expose it
	r.readFromAPI(ctx, &pg, *pgInfoRes.Payload())

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
		r.readFromAddon(&pg, *addonRes.Payload())
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

		r.readFromAPI(ctx, &pg, *addonPG)
	}

	pg.Networkgroups = resources.ReadNetworkGroups(ctx, r, addonID, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, pg)...)
}

func (r *ResourcePostgreSQL) readFromAddon(state *PostgreSQL, addon tmp.AddonResponse) {
	state.Name = pkg.FromStr(addon.Name)
	state.Plan = pkg.FromStr(addon.Plan.Slug)
	state.Region = pkg.FromStr(addon.Region)
	state.CreationDate = pkg.FromI(addon.CreationDate)
}

// defaultLocale is the locale the platform uses when none is requested at creation
const defaultLocale = "en_GB"

func (r *ResourcePostgreSQL) readFromAPI(ctx context.Context, state *PostgreSQL, pg tmp.PostgreSQL) {
	state.Host = pkg.FromStr(pg.Host)
	state.Port = pkg.FromI(int64(pg.Port))
	state.Database = pkg.FromStr(pg.Database)
	state.User = pkg.FromStr(pg.User)
	state.Password = pkg.FromStr(pg.Password)
	state.Version = pkg.FromStr(pg.Version)
	state.Uri = pkg.FromStr(pg.Uri())

	// The locale is fixed at creation time (LC_COLLATE cannot change afterwards)
	// and exposed by the API. When the API does not return it, keep the value
	// already in state (config or default) and only fall back on import.
	switch {
	case pg.Locale != "":
		state.Locale = pkg.FromStr(pg.Locale)
	case state.Locale.IsNull() || state.Locale.IsUnknown():
		tflog.Warn(ctx, "API did not return the PostgreSQL locale, using default", map[string]any{"locale": defaultLocale})
		state.Locale = pkg.FromStr(defaultLocale)
	}

	// Initialize to defaults so attributes are never null in state after import.
	// The features loop below overrides with actual API values if present.
	state.Backup = pkg.FromBool(true)
	state.Encryption = pkg.FromBool(false)
	state.DirectHostOnly = pkg.FromBool(false)

	for _, feature := range pg.Features {
		switch feature.Name {
		case "do-backup":
			state.Backup = pkg.FromBool(feature.Enabled)
		case "encryption":
			state.Encryption = pkg.FromBool(feature.Enabled)
		case "direct-host-only":
			state.DirectHostOnly = pkg.FromBool(feature.Enabled)
		}
	}
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

				// A migration provisions a new instance: connection details change.
				// Refresh them explicitly since they are pinned via UseStateForUnknown.
				if pgRes := tmp.GetPostgreSQL(ctx, r.Client(), addonID); pgRes.HasError() {
					diags.AddWarning("failed to refresh connection info after migration", pgRes.Error().Error())
				} else {
					addonPG := pgRes.Payload()
					state.Host = pkg.FromStr(addonPG.Host)
					state.Port = pkg.FromI(int64(addonPG.Port))
					state.Database = pkg.FromStr(addonPG.Database)
					state.User = pkg.FromStr(addonPG.User)
					state.Password = pkg.FromStr(addonPG.Password)
					state.Version = pkg.FromStr(addonPG.Version)
					state.Uri = pkg.FromStr(addonPG.Uri())
				}
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
