package staticapache

import (
	"context"

	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/resources"
	"go.clever-cloud.com/terraform-provider/pkg/resources/application"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type ResourceStaticApache struct {
	application.Configurer[*StaticApache]
}

func NewResourceStaticApache() func() resource.Resource {
	return func() resource.Resource {
		return &ResourceStaticApache{}
	}
}

func (r *ResourceStaticApache) Metadata(ctx context.Context, req resource.MetadataRequest, res *resource.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_static_apache"
}

// UpgradeState implements state migration from version 0 to 1 for vhosts attribute
func (r *ResourceStaticApache) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		0: {
			PriorSchema: &schemaStaticApacheV0,
			StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, res *resource.UpgradeStateResponse) {
				tflog.Info(ctx, "Upgrading StaticApache resource state from version 0 to 1")

				old := helper.StateFrom[StaticApacheV0](ctx, *req.State, &res.Diagnostics)
				if res.Diagnostics.HasError() {
					return
				}

				oldVhosts := []string{}
				res.Diagnostics.Append(old.VHosts.ElementsAs(ctx, &oldVhosts, false)...)
				vhosts := helper.VHostsFromAPIHosts(ctx, oldVhosts, old.VHosts, &res.Diagnostics)

				newState := StaticApache{
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
				}

				res.Diagnostics.Append(res.State.Set(ctx, newState)...)
			},
		},
	}
}

func (r *ResourceStaticApache) GetVariantSlug() string {
	return "static-apache"
}

// MigrationHint satisfies the application.VariantGuard interface. Rare path: a user
// pointed clevercloud_static_apache at an app that is actually a pure Static (nginx).
func (r *ResourceStaticApache) MigrationHint(actualSlug string) string {
	if actualSlug == "static" {
		return "This app uses the lightweight Static runtime (no Apache). " +
			"Manage it with clevercloud_static instead of clevercloud_static_apache."
	}
	return "Use the Terraform resource matching variant \"" + actualSlug + "\"."
}
