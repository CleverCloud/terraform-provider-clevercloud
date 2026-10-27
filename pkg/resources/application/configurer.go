package application

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
)

type Configurer[T RuntimePlan] struct {
	helper.Configurer
}

func (c Configurer[T]) Delete(ctx context.Context, req resource.DeleteRequest, res *resource.DeleteResponse) {
	state := helper.From[T](ctx, req.State, &res.Diagnostics)
	if res.Diagnostics.HasError() {
		return
	}

	runtime := (*state).GetRuntimePtr()

	deleteRes := c.SDK.V2().Organisations().ID(c.Organization()).Applications().Appid(runtime.ID.ValueString()).Deleteorganisationapplication(ctx)
	if deleteRes.HasError() && !deleteRes.IsNotFoundError() {
		res.Diagnostics.AddError("failed to delete app", deleteRes.Error().Error())
	} else {
		res.State.RemoveResource(ctx)
	}
}
