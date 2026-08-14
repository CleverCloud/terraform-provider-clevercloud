package otoroshi

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
)

var _ resource.ResourceWithUpgradeState = &ResourceOtoroshi{}

type ResourceOtoroshi struct {
	helper.Configurer
}

func NewResourceOtoroshi() resource.Resource {
	return &ResourceOtoroshi{}
}

func (r *ResourceOtoroshi) Metadata(ctx context.Context, req resource.MetadataRequest, res *resource.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_otoroshi"
}

// UpgradeState implements state migration from version 0 to 1, where networkgroups
// was dropped: membership was piloted on the backing Java application, which removed
// the add-on from its own dedicated networkgroup on every apply
func (r *ResourceOtoroshi) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		0: {
			PriorSchema: &schemaOtoroshiV0,
			StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, res *resource.UpgradeStateResponse) {
				tflog.Info(ctx, "Upgrading Otoroshi resource state from version 0 to 1")

				old := helper.StateFrom[OtoroshiV0](ctx, *req.State, &res.Diagnostics)
				if res.Diagnostics.HasError() {
					return
				}

				// networkgroup and networkgroup_id are left null on purpose, Read fills
				// them from the add-on features on the next refresh
				res.Diagnostics.Append(res.State.Set(ctx, Otoroshi{
					ID:                   old.ID,
					Name:                 old.Name,
					Region:               old.Region,
					CreationDate:         old.CreationDate,
					Version:              old.Version,
					APIClientID:          old.APIClientID,
					APIClientSecret:      old.APIClientSecret,
					APIURL:               old.APIURL,
					InitialAdminLogin:    old.InitialAdminLogin,
					InitialAdminPassword: old.InitialAdminPassword,
					URL:                  old.URL,
				})...)
			},
		},
	}
}
