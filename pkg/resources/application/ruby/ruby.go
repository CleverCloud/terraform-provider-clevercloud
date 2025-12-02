package ruby

import (
	"context"

	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/resources"
	"go.clever-cloud.com/terraform-provider/pkg/resources/application"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type ResourceRuby struct {
	helper.Configurer
}

func NewResourceRuby() resource.Resource {
	return &ResourceRuby{}
}

func (r *ResourceRuby) Metadata(ctx context.Context, req resource.MetadataRequest, res *resource.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_ruby"
}

// UpgradeState implements state migration from version 0 to 1 for vhosts attribute
func (r *ResourceRuby) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		0: {
			PriorSchema: &schemaRubyV0,
			StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, res *resource.UpgradeStateResponse) {
				tflog.Info(ctx, "Upgrading Ruby resource state from version 0 to 1")

				old := helper.StateFrom[RubyV0](ctx, *req.State, &res.Diagnostics)
				if res.Diagnostics.HasError() {
					return
				}

				oldVhosts := []string{}
				res.Diagnostics.Append(old.VHosts.ElementsAs(ctx, &oldVhosts, false)...)
				vhosts := helper.VHostsFromAPIHosts(ctx, oldVhosts, old.VHosts, &res.Diagnostics)

				newState := Ruby{
					Runtime: application.Runtime{
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
						Networkgroups:    resources.NullNetworkgroupConfig,
					},
					RubyVersion:           old.RubyVersion,
					EnableSidekiq:         old.EnableSidekiq,
					RackupServer:          old.RackupServer,
					RakeGoals:             old.RakeGoals,
					SidekiqFiles:          old.SidekiqFiles,
					HTTPBasicAuth:         old.HTTPBasicAuth,
					NginxProxyBuffers:     old.NginxProxyBuffers,
					NginxProxyBufferSize:  old.NginxProxyBufferSize,
					EnableGzipCompression: old.EnableGzipCompression,
					GzipTypes:             old.GzipTypes,
					NginxReadTimeout:      old.NginxReadTimeout,
					RackEnv:               old.RackEnv,
					RailsEnv:              old.RailsEnv,
					StaticFilesPath:       old.StaticFilesPath,
					StaticURLPrefix:       old.StaticURLPrefix,
					StaticWebroot:         old.StaticWebroot,
				}

				res.Diagnostics.Append(res.State.Set(ctx, newState)...)
			},
		},
	}
}
