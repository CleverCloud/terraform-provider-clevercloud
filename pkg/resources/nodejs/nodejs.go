package nodejs

import (
	"context"

	"go.clever-cloud.com/terraform-provider/pkg/attributes"
	"go.clever-cloud.com/terraform-provider/pkg/helper"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type ResourceNodeJS struct {
	helper.Configurer
}

func NewResourceNodeJS() resource.Resource {
	return &ResourceNodeJS{}
}

func (r *ResourceNodeJS) Metadata(ctx context.Context, req resource.MetadataRequest, res *resource.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_nodejs"
}

// UpgradeState implements state migration from version 0 to 1 for vhosts attribute
func (r *ResourceNodeJS) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		0: {
			PriorSchema: &schemaNodeJSV0,
			StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, res *resource.UpgradeStateResponse) {
				tflog.Info(ctx, "Upgrading NodeJS resource state from version 0 to 1")

				old := helper.StateFrom[NodeJSV0](ctx, *req.State, &res.Diagnostics)
				if res.Diagnostics.HasError() {
					return
				}

				oldVhosts := []string{}
				res.Diagnostics.Append(old.VHosts.ElementsAs(ctx, &oldVhosts, false)...)
				vhosts := helper.VHostsFromAPIHosts(ctx, oldVhosts, old.VHosts, &res.Diagnostics)

				newState := NodeJS{
					Runtime: attributes.Runtime{
						ID:               old.ID,
						Name:             old.Name,
						Description:      old.Description,
						MinInstanceCount: old.MinInstanceCount,
						MaxInstanceCount: old.MaxInstanceCount,
						SmallestFlavor:   old.SmallestFlavor,
						BiggestFlavor:    old.BiggestFlavor,
						BuildFlavor:      old.BuildFlavor,
						Region:           old.Region,
						StickySessions:   old.StickySessions,
						RedirectHTTPS:    old.RedirectHTTPS,
						VHosts:           vhosts,
						DeployURL:        old.DeployURL,
						Dependencies:     old.Dependencies,
						Deployment:       old.Deployment,
						Hooks:            old.Hooks,
						AppFolder:        old.AppFolder,
						Environment:      old.Environment,
					},
					DevDependencies: old.DevDependencies,
					StartScript:     old.StartScript,
					PackageManager:  old.PackageManager,
					Registry:        old.Registry,
					RegistryToken:   old.RegistryToken,
				}

				res.Diagnostics.Append(res.State.Set(ctx, newState)...)
			},
		},
	}
}
