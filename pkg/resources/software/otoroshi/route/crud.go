package route

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
)

// otoroshiBootTimeout bounds how long we wait for a freshly created add-on to
// serve its admin API
const otoroshiBootTimeout = 5 * time.Minute

func (r *ResourceOtoroshiRoute) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	plan := helper.PlanFrom[Route](ctx, req.Plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	oto := r.otoroshiContext(ctx, plan.OtoroshiID.ValueString(), &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := oto.client.WaitReady(ctx, otoroshiBootTimeout); err != nil {
		resp.Diagnostics.AddError("Otoroshi is not ready to accept routes", err.Error())
		return
	}

	route := oto.routeFrom(plan, "")

	created, err := oto.client.CreateRoute(ctx, route)
	if err != nil {
		resp.Diagnostics.AddError("failed to create Otoroshi route", err.Error())
		return
	}

	plan.ID = pkg.FromStr(created.ID)
	applyComputed(&plan, route)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *ResourceOtoroshiRoute) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	state := helper.StateFrom[Route](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	oto := r.otoroshiContext(ctx, state.OtoroshiID.ValueString(), &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	route, found, err := oto.client.GetRoute(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("failed to get Otoroshi route", err.Error())
		return
	}
	if !found {
		resp.State.RemoveResource(ctx)
		return
	}

	state.Name = pkg.FromStr(route.Name)
	state.Enabled = pkg.FromBool(route.Enabled)

	if len(route.Frontend.Domains) > 0 {
		domain, path := splitFrontendDomain(route.Frontend.Domains[0])
		state.Domain = pkg.FromStr(domain)
		state.Path = pkg.FromStr(path)
		state.URL = pkg.FromStr(routeURL(domain, path))
	}
	state.StripPath = pkg.FromBool(route.Frontend.StripPath)

	if len(route.Backend.Targets) > 0 {
		state.BackendHostname = pkg.FromStr(route.Backend.Targets[0].Hostname)
		state.Port = pkg.FromI(route.Backend.Targets[0].Port)
		state.ApplicationID = pkg.FromStr(applicationIDFromHostname(route.Backend.Targets[0].Hostname))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *ResourceOtoroshiRoute) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	plan := helper.PlanFrom[Route](ctx, req.Plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	state := helper.StateFrom[Route](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	oto := r.otoroshiContext(ctx, plan.OtoroshiID.ValueString(), &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Otoroshi replaces the whole entity on PUT, the ID comes from the current state
	// since it is never part of the plan
	route := oto.routeFrom(plan, state.ID.ValueString())

	if err := oto.client.UpdateRoute(ctx, route); err != nil {
		resp.Diagnostics.AddError("failed to update Otoroshi route", err.Error())
		return
	}

	plan.ID = state.ID
	applyComputed(&plan, route)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *ResourceOtoroshiRoute) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	state := helper.StateFrom[Route](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	oto := r.otoroshiContext(ctx, state.OtoroshiID.ValueString(), &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := oto.client.DeleteRoute(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("failed to delete Otoroshi route", err.Error())
		return
	}

	resp.State.RemoveResource(ctx)
}

// routeFrom builds the Otoroshi payload a plan describes. The ID is empty on
// create, so that Otoroshi generates one
func (oto *otoroshiContext) routeFrom(plan Route, id string) ngRoute {
	domain := plan.Domain.ValueString()
	if domain == "" {
		domain = oto.defaultDomain
	}

	return ngRoute{
		ID:      id,
		Name:    plan.Name.ValueString(),
		Enabled: plan.Enabled.ValueBool(),
		Frontend: ngFrontend{
			Domains:   []string{frontendDomain(domain, plan.Path.ValueString())},
			StripPath: plan.StripPath.ValueBool(),
		},
		Backend: ngBackend{Targets: []ngTarget{{
			Hostname: backendHostname(plan.ApplicationID.ValueString(), oto.networkgroup),
			Port:     plan.Port.ValueInt64(),
		}}},
	}
}

// applyComputed fills the attributes the provider resolves rather than reads back,
// so that state matches what was actually sent to Otoroshi
func applyComputed(plan *Route, route ngRoute) {
	domain, path := splitFrontendDomain(route.Frontend.Domains[0])

	plan.Domain = pkg.FromStr(domain)
	plan.BackendHostname = pkg.FromStr(route.Backend.Targets[0].Hostname)
	plan.URL = pkg.FromStr(routeURL(domain, path))
}

func routeURL(domain, path string) string {
	return fmt.Sprintf("https://%s%s", domain, path)
}
