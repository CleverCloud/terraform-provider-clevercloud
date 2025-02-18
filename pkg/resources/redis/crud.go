package redis

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/provider"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

// Weird behaviour, but TF can ask for a Resource without having configured a Provider (maybe for Meta and Schema)
// So we need to handle the case there is no ProviderData
func (r *ResourceRedis) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	tflog.Debug(ctx, "ResourceRedis.Configure()")

	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	provider, ok := req.ProviderData.(provider.Provider)
	if ok {
		r.cc = provider.Client()
		r.org = provider.Organization()
	}
}

// Create a new resource
func (r *ResourceRedis) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	kv := Redis{}

	resp.Diagnostics.Append(req.Plan.Get(ctx, &kv)...)
	if resp.Diagnostics.HasError() {
		return
	}

	addonsProvidersRes := tmp.GetAddonsProviders(ctx, r.cc)
	if addonsProvidersRes.HasError() {
		resp.Diagnostics.AddError("failed to get addon providers", addonsProvidersRes.Error().Error())
		return
	}

	addonsProviders := addonsProvidersRes.Payload()
	provider := pkg.LookupAddonProvider(*addonsProviders, "redis-addon")
	plan := pkg.LookupProviderPlan(provider, kv.Plan.ValueString())
	if plan == nil {
		resp.Diagnostics.AddError("This plan does not exists", "available plans are: "+strings.Join(pkg.ProviderPlansAsList(provider), ", "))
		return
	}

	addonReq := tmp.AddonRequest{
		Name:       kv.Name.ValueString(),
		Plan:       plan.ID,
		ProviderID: "redis-addon",
		Region:     kv.Region.ValueString(),
	}

	res := tmp.CreateAddon(ctx, r.cc, r.org, addonReq)
	if res.HasError() {
		resp.Diagnostics.AddError("failed to create addon", res.Error().Error())
		return
	}

	kv.ID = pkg.FromStr(res.Payload().RealID)
	kv.CreationDate = pkg.FromI(res.Payload().CreationDate)

	resp.Diagnostics.Append(resp.State.Set(ctx, kv)...)
	if resp.Diagnostics.HasError() {
		return
	}

	envRes := tmp.GetAddonEnv(ctx, r.cc, r.org, kv.ID.ValueString())
	if envRes.HasError() {
		resp.Diagnostics.AddError("failed to get Redis connection infos", envRes.Error().Error())
		return
	}

	env := *envRes.Payload()
	envAsMap := pkg.Reduce(env, map[string]types.String{}, func(acc map[string]types.String, v tmp.EnvVar) map[string]types.String {
		acc[v.Name] = pkg.FromStr(v.Value)
		return acc
	})
	tflog.Info(ctx, "API response", map[string]interface{}{
		"payload": fmt.Sprintf("%+v", envAsMap),
	})
	port, err := strconv.ParseInt(envAsMap["REDIS_PORT"].ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("invalid port received", "expect REDIS_PORT to be an Integer")
	}
	kv.Host = envAsMap["REDIS_HOST"]
	kv.Port = pkg.FromI(port)
	kv.Token = envAsMap["REDIS_PASSWORD"]

	resp.Diagnostics.Append(resp.State.Set(ctx, kv)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read resource information
func (r *ResourceRedis) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Debug(ctx, "Redis READ", map[string]interface{}{"request": req})

	var kv Redis
	diags := req.State.Get(ctx, &kv)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	addonKVRes := tmp.GetAddon(ctx, r.cc, r.org, kv.ID.ValueString())
	if addonKVRes.IsNotFoundError() {
		diags = resp.State.SetAttribute(ctx, path.Root("id"), types.StringUnknown())
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
	if addonKVRes.IsNotFoundError() {
		resp.State.RemoveResource(ctx)
		return
	}
	if addonKVRes.HasError() {
		resp.Diagnostics.AddError("failed to get Redis resource", addonKVRes.Error().Error())
	}

	addonKV := addonKVRes.Payload()
	tflog.Debug(ctx, "redis", map[string]interface{}{"payload": fmt.Sprintf("%+v", addonKV)})

	envRes := tmp.GetAddonEnv(ctx, r.cc, r.org, kv.ID.ValueString())
	if envRes.HasError() {
		resp.Diagnostics.AddError("failed to get Redis connection infos", envRes.Error().Error())
		return
	}

	env := *envRes.Payload()
	envAsMap := pkg.Reduce(env, map[string]types.String{}, func(acc map[string]types.String, v tmp.EnvVar) map[string]types.String {
		acc[v.Name] = pkg.FromStr(v.Value)
		return acc
	})
	tflog.Info(ctx, "API response", map[string]interface{}{
		"payload": fmt.Sprintf("%+v", envAsMap),
	})
	port, err := strconv.ParseInt(envAsMap["REDIS_PORT"].ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("invalid port received", "expect REDIS_PORT to be an Integer")
	}

	kv.Name = pkg.FromStr(addonKV.Name)
	kv.Host = envAsMap["REDIS_HOST"]
	kv.Plan = pkg.FromStr(addonKV.Plan.Slug)
	kv.Port = pkg.FromI(port)
	kv.Region = pkg.FromStr(addonKV.Region)
	kv.Token = envAsMap["REDIS_PASSWORD"]

	diags = resp.State.Set(ctx, kv)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update resource
func (r *ResourceRedis) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// TODO
}

// Delete resource
func (r *ResourceRedis) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	kv := Redis{}

	diags := req.State.Get(ctx, &kv)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Redis DELETE", map[string]interface{}{"kv": kv})

	res := tmp.DeleteAddon(ctx, r.cc, r.org, kv.ID.ValueString())
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

// Import resource
func (r *ResourceRedis) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Save the import identifier in the id attribute
	// and call Read() to fill fields
	attr := path.Root("id")
	resource.ImportStatePassthroughID(ctx, attr, req, resp)
}
