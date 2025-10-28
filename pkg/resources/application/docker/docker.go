package docker

import (
	application "go.clever-cloud.com/terraform-provider/pkg/helper/application"
	"context"

	"go.clever-cloud.com/terraform-provider/pkg/helper"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type ResourceDocker struct {
	helper.Configurer
}

func NewResourceDocker() resource.Resource {
	return &ResourceDocker{}
}

func (r *ResourceDocker) Metadata(ctx context.Context, req resource.MetadataRequest, res *resource.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_docker"
}

// https://developer.hashicorp.com/terraform/plugin/framework/resources/state-upgrade#implementing-state-upgrade-support
// State is only available if you provide PriorSchema, else, use RawState
func (r *ResourceDocker) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		0: {
			PriorSchema: &schemaDockerV0,
			StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, res *resource.UpgradeStateResponse) {
				old := helper.StateFrom[Docker](ctx, *req.State, &res.Diagnostics)
				if res.Diagnostics.HasError() {
					return
				}

				oldVhosts := []string{}
				res.Diagnostics.Append(old.VHosts.ElementsAs(ctx, &oldVhosts, false)...)
				vhosts := helper.VHostsFromAPIHosts(ctx, oldVhosts, old.VHosts, &res.Diagnostics)

				newState := Docker{
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
					},
					Dockerfile:        old.Dockerfile,
					ContainerPort:     old.ContainerPort,
					ContainerPortTCP:  old.ContainerPortTCP,
					EnableIPv6:        old.EnableIPv6,
					IPv6Cidr:          old.IPv6Cidr,
					RegistryURL:       old.RegistryURL,
					RegistryUser:      old.RegistryUser,
					RegistryPassword:  old.RegistryPassword,
					DaemonSocketMount: old.DaemonSocketMount,
				}
				tflog.Warn(ctx, "\n\n##### TO MIGRATE", map[string]any{"old": newState.VHosts.Elements()[0].String()})

				res.Diagnostics.Append(res.State.Set(ctx, newState)...)
			},
		},
	}
}
