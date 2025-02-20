package resources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
	"go.clever-cloud.dev/client"
)

// Helper to manage states and providers plans
type Controller[STATE any] struct {
	provider *tmp.AddonProvider
}

func (c *Controller[STATE]) GetState(ctx context.Context, plan StateOrPlan, diags diag.Diagnostics) *STATE {
	state := new(STATE)
	diags.Append(plan.Get(ctx, state)...)

	return state
}

func (c *Controller[STATE]) SetState(ctx context.Context, plan StateOrPlan, cstate STATE, diags diag.Diagnostics) {
	diags.Append(plan.Set(ctx, cstate)...)
}

type StateOrPlan interface {
	Get(ctx context.Context, target interface{}) diag.Diagnostics
	Set(ctx context.Context, val interface{}) diag.Diagnostics
}

// Load informations about plans for a provider
func (c *Controller[STATE]) Init(ctx context.Context, client *client.Client, orga, providerId string, diags diag.Diagnostics) {
	providersRes := tmp.GetAddonsProviders(ctx, client)
	if providersRes.HasError() {
		diags.AddError("failed to get addon providers", providersRes.Error().Error())
		return
	}

	p := pkg.LookupAddonProvider(*providersRes.Payload(), providerId)
	if p == nil {
		diags.AddError("provider does not exists", fmt.Sprintf("there is no provider named '%s'", providerId))
	}

	c.provider = p
}

func (c *Controller[STATE]) Plans() []tmp.AddonPlan {
	if c.provider == nil {
		return []tmp.AddonPlan{}
	}

	return c.provider.Plans
}

func (c *Controller[STATE]) LookupFirstPlan() *tmp.AddonPlan {
	plans := c.Plans()
	if len(plans) == 0 {
		return nil
	}

	return &plans[0]
}

func (c *Controller[STATE]) LookupPlanBySlug(planSlug string) *tmp.AddonPlan {
	plans := c.Plans()
	if len(plans) == 0 {
		return nil
	}

	return pkg.LookupProviderPlan(c.provider, planSlug)
}
