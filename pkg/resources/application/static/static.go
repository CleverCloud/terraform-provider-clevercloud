package static

import (
	"context"

	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/resources"
	"go.clever-cloud.com/terraform-provider/pkg/resources/application"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type ResourceStatic struct {
	application.Configurer[*Static]
}

func NewResourceStatic() func() resource.Resource {
	return func() resource.Resource {
		return &ResourceStatic{}
	}
}

func (r *ResourceStatic) Metadata(ctx context.Context, req resource.MetadataRequest, res *resource.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_static"
}

// UpgradeState implements state migration from version 0 to 1 for vhosts attribute
func (r *ResourceStatic) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		0: {
			PriorSchema: &schemaStaticV0,
			StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, res *resource.UpgradeStateResponse) {
				tflog.Info(ctx, "Upgrading Static resource state from version 0 to 1")

				old := helper.StateFrom[StaticV0](ctx, *req.State, &res.Diagnostics)
				if res.Diagnostics.HasError() {
					return
				}

				oldVhosts := []string{}
				res.Diagnostics.Append(old.VHosts.ElementsAs(ctx, &oldVhosts, false)...)
				vhosts := helper.VHostsFromAPIHosts(ctx, oldVhosts, old.VHosts, &res.Diagnostics)

				newState := Static{
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
						Tags:               resources.NullTags,
						ExposedEnvironment: application.NullExposedEnv,
					},
				}

				res.Diagnostics.Append(res.State.Set(ctx, newState)...)
			},
		},
	}
}

func (r *ResourceStatic) GetVariantSlug() string {
	return "static"
}

// MigrationHint satisfies the application.VariantGuard interface. It guides users
// whose state was created under the legacy clevercloud_static (which at the time
// meant "Static with Apache") after the v1.12.0 rename.
func (r *ResourceStatic) MigrationHint(actualSlug string) string {
	if actualSlug == "static-apache" {
		return "This app was created as Static with Apache. The clevercloud_static " +
			"resource now manages the lighter Static runtime (no Apache).\n" +
			"To continue managing this app, migrate your state with a moved block:\n\n" +
			"  moved {\n" +
			"    from = clevercloud_static.<name>\n" +
			"    to   = clevercloud_static_apache.<name>\n" +
			"  }\n\n" +
			"Then rename your resource block from \"clevercloud_static\" to \"clevercloud_static_apache\".\n" +
			"See the v1.12.0 release notes for details."
	}
	return "Use the Terraform resource matching variant \"" + actualSlug + "\"."
}
