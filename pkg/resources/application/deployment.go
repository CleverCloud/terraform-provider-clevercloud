package application

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg/attributes"
)

// Deploy pushes the configured repository to the application Clever remote on
// Create, then resolves the computed `deployment.commit` attribute with the
// commit actually deployed (null when no push happened).
func Deploy(ctx context.Context, resource RuntimeResource, plan RuntimePlan, diags *diag.Diagnostics) {
	runtime := plan.GetRuntimePtr()

	commit, _ := GitDeploy(ctx, plan.ToDeployment(resource.GitAuth()), runtime.DeployURL.ValueString(), "", diags)
	resolveUnknownCommit(runtime.Deployment, commit)
}

// resolveUnknownCommit fills the computed `deployment.commit` attribute after
// a Create/Update: when the commit is not set in the configuration, its
// planned value is unknown and must be resolved to a known value before
// saving the state. Values coming from the configuration or the previous
// state are already known and kept as-is.
func resolveUnknownCommit(d *attributes.Deployment, commit string) {
	if d == nil || !d.Commit.IsUnknown() {
		return
	}

	if commit == "" {
		d.Commit = types.StringNull()
	} else {
		d.Commit = types.StringValue(commit)
	}
}

// syncDeploymentCommit reflects in the state the commit currently running on
// the application (CommitID from the API):
//   - no deployment block: the user does not manage deployments with
//     Terraform, nothing is reconciled so no diff shows up
//   - commit holding a git reference (refs/heads/...) or `github_hook`: kept
//     as-is, overwriting it with the running hash would fight the
//     configuration and create a permanent diff
//   - otherwise (null or commit hash): the running commit is reported, so a
//     deployment done outside Terraform is visible in the state and a pinned
//     commit detects drift
func syncDeploymentCommit(d *attributes.Deployment, runningCommit string) {
	if d == nil || runningCommit == "" {
		return
	}

	if !d.Commit.IsNull() && !IsSHA1(d.Commit.ValueString()) {
		return
	}

	d.Commit = types.StringValue(runningCommit)
}
