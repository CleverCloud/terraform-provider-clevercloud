package frankenphp

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/resources"
	"go.clever-cloud.com/terraform-provider/pkg/resources/application"
)

type ResourceFrankenPHP struct {
	application.Configurer[*FrankenPHP]
}

func NewResourceFrankenPHP() resource.Resource {
	return &ResourceFrankenPHP{}
}

func (r *ResourceFrankenPHP) Metadata(ctx context.Context, req resource.MetadataRequest, res *resource.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_frankenphp"
}

// UpgradeState implements state migration from version 0 to 1: vhosts turn from
// "fqdn/path" strings into {fqdn, path_begin} objects, and the attributes added since
// (networkgroups, exposed_environment, integrations, redirection) start out null.
func (r *ResourceFrankenPHP) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		0: {
			PriorSchema: &schemaFrankenPHPV0,
			StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, res *resource.UpgradeStateResponse) {
				tflog.Info(ctx, "Upgrading FrankenPHP resource state from version 0 to 1")

				old := helper.StateFrom[FrankenPHPV0](ctx, *req.State, &res.Diagnostics)
				if res.Diagnostics.HasError() {
					return
				}

				oldVhosts := []string{}
				res.Diagnostics.Append(old.VHosts.ElementsAs(ctx, &oldVhosts, false)...)
				if res.Diagnostics.HasError() {
					return
				}
				vhosts := helper.VHostsFromAPIHosts(ctx, oldVhosts, old.VHosts, &res.Diagnostics)

				newState := FrankenPHP{
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
						Redirection:        nil,
						AppFolder:          old.AppFolder,
						Environment:        old.Environment,
						Networkgroups:      resources.NullNetworkgroupConfig,
						ExposedEnvironment: application.NullExposedEnv,
					},
					DevDependencies: old.DevDependencies,
				}

				res.Diagnostics.Append(res.State.Set(ctx, newState)...)
			},
		},
	}
}

func (r *ResourceFrankenPHP) GetVariantSlug() string {
	return "frankenphp"
}
