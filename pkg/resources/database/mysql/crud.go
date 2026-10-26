package mysql

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/addonid"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/resources"
	"go.clever-cloud.com/terraform-provider/pkg/resources/addon"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
	"go.clever-cloud.dev/sdk/models"
	"go.clever-cloud.dev/sdk/services/addons"
)

func (r *ResourceMySQL) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.FetchMysqlInfos(ctx, &resp.Diagnostics)
	r.Configurer.Configure(ctx, req, resp)
}

func (r *ResourceMySQL) FetchMysqlInfos(ctx context.Context, diags *diag.Diagnostics) {
	// Skip fetching during schema validation (before provider is configured)
	if r.Provider == nil || r.Client() == nil {
		tflog.Debug(ctx, "Skipping mysql infos fetch - provider not configured yet")
		return
	}

	res := tmp.GetMysqlInfos(ctx, r.Client())
	if res.HasError() {
		tflog.Error(ctx, "failed to get mysql infos", map[string]any{"error": res.Error().Error()})
		return
	}
	r.infos = res.Payload()
	for k := range r.infos.Dedicated {
		r.dedicatedVersions = append(r.dedicatedVersions, k)
	}
}

func (r *ResourceMySQL) Infos(ctx context.Context, diags *diag.Diagnostics) *tmp.ProvidersConsoleView {
	if r.infos == nil {
		r.FetchMysqlInfos(ctx, diags)
	}

	return r.infos
}

// Create a new resource
func (r *ResourceMySQL) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	my := helper.PlanFrom[MySQL](ctx, req.Plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	addonsProvidersRes := tmp.GetAddonsProviders(ctx, r.Client())
	if addonsProvidersRes.HasError() {
		resp.Diagnostics.AddError("failed to get addon providers", addonsProvidersRes.Error().Error())
		return
	}
	addonsProviders := addonsProvidersRes.Payload()

	prov := pkg.LookupAddonProvider(*addonsProviders, "mysql-addon")
	plan := pkg.LookupProviderPlan(prov, my.Plan.ValueString())
	if plan == nil {
		resp.Diagnostics.AddError("failed to find plan", "expect: "+strings.Join(pkg.ProviderPlansAsList(prov), ", ")+", got: "+my.Plan.String())
		return
	}

	addonReq := models.WannabeAddonProvision{
		Name:       my.Name.ValueString(),
		Plan:       plan.ID,
		ProviderID: "mysql-addon",
		Region:     my.Region.ValueString(),
		Options:    map[string]string{},
	}

	if !my.Version.IsNull() && !my.Version.IsUnknown() {
		addonReq.Options["version"] = my.Version.ValueString()
	}

	addonReq.Options["do-backup"] = "true"
	if !my.Backup.IsNull() && !my.Backup.IsUnknown() {
		addonReq.Options["do-backup"] = fmt.Sprintf("%t", my.Backup.ValueBool())
	}

	if !my.Encryption.IsNull() && !my.Encryption.IsUnknown() {
		addonReq.Options["encryption"] = fmt.Sprintf("%t", my.Encryption.ValueBool())
	}

	if !my.DirectHostOnly.IsNull() && !my.DirectHostOnly.IsUnknown() {
		addonReq.Options["direct-host-only"] = fmt.Sprintf("%t", my.DirectHostOnly.ValueBool())
	}

	if !my.SkipLogBin.IsNull() && !my.SkipLogBin.IsUnknown() {
		addonReq.Options["skip-log-bin"] = fmt.Sprintf("%t", my.SkipLogBin.ValueBool())
	}

	res := resources.CreateAddon(ctx, r.Client(), r.Organization(), addonReq)
	if res.HasError() {
		resp.Diagnostics.AddError("failed to create addon", res.Error().Error())
		return
	}
	createdMy := res.Payload()

	my.ID = pkg.FromStr(pkg.Deref(createdMy.RealID))
	r.readFromAddon(&my, *createdMy)

	resp.Diagnostics.Append(resp.State.Set(ctx, my)...)

	myInfoRes := tmp.GetMySQL(ctx, r.Client(), createdMy.ID)
	if myInfoRes.HasError() {
		resp.Diagnostics.AddError("failed to get mysql connection infos", myInfoRes.Error().Error())
		return
	} else {
		r.readFromAPI(&my, *myInfoRes.Payload())
	}

	addon.SyncNetworkGroups(
		ctx,
		r,
		createdMy.ID,
		my.Networkgroups,
		&my.Networkgroups,
		&resp.Diagnostics,
	)
	my.Tags = resources.SyncAddonTags(ctx, r, createdMy.ID, my.Tags, resources.NullTags, &resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, my)...)
}

// Read resource information
func (r *ResourceMySQL) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Debug(ctx, "Mysql READ", map[string]any{"request": req})

	// State
	my := helper.StateFrom[MySQL](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	if my.ID.ValueString() == "" {
		resp.State.RemoveResource(ctx)
		return
	}

	// IDs
	addonId, err := addonid.FromRealID(ctx, r.Client(), r.Organization(), my.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("failed to get addon ID", err.Error())
		return
	}

	// Objects
	addonRes := r.SDK.V2().Organisations().ID(r.Organization()).Addons().Addonid(addonId).Getaddonbyorgaandaddonid(ctx)
	if addonRes.IsNotFoundError() {
		resp.State.RemoveResource(ctx)
		return
	}
	if addonRes.HasError() {
		resp.Diagnostics.AddError("failed to get Mysql resource", addonRes.Error().Error())
		return
	} else {
		addonInfo := addonRes.Payload()
		r.readFromAddon(&my, *addonInfo)
	}

	addonMyRes := tmp.GetMySQL(ctx, r.Client(), addonId)
	if addonMyRes.IsNotFoundError() {
		resp.State.RemoveResource(ctx)
		return
	}
	if addonMyRes.HasError() {
		resp.Diagnostics.AddError("failed to get Mysql resource", addonMyRes.Error().Error())
		return
	} else {
		addonMy := addonMyRes.Payload()
		if pkg.Deref(addonMy.Status) == "TO_DELETE" {
			resp.State.RemoveResource(ctx)
			return
		}

		r.readFromAPI(&my, *addonMy)
	}

	my.Networkgroups = resources.ReadNetworkGroups(ctx, r, addonId, &resp.Diagnostics)
	my.Tags = resources.ReadAddonTags(ctx, r, addonId, my.Tags, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, my)...)
}

func (r *ResourceMySQL) readFromAddon(state *MySQL, addon models.AddonView) {
	state.Name = pkg.FromStr(pkg.Deref(addon.Name))
	state.Plan = pkg.FromStr(pkg.Deref(addon.Plan.Slug))
	state.Region = pkg.FromStr(addon.Region)
	state.CreationDate = pkg.FromI(pkg.Deref(addon.CreationDate))
}

func (r *ResourceMySQL) readFromAPI(state *MySQL, my tmp.MysqlAddonView) {
	state.Host = pkg.FromStr(my.Host)
	state.Port = pkg.FromI(int64(my.Port))
	state.Database = pkg.FromStr(my.Database)
	state.User = pkg.FromStr(my.User)
	state.Password = pkg.FromStr(my.Password)
	state.Version = pkg.FromStr(my.Version)
	state.Uri = pkg.FromStr(uri(my))
	state.ReadOnlyUsers = fromReadOnlyUsers(my.ReadOnlyUsers)

	// Initialize to defaults so attributes are never null in state after import.
	// The features loop below overrides with actual API values if present.
	state.Backup = pkg.FromBool(true)
	state.Encryption = pkg.FromBool(false)
	state.DirectHostOnly = pkg.FromBool(false)
	state.SkipLogBin = pkg.FromBool(false)

	for _, feature := range my.Features {
		switch feature.Name {
		case "do-backup":
			state.Backup = pkg.FromBool(pkg.Deref(feature.Enabled))
		case "encryption":
			state.Encryption = pkg.FromBool(pkg.Deref(feature.Enabled))
		case "direct-host-only":
			state.DirectHostOnly = pkg.FromBool(pkg.Deref(feature.Enabled))
		case "skip-log-bin":
			state.SkipLogBin = pkg.FromBool(pkg.Deref(feature.Enabled))
		}
	}
}

// Update resource
func (r *ResourceMySQL) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	plan := helper.PlanFrom[MySQL](ctx, req.Plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	state := helper.StateFrom[MySQL](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.ID.ValueString() != state.ID.ValueString() {
		resp.Diagnostics.AddError("mysql cannot be updated", "mismatched IDs")
		return
	}

	// Only name can be edited
	addonRes := r.SDK.V2().Organisations().ID(r.Organization()).Addons().Addonid(plan.ID.ValueString()).Updateaddoninfo(ctx, addons.UpdateaddoninfoBody{
		Name: plan.Name.ValueString(),
	})
	if addonRes.HasError() {
		resp.Diagnostics.AddError("failed to update Mysql", addonRes.Error().Error())
		return
	}
	state.Name = pkg.FromStr(pkg.Deref(addonRes.Payload().Name))

	addon.SyncNetworkGroups(
		ctx,
		r,
		plan.ID.ValueString(),
		plan.Networkgroups,
		&state.Networkgroups,
		&resp.Diagnostics,
	)
	if !plan.Tags.Equal(state.Tags) {
		state.Tags = resources.SyncAddonTags(ctx, r, plan.ID.ValueString(), plan.Tags, state.Tags, &resp.Diagnostics)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Delete resource
func (r *ResourceMySQL) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	my := helper.StateFrom[MySQL](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "MySQL DELETE", map[string]any{"my": my})

	addonId, err := addonid.FromRealID(ctx, r.Client(), r.Organization(), my.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("failed to get addon ID", err.Error())
		return
	}

	res := r.SDK.V2().Organisations().ID(r.Organization()).Addons().Addonid(addonId).Deprovisionaddonbyorgaandaddonid(ctx)
	if res.IsNotFoundError() {
		resp.State.RemoveResource(ctx)
		return
	}
	if res.HasError() {
		resp.Diagnostics.AddError("failed to delete addon", res.Error().Error())
		return
	}

	resp.State.RemoveResource(ctx)
}

// uri is the connection URI of a MySQL add-on.
func uri(my tmp.MysqlAddonView) string {
	return fmt.Sprintf("mysql://%s:%s@%s:%d/%s", my.User, my.Password, my.Host, my.Port, my.Database)
}

// fromReadOnlyUsers converts the read-only users of a MySQL add-on to a types.List with nested objects
func fromReadOnlyUsers(users []tmp.ReadOnlyUserView) types.List {
	objectType := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"user":     types.StringType,
			"password": types.StringType,
		},
	}

	if len(users) == 0 {
		return types.ListNull(objectType)
	}

	objects := make([]attr.Value, len(users))
	for i, user := range users {
		objects[i] = types.ObjectValueMust(objectType.AttrTypes, map[string]attr.Value{
			"user":     types.StringValue(user.User),
			"password": types.StringValue(user.Password),
		})
	}

	return types.ListValueMust(objectType, objects)
}
