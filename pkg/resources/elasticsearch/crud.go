package elasticsearch

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
	"go.clever-cloud.dev/client"
)

func (r *ResourceElasticsearch) Configure(ctx context.Context, req resource.ConfigureRequest, res *resource.ConfigureResponse) {
	r.fetchElasticsearchInfos(ctx, &res.Diagnostics)
	r.Configurer.Configure(ctx, req, res)
}

func (r *ResourceElasticsearch) fetchElasticsearchInfos(ctx context.Context, diags *diag.Diagnostics) {
	cc := client.New()

	res := tmp.GetElasticsearchInfos(ctx, cc)
	if res.HasError() {
		diags.AddError("failed to get elasticsearch infos", res.Error().Error())
		return
	}

	for k := range res.Payload().Dedicated {
		r.versions.Add(k)
	}
}

func (r *ResourceElasticsearch) Create(ctx context.Context, req resource.CreateRequest, res *resource.CreateResponse) {
	identity := ElasticsearchIdentity{}
	plan := helper.PlanFrom[Elasticsearch](ctx, req.Plan, &res.Diagnostics)
	if res.Diagnostics.HasError() {
		return
	}

	addonsProvidersRes := tmp.GetAddonsProviders(ctx, r.Client())
	if addonsProvidersRes.HasError() {
		res.Diagnostics.AddError("failed to get addon providers", addonsProvidersRes.Error().Error())
		return
	}
	addonsProviders := addonsProvidersRes.Payload()

	prov := pkg.LookupAddonProvider(*addonsProviders, "es-addon")
	providerPlan := pkg.LookupProviderPlan(prov, plan.Plan.ValueString())
	if providerPlan == nil {
		res.Diagnostics.AddError("failed to find plan", "expect: "+strings.Join(pkg.ProviderPlansAsList(prov), ", ")+", got: "+plan.Plan.String())
		return
	}

	addonReq := tmp.AddonRequest{
		Name:       plan.Name.ValueString(),
		Plan:       providerPlan.ID,
		ProviderID: "es-addon",
		Region:     plan.Region.ValueString(),
		Options:    map[string]string{},
	}

	if !plan.Version.IsNull() && !plan.Version.IsUnknown() {
		addonReq.Options["version"] = plan.Version.ValueString()
	}

	if plan.Encryption.ValueBool() {
		addonReq.Options["encryption"] = "true"
	}

	if plan.Kibana.ValueBool() {
		addonReq.Options["kibana"] = "true"
	}

	if plan.Apm.ValueBool() {
		addonReq.Options["apm"] = "true"
	}

	plugins := pkg.SetTo[string](ctx, plan.Plugins, &res.Diagnostics)
	if len(plugins) > 0 {
		addonReq.Options["plugins"] = strings.Join(plugins, ",")
	}

	addonRes := tmp.CreateAddon(ctx, r.Client(), r.Organization(), addonReq)
	if addonRes.HasError() {
		res.Diagnostics.AddError("failed to create addon", addonRes.Error().Error())
		return
	}
	addon := addonRes.Payload()

	identity.ID = pkg.FromStr(addon.RealID)
	res.Diagnostics.Append(res.Identity.Set(ctx, identity)...)

	r.readFromAddon(&plan, *addon)

	envRes := tmp.GetAddonEnv(ctx, r.Client(), r.Organization(), addon.RealID)
	if envRes.HasError() {
		res.Diagnostics.AddError("failed to get addon env", envRes.Error().Error())
	} else {
		r.readFromEnv(&plan, *envRes.Payload())
	}

	esRes := tmp.GetElasticsearch(ctx, r.Client(), addon.ID)
	if esRes.HasError() {
		res.Diagnostics.AddError("failed to get Elasticsearch", esRes.Error().Error())
	} else {
		r.readFromAPI(&plan, *esRes.Payload())
	}

	res.Diagnostics.Append(res.State.Set(ctx, plan)...)
}

func (r *ResourceElasticsearch) Read(ctx context.Context, req resource.ReadRequest, res *resource.ReadResponse) {
	state := helper.StateFrom[Elasticsearch](ctx, req.State, &res.Diagnostics)
	identity := helper.IdentityFrom[ElasticsearchIdentity](ctx, *req.Identity, &res.Diagnostics)
	if res.Diagnostics.HasError() {
		return
	}

	addonRes := tmp.GetAddon(ctx, r.Client(), r.Organization(), identity.ID.ValueString())
	if addonRes.IsNotFoundError() {
		res.State.RemoveResource(ctx)
		return
	}
	if addonRes.HasError() {
		res.Diagnostics.AddError("failed to get addon", addonRes.Error().Error())
	} else {
		r.readFromAddon(&state, *addonRes.Payload())
	}

	addonId, err := tmp.RealIDToAddonID(ctx, r.Client(), r.Organization(), identity.ID.ValueString())
	if err != nil {
		res.Diagnostics.AddError("failed to get addon ID", err.Error())
	} else {
		elasticRes := tmp.GetElasticsearch(ctx, r.Client(), addonId)
		if elasticRes.HasError() {
			res.Diagnostics.AddError("failed to get Elasticsearch resource", elasticRes.Error().Error())
		} else {
			r.readFromAPI(&state, *elasticRes.Payload())
		}
	}

	res.Diagnostics.Append(res.State.Set(ctx, state)...)
}

func (r *ResourceElasticsearch) readFromAPI(state *Elasticsearch, elastic tmp.Elasticsearch) {
	state.Host = pkg.FromStr(elastic.Host)
	state.User = pkg.FromStr(elastic.User)
	state.Password = pkg.FromStr(elastic.Password)
	state.Version = pkg.FromStr(elastic.Version)

	features := pkg.Reduce(
		elastic.Features,
		map[string]bool{},
		func(acc map[string]bool, feature tmp.ElasticsearchFeature) map[string]bool {
			acc[feature.Name] = feature.Enabled
			return acc
		})

	state.KibanaUser = basetypes.NewStringNull()
	state.KibanaPassword = basetypes.NewStringNull()
	if features["kibana"] {
		state.KibanaUser = pkg.FromStr(elastic.KibanaUser)
		state.KibanaPassword = pkg.FromStr(elastic.KibanaPassword)
	}

	state.ApmUser = basetypes.NewStringNull()
	state.ApmPassword = basetypes.NewStringNull()
	state.ApmToken = basetypes.NewStringNull()
	if features["apm"] {
		state.ApmUser = pkg.FromStr(elastic.ApmUser)
		state.ApmPassword = pkg.FromStr(elastic.ApmPassword)
		state.ApmToken = pkg.FromStr(elastic.ApmAuthToken)
	}

	state.Encryption = pkg.FromBool(features["encryption"])
	state.Plugins = basetypes.NewSetNull(types.StringType)
}

func (r *ResourceElasticsearch) readFromAddon(state *Elasticsearch, addon tmp.AddonResponse) {
	state.Plan = pkg.FromStr(addon.Plan.Slug)
	state.Name = pkg.FromStr(addon.Name)
	state.Region = pkg.FromStr(addon.Region)
}

func (r *ResourceElasticsearch) readFromEnv(state *Elasticsearch, env tmp.EnvVars) {
	m := env.Map()

	state.KibanaHost = basetypes.NewStringNull()
	if kibanaHost := m["ES_ADDON_KIBANA_HOST"]; kibanaHost != "" {
		state.KibanaHost = pkg.FromStr(kibanaHost)
	}

	state.ApmHost = basetypes.NewStringNull()
	if apmHost := m["ES_ADDON_APM_HOST"]; apmHost != "" {
		state.ApmHost = pkg.FromStr(apmHost)
	}
}

func (r *ResourceElasticsearch) Update(ctx context.Context, req resource.UpdateRequest, res *resource.UpdateResponse) {
	// TODO
	// during upgrade, give new plan, new version, new zone
	// 8 -> 7 = Version changes are not supported
	// POST https://api.clever-cloud.com/v2/organisations/user_xx/addons/addon_e1b59330-3b1b-4739-b23c-31af436df2e7/migrations
	// {"planId":"plan_7675a239-057e-448e-85fb-77b5aa2ef47e","region":"par","version":"8"}
	// {"migrationId":"migration_4f84a9f6-e9bd-4d09-b60d-9210e25795cb","requestDate":"2025-10-29T16:25:52.525017Z","steps":[],"status":"RUNNING"}

	// GET https: //api.clever-cloud.com/v2/organisations/user_xxx/addons/addon_e1b59330-3b1b-4739-b23c-31af436df2e7/migrations/migration_4f84a9f6-e9bd-4d09-b60d-9210e25795cb
	// {
	// 	"migrationId":"migration_4f84a9f6-e9bd-4d09-b60d-9210e25795cb",
	// 	"requestDate":"2025-10-29T16:25:52.525Z",
	// 	"steps":[
	// 		{"value":"RETRIEVE_ADDON","status":"OK","startDate":"2025-10-29T16:25:52.553912Z","endDate":"2025-10-29T16:25:52.5939Z"},
	// 		{"value":"CHECK_NO_MIGRATION_ALREADY_RUNNING_FOR_ADDON","status":"OK","startDate":"2025-10-29T16:25:52.597531Z","endDate":"2025-10-29T16:25:52.616091Z"},
	// 		{"value":"ASK_MIGRATION_INSTANCE_BOOT","status":"OK","startDate":"2025-10-29T16:25:53.091656Z","endDate":"2025-10-29T16:25:53.165511Z"},
	// 		{"value":"QUEUE_MIGRATION_INSTANCE_BOOT","status":"OK","startDate":"2025-10-29T16:25:53.168737Z","endDate":"2025-10-29T16:25:53.781705Z"},
	// 		{"value":"DEPLOY_MIGRATION_INSTANCE","status":"RUNNING","startDate":"2025-10-29T16:25:53.78642Z"}
	// ],"status":"RUNNING"}
	// status	"OK"

	identity := helper.IdentityFrom[ElasticsearchIdentity](ctx, *req.Identity, &res.Diagnostics)
	plan := helper.PlanFrom[Elasticsearch](ctx, req.Plan, &res.Diagnostics)
	state := helper.StateFrom[Elasticsearch](ctx, req.State, &res.Diagnostics)
	if res.Diagnostics.HasError() {
		return
	}

	// Only name can be edited
	addonRes := tmp.UpdateAddon(ctx, r.Client(), r.Organization(), identity.ID.ValueString(), map[string]string{
		"name": plan.Name.ValueString(),
	})
	if addonRes.HasError() {
		res.Diagnostics.AddError("failed to update Cellar", addonRes.Error().Error())
	} else {
		state.Name = plan.Name
	}

	res.Diagnostics.Append(res.State.Set(ctx, state)...)
}

func (r *ResourceElasticsearch) Delete(ctx context.Context, req resource.DeleteRequest, res *resource.DeleteResponse) {
	identity := helper.IdentityFrom[ElasticsearchIdentity](ctx, *req.Identity, &res.Diagnostics)

	deleteRes := tmp.DeleteAddon(ctx, r.Client(), r.Organization(), identity.ID.ValueString())
	if deleteRes.HasError() && !deleteRes.IsNotFoundError() {
		res.Diagnostics.AddError("failed to delete addon", deleteRes.Error().Error())
		return
	}

	res.State.RemoveResource(ctx)
}
