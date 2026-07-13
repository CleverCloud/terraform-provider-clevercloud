package application

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"go.clever-cloud.com/terraform-provider/pkg/attributes"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

type Configurer[T RuntimePlan] struct {
	helper.Configurer
}

// ValidateConfig checks cross-attribute constraints shared by all runtimes:
// `branch` selects the branch deployed by a GitHub-linked application
// (`deployment.commit = "github_hook"`), so it is required there and must not
// be defined anywhere else.
func (c Configurer[T]) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, res *resource.ValidateConfigResponse) {
	if req.Config.Raw.IsNull() {
		return
	}

	config := helper.From[T](ctx, req.Config, &res.Diagnostics)
	if res.Diagnostics.HasError() || config == nil {
		return
	}
	runtime := (*config).GetRuntimePtr()

	deployment := runtime.Deployment
	if runtime.Branch.IsUnknown() || (deployment != nil && deployment.Commit.IsUnknown()) {
		return // cannot judge until the values are known
	}

	branchSet := runtime.Branch.ValueString() != "" // unset, null and empty all count as unset
	githubLinked := deployment != nil && strings.HasPrefix(deployment.Commit.ValueString(), attributes.GITHUB_COMMIT_PREFIX)
	if branchSet == githubLinked {
		return // consistent configuration
	}

	if githubLinked {
		res.Diagnostics.AddAttributeError(
			path.Root("branch"),
			"branch is required for GitHub-linked applications",
			"An application linked to a GitHub repository (`deployment.commit = \"github_hook\"`) deploys the branch selected by `branch`: set it to the repository branch to deploy.",
		)
	} else {
		res.Diagnostics.AddAttributeError(
			path.Root("branch"),
			"branch only applies to GitHub-linked applications",
			"`branch` selects the branch deployed by a linked GitHub repository: it requires `deployment.commit` to be set to `github_hook`.",
		)
	}
}

func (c Configurer[T]) Delete(ctx context.Context, req resource.DeleteRequest, res *resource.DeleteResponse) {
	state := helper.From[T](ctx, req.State, &res.Diagnostics)
	if res.Diagnostics.HasError() {
		return
	}

	runtime := (*state).GetRuntimePtr()

	deleteRes := tmp.DeleteApp(ctx, c.Client(), c.Organization(), runtime.ID.ValueString())
	if deleteRes.HasError() && !deleteRes.IsNotFoundError() {
		res.Diagnostics.AddError("failed to delete app", deleteRes.Error().Error())
	} else {
		res.State.RemoveResource(ctx)
	}
}
