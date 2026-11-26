package matomo

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
)

var (
	_ resource.Resource               = &ResourceMatomo{}
	_ resource.ResourceWithModifyPlan = &ResourceMatomo{}
)

type ResourceMatomo struct {
	helper.AddonConfigurer
}

func NewResourceMatomo() resource.Resource {
	return &ResourceMatomo{}
}

func (r *ResourceMatomo) Metadata(ctx context.Context, req resource.MetadataRequest, res *resource.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_matomo"
}

// ModifyPlan refuses, at plan time, a version change to a version the add-on
// cannot move to. The versions on offer are read from the add-on itself, so a
// version asked for at creation is only checked by the API, at apply.
func (r *ResourceMatomo) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, res *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() || req.State.Raw.IsNull() || r.Provider == nil || r.Client() == nil {
		return
	}

	var id, current, wanted types.String
	res.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("id"), &id)...)
	res.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("version"), &current)...)
	res.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("version"), &wanted)...)
	if res.Diagnostics.HasError() || wanted.IsNull() || wanted.IsUnknown() || wanted.Equal(current) {
		return
	}

	checkRes := r.SDK.V4().AddonProviders().AddonMatomo().Addons().Addonmatomoid(id.ValueString()).Version().Check().
		CheckMatomoVersion(ctx)
	if checkRes.HasError() { // the API refuses a wrong version at apply
		tflog.Warn(ctx, "Skipping Matomo version validation", map[string]any{"error": checkRes.Error().Error()})
		return
	}

	available := checkRes.Payload().Available
	if !slices.Contains(available, wanted.ValueString()) {
		res.Diagnostics.AddAttributeError(
			path.Root("version"),
			"Matomo version not available",
			fmt.Sprintf("version %s not available, available versions: %s", wanted.ValueString(), strings.Join(available, ", ")),
		)
	}
}
