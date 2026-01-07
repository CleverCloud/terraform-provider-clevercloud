package php

import (
	"context"

	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/resources"
	"go.clever-cloud.com/terraform-provider/pkg/resources/application"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type ResourcePHP struct {
	application.Configurer[*PHP]
}

func NewResourcePHP() resource.Resource {
	return &ResourcePHP{}
}

func (r *ResourcePHP) Metadata(ctx context.Context, req resource.MetadataRequest, res *resource.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_php"
}

// UpgradeState implements state migration from version 0 to 1 for vhosts attribute
func (r *ResourcePHP) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		0: {
			PriorSchema: &schemaPHPV0,
			StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, res *resource.UpgradeStateResponse) {
				tflog.Info(ctx, "Upgrading PHP resource state from version 0 to 1")

				old := helper.StateFrom[PHPV0](ctx, *req.State, &res.Diagnostics)
				if res.Diagnostics.HasError() {
					return
				}

				oldVhosts := []string{}
				res.Diagnostics.Append(old.VHosts.ElementsAs(ctx, &oldVhosts, false)...)
				vhosts := helper.VHostsFromAPIHosts(ctx, oldVhosts, old.VHosts, &res.Diagnostics)

				newState := PHP{
					Runtime: application.Runtime{
						ID:                 old.ID,
						Name:               old.Name,
						Description:        old.Description,
						MinInstanceCount:   old.MinInstanceCount,
						MaxInstanceCount:   old.MaxInstanceCount,
						SmallestFlavor:     old.SmallestFlavor,
						BiggestFlavor:      old.BiggestFlavor,
						BuildFlavor:        old.BuildFlavor,
						Region:             old.Region,
						StickySessions:     old.StickySessions,
						RedirectHTTPS:      old.RedirectHTTPS,
						VHosts:             vhosts,
						DeployURL:          old.DeployURL,
						Dependencies:       old.Dependencies,
						Deployment:         old.Deployment,
						Hooks:              old.Hooks,
						Integrations:       nil,
						AppFolder:          old.AppFolder,
						Environment:        old.Environment,
						Networkgroups:      resources.NullNetworkgroupConfig,
						ExposedEnvironment: application.NullExposedEnv,
					},
					PHPVersion:      old.PHPVersion,
					WebRoot:         old.WebRoot,
					RedisSessions:   old.RedisSessions,
					DevDependencies: old.DevDependencies,
				}

				res.Diagnostics.Append(res.State.Set(ctx, newState)...)
			},
		},
	}
}

func (r *ResourcePHP) GetVariantSlug() string {
	return "php"
}
