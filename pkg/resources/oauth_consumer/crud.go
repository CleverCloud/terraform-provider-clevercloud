package oauth_consumer

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.dev/sdk/models"
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
	createReq := models.WannabeOAuth1Consumer{
		Name:        pkg.Ptr(consumer.Name.ValueString()),
		Description: pkg.Ptr(consumer.Description.ValueString()),
		BaseURL:     pkg.Ptr(consumer.BaseURL.ValueString()),
		Picture:     pkg.Ptr(consumer.LogoURL.ValueString()),
		URL:         pkg.Ptr(consumer.WebsiteURL.ValueString()),
		Rights:      rightsReq,
	}

	res := r.SDK.V2().Organisations().ID(r.Organization()).Consumers().Createorganisationconsumer(ctx, &createReq)
	if res.HasError() {
		resp.Diagnostics.AddError("failed to create OAuth consumer", res.Error().Error())
		return
	}
	created := res.Payload()

	// Set computed values (ID is the OAuth consumer key)
	consumer.ID = pkg.FromStr(created.Key)

	resp.Diagnostics.Append(resp.State.Set(ctx, consumer)...)

	// Get the secret (requires separate API call)
	secretRes := r.SDK.V2().Organisations().ID(r.Organization()).Consumers().Key(created.Key).Secret().Getorganisationconsumersecret(ctx)
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

	consumerRes := r.SDK.V2().Organisations().ID(r.Organization()).Consumers().Key(consumer.ID.ValueString()).Getorganisationconsumer(ctx)
	if consumerRes.IsNotFoundError() {
		resp.State.RemoveResource(ctx)
		return
	} else if consumerRes.HasError() {
		resp.Diagnostics.AddError("failed to get OAuth consumer", consumerRes.Error().Error())
	} else {
		consumerData := consumerRes.Payload()
		consumer.Name = pkg.FromStr(pkg.Deref(consumerData.Name))
		consumer.Description = pkg.FromStr(pkg.Deref(consumerData.Description))
		consumer.BaseURL = pkg.FromStr(pkg.Deref(consumerData.BaseURL))
		consumer.LogoURL = pkg.FromStr(pkg.Deref(consumerData.Picture))
		consumer.WebsiteURL = pkg.FromStr(pkg.Deref(consumerData.URL))
		consumer.Rights = r.rightsResponseToSet(ctx, consumerData.Rights, &resp.Diagnostics)
	}

	secretRes := r.SDK.V2().Organisations().ID(r.Organization()).Consumers().Key(consumer.ID.ValueString()).Secret().Getorganisationconsumersecret(ctx)
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
	updateReq := models.WannabeOAuth1Consumer{
		Name:        pkg.Ptr(plan.Name.ValueString()),
		Description: pkg.Ptr(plan.Description.ValueString()),
		BaseURL:     pkg.Ptr(plan.BaseURL.ValueString()),
		Picture:     pkg.Ptr(plan.LogoURL.ValueString()),
		URL:         pkg.Ptr(plan.WebsiteURL.ValueString()),
		Rights:      rightsReq,
	}

	res := r.SDK.V2().Organisations().ID(r.Organization()).Consumers().Key(plan.ID.ValueString()).Updateorganisationconsumer(ctx, &updateReq)
	if res.HasError() {
		resp.Diagnostics.AddError("failed to update OAuth consumer", res.Error().Error())
		return
	} else {
		updated := res.Payload()
		state.Name = pkg.FromStr(pkg.Deref(updated.Name))
		state.Description = pkg.FromStr(pkg.Deref(updated.Description))
		state.BaseURL = pkg.FromStr(pkg.Deref(updated.BaseURL))
		state.LogoURL = pkg.FromStr(pkg.Deref(updated.Picture))
		state.WebsiteURL = pkg.FromStr(pkg.Deref(updated.URL))
		state.Rights = r.rightsResponseToSet(ctx, updated.Rights, &resp.Diagnostics)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *ResourceOAuthConsumer) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	consumer := helper.StateFrom[OAuthConsumer](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	res := r.SDK.V2().Organisations().ID(r.Organization()).Consumers().Key(consumer.ID.ValueString()).Deleteorganisationconsumer(ctx)
	if res.HasError() && !res.IsNotFoundError() {
		resp.Diagnostics.AddError("failed to delete OAuth consumer", res.Error().Error())
	} else {
		resp.State.RemoveResource(ctx)
	}
}
