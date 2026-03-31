package oauth_consumer

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

func (r *ResourceOAuthConsumer) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	consumer := helper.PlanFrom[OAuthConsumer](ctx, req.Plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	rightsReq := r.setToRightsRequest(ctx, consumer.Rights, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create request
	createReq := tmp.OAuthConsumerRequest{
		Name:        consumer.Name.ValueString(),
		Description: consumer.Description.ValueString(),
		BaseURL:     consumer.BaseURL.ValueString(),
		LogoURL:     consumer.LogoURL.ValueString(),
		WebsiteURL:  consumer.WebsiteURL.ValueString(),
		Rights:      rightsReq,
	}

	res := tmp.CreateOAuthConsumer(ctx, r.Client(), r.Organization(), createReq)
	if res.HasError() {
		resp.Diagnostics.AddError("failed to create OAuth consumer", res.Error().Error())
		return
	}
	created := res.Payload()

	// Set computed values (ID is the OAuth consumer key)
	consumer.ID = pkg.FromStr(created.Key)

	resp.Diagnostics.Append(resp.State.Set(ctx, consumer)...)

	// Get the secret (requires separate API call)
	secretRes := tmp.GetOAuthConsumerSecret(ctx, r.Client(), r.Organization(), created.Key)
	if secretRes.HasError() {
		resp.Diagnostics.AddError("failed to get OAuth consumer secret", secretRes.Error().Error())
		return
	}
	secretData := secretRes.Payload()
	consumer.Secret = pkg.FromStr(secretData.Secret)

	resp.Diagnostics.Append(resp.State.Set(ctx, consumer)...)
}

func (r *ResourceOAuthConsumer) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	consumer := helper.StateFrom[OAuthConsumer](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	if consumer.ID.ValueString() == "" {
		resp.State.RemoveResource(ctx)
		return
	}

	consumerRes := tmp.GetOAuthConsumer(ctx, r.Client(), r.Organization(), consumer.ID.ValueString())
	if consumerRes.IsNotFoundError() {
		resp.State.RemoveResource(ctx)
		return
	} else if consumerRes.HasError() {
		resp.Diagnostics.AddError("failed to get OAuth consumer", consumerRes.Error().Error())
	} else {
		consumerData := consumerRes.Payload()
		consumer.Name = pkg.FromStr(consumerData.Name)
		consumer.Description = pkg.FromStr(consumerData.Description)
		consumer.BaseURL = pkg.FromStr(consumerData.BaseURL)
		consumer.LogoURL = pkg.FromStr(consumerData.LogoURL)
		consumer.WebsiteURL = pkg.FromStr(consumerData.WebsiteURL)
		consumer.Rights = r.rightsResponseToSet(ctx, consumerData.Rights, &resp.Diagnostics)
	}

	secretRes := tmp.GetOAuthConsumerSecret(ctx, r.Client(), r.Organization(), consumer.ID.ValueString())
	if secretRes.HasError() {
		resp.Diagnostics.AddError("failed to get OAuth consumer secret", secretRes.Error().Error())
	} else {
		secretData := secretRes.Payload()
		consumer.Secret = pkg.FromStr(secretData.Secret)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, consumer)...)
}

func (r *ResourceOAuthConsumer) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	plan := helper.PlanFrom[OAuthConsumer](ctx, req.Plan, &resp.Diagnostics)
	state := helper.StateFrom[OAuthConsumer](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.ID.ValueString() != state.ID.ValueString() {
		resp.Diagnostics.AddError("oauth_consumer cannot be updated", "mismatched IDs")
		return
	}

	rightsReq := r.setToRightsRequest(ctx, plan.Rights, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create update request
	updateReq := tmp.OAuthConsumerRequest{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
		BaseURL:     plan.BaseURL.ValueString(),
		LogoURL:     plan.LogoURL.ValueString(),
		WebsiteURL:  plan.WebsiteURL.ValueString(),
		Rights:      rightsReq,
	}

	res := tmp.UpdateOAuthConsumer(ctx, r.Client(), r.Organization(), plan.ID.ValueString(), updateReq)
	if res.HasError() {
		resp.Diagnostics.AddError("failed to update OAuth consumer", res.Error().Error())
		return
	} else {
		updated := res.Payload()
		state.Name = pkg.FromStr(updated.Name)
		state.Description = pkg.FromStr(updated.Description)
		state.BaseURL = pkg.FromStr(updated.BaseURL)
		state.LogoURL = pkg.FromStr(updated.LogoURL)
		state.WebsiteURL = pkg.FromStr(updated.WebsiteURL)
		state.Rights = r.rightsResponseToSet(ctx, updated.Rights, &resp.Diagnostics)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *ResourceOAuthConsumer) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	consumer := helper.StateFrom[OAuthConsumer](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	res := tmp.DeleteOAuthConsumer(ctx, r.Client(), r.Organization(), consumer.ID.ValueString())
	if res.HasError() && !res.IsNotFoundError() {
		resp.Diagnostics.AddError("failed to delete OAuth consumer", res.Error().Error())
	} else {
		resp.State.RemoveResource(ctx)
	}
}
