package java

import (
	"context"

	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/resources/application"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const (
	CC_DISABLE_MAX_METASPACE = "CC_DISABLE_MAX_METASPACE"
	CC_EXTRA_JAVA_ARGS       = "CC_EXTRA_JAVA_ARGS"
	CC_JAR_ARGS              = "CC_JAR_ARGS"
	CC_JAR_PATH              = "CC_JAR_PATH"
	CC_JAVA_VERSION          = "CC_JAVA_VERSION"
	CC_MAVEN_PROFILES        = "CC_MAVEN_PROFILES"
	CC_RUN_COMMAND           = "CC_RUN_COMMAND"
	CC_SBT_TARGET_BIN        = "CC_SBT_TARGET_BIN"
	CC_SBT_TARGET_DIR        = "CC_SBT_TARGET_DIR"
	GRADLE_DEPLOY_GOAL       = "GRADLE_DEPLOY_GOAL"
	MAVEN_DEPLOY_GOAL        = "MAVEN_DEPLOY_GOAL"
	NUDGE_APPID              = "NUDGE_APPID"
	PLAY1_VERSION            = "PLAY1_VERSION"
	SBT_DEPLOY_GOAL          = "SBT_DEPLOY_GOAL"
)

type ResourceJava struct {
	helper.Configurer
	profile string
}

func NewResourceJava(profile string) func() resource.Resource {
	return func() resource.Resource {
		return &ResourceJava{profile: profile}
	}
}

func (r *ResourceJava) Metadata(ctx context.Context, req resource.MetadataRequest, res *resource.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_java"

	if r.profile != "" {
		res.TypeName = res.TypeName + "_" + r.profile
	}
}

// UpgradeState implements state migration from version 0 to 1 for vhosts attribute
func (r *ResourceJava) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		0: {
			PriorSchema: &schemaJavaV0,
			StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, res *resource.UpgradeStateResponse) {
				tflog.Info(ctx, "Upgrading Java resource state from version 0 to 1")

				old := helper.StateFrom[JavaV0](ctx, *req.State, &res.Diagnostics)
				if res.Diagnostics.HasError() {
					return
				}

				oldVhosts := []string{}
				res.Diagnostics.Append(old.VHosts.ElementsAs(ctx, &oldVhosts, false)...)
				vhosts := helper.VHostsFromAPIHosts(ctx, oldVhosts, old.VHosts, &res.Diagnostics)

				newState := Java{
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
					JavaVersion: old.JavaVersion,
				}

				res.Diagnostics.Append(res.State.Set(ctx, newState)...)
			},
		},
	}
}
