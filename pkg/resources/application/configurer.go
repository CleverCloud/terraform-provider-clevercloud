package application

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
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

	deleteRes := tmp.DeleteApp(ctx, c.Client(), c.Organization(), runtime.ID.ValueString())
	if deleteRes.HasError() && !deleteRes.IsNotFoundError() {
		res.Diagnostics.AddError("failed to delete app", deleteRes.Error().Error())
	} else {
		res.State.RemoveResource(ctx)
	}
}
