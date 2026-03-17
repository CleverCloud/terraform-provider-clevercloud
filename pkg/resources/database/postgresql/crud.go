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
	tflog.Debug(ctx, "ResourcePostgreSQL.Create()")

	plan := helper.PlanFrom[PostgreSQL](ctx, req.Plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	addonID, createDiags := addon.Create(ctx, r, &plan)
	resp.Diagnostics.Append(createDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	addon.SyncNetworkGroups(ctx, r, addonID, plan.Networkgroups, &plan.Networkgroups, &resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *ResourcePostgreSQL) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Debug(ctx, "ResourcePostgreSQL.Read()")

	state := helper.StateFrom[PostgreSQL](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	addonIsDeleted, diags := addon.Read(ctx, r, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if addonIsDeleted {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *ResourcePostgreSQL) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	tflog.Debug(ctx, "ResourcePostgreSQL.Update()")

	plan := helper.PlanFrom[PostgreSQL](ctx, req.Plan, &resp.Diagnostics)
	state := helper.StateFrom[PostgreSQL](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	addonID, updateDiags := addon.Update(ctx, r, &plan, &state)
	resp.Diagnostics.Append(updateDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.SetFromResponse(ctx, r.Client(), r.Organization(), addonID, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save plan to state (plan pattern)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	addon.SyncNetworkGroups(ctx, r, addonID, plan.Networkgroups, &plan.Networkgroups, &resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)

	// Handle plan, region, or version changes via migration
	needsMigration := !plan.Plan.Equal(state.Plan) ||
		!plan.Region.Equal(state.Region) ||
		!plan.Version.Equal(state.Version)
	if needsMigration {
		r.migrate(ctx, plan, &plan, &resp.Diagnostics)
		resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
	}
}

func (r *ResourcePostgreSQL) migrate(ctx context.Context, plan PostgreSQL, target *PostgreSQL, diags *diag.Diagnostics) {
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

	addonID, err := tmp.RealIDToAddonID(ctx, r.Client(), r.Organization(), target.ID.ValueString())
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

	migrationReq := tmp.AddonMigrationRequest{Region: plan.Region.ValueString(), PlanID: billingPlan.ID}
	if plan.Version.IsNull() || plan.Version.IsUnknown() {
		migrationReq.Version = target.Version.ValueStringPointer()
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
				"addon":        target.ID.ValueString(),
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
				target.Plan = pkg.FromStr(billingPlan.Slug)
				target.Region = pkg.FromStr(migrationReq.Region)
				target.Version = pkg.FromStr(*migrationReq.Version)
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
	tflog.Debug(ctx, "ResourcePostgreSQL.Delete()")

	state := helper.StateFrom[PostgreSQL](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(addon.Delete(ctx, r, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.State.RemoveResource(ctx)
}
