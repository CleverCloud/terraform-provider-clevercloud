package application

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.clever-cloud.com/terraform-provider/pkg/attributes"
)

// PlanDeploymentCommit resolves, at plan time, the commit an Update will
// deploy when `deployment.commit` is not configured: the repository HEAD.
//
// `commit` is Optional+Computed with UseStateForUnknown, so without this step
// the plan keeps the last deployed hash and a repository whose HEAD moved
// (new commit, directory recreated...) never shows a diff, hence is never
// deployed again. Setting the planned value to the current HEAD makes the
// change visible and triggers the Update, which pushes it.
//
// Nothing is done on create (the commit is resolved after the push), on
// destroy, when no repository is configured or when the user pinned a commit
// or reference. A HEAD resolution failure only emits a warning so an
// unreachable repository does not break unrelated plans.
func PlanDeploymentCommit(ctx context.Context, req resource.ModifyPlanRequest, res *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() || req.State.Raw.IsNull() {
		return
	}

	var cfg *attributes.Deployment
	res.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("deployment"), &cfg)...)
	if res.Diagnostics.HasError() || cfg == nil {
		return
	}
	if cfg.Repository.IsNull() || cfg.Repository.IsUnknown() || !cfg.Commit.IsNull() {
		return
	}

	var username, password *string
	if !cfg.BasicAuthentication.IsNull() && !cfg.BasicAuthentication.IsUnknown() {
		splits := strings.SplitN(cfg.BasicAuthentication.ValueString(), ":", 2)
		if len(splits) == 2 {
			username, password = &splits[0], &splits[1]
		}
	}

	head, diags := ResolveHeadCommit(ctx, cfg.Repository.ValueString(), username, password)
	if diags.HasError() {
		for _, d := range diags.Errors() {
			res.Diagnostics.AddAttributeWarning(
				path.Root("deployment").AtName("commit"),
				"cannot resolve repository HEAD, the commit to deploy is not planned",
				fmt.Sprintf("%s: %s", d.Summary(), d.Detail()),
			)
		}
		return
	}

	commitPath := path.Root("deployment").AtName("commit")

	var stateCommit types.String
	res.Diagnostics.Append(req.State.GetAttribute(ctx, commitPath, &stateCommit)...)
	if res.Diagnostics.HasError() {
		return
	}

	if !stateCommit.IsNull() && stateCommit.ValueString() == head {
		return
	}

	tflog.Info(ctx, "repository HEAD differs from the deployed commit, planning a deployment", map[string]any{
		"deployed": stateCommit.ValueString(),
		"head":     head,
	})
	res.Diagnostics.Append(res.Plan.SetAttribute(ctx, commitPath, types.StringValue(head))...)
}

// ResolveHeadCommit returns the HEAD commit hash of a repository without
// cloning it: a local `file://` repository is opened, a remote one is only
// listed (like `git ls-remote`).
func ResolveHeadCommit(ctx context.Context, repoURL string, username, password *string) (string, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	if strings.HasPrefix(repoURL, "file://") {
		repo, diags := open(repoURL)
		if diags.HasError() {
			return "", diags
		}

		head, err := repo.Head()
		if err != nil {
			diags.AddError("failed to get repository HEAD", err.Error())
			return "", diags
		}

		return head.Hash().String(), diags
	}

	remote := git.NewRemote(memory.NewStorage(), &config.RemoteConfig{
		Name: "origin",
		URLs: []string{repoURL},
	})

	listOpts := &git.ListOptions{}
	if username != nil && password != nil {
		listOpts.Auth = &http.BasicAuth{Username: *username, Password: *password}
	}

	refs, err := remote.ListContext(ctx, listOpts)
	if err != nil {
		diags.AddError("failed to list remote repository references", fmt.Sprintf("cannot list '%s': %s", repoURL, err.Error()))
		return "", diags
	}

	byName := map[plumbing.ReferenceName]*plumbing.Reference{}
	for _, ref := range refs {
		byName[ref.Name()] = ref
	}

	head, ok := byName[plumbing.HEAD]
	if !ok {
		diags.AddError("failed to get repository HEAD", fmt.Sprintf("no HEAD reference on '%s'", repoURL))
		return "", diags
	}

	// HEAD may be advertised as a symbolic reference to the default branch
	for range 5 {
		if head.Type() == plumbing.HashReference {
			return head.Hash().String(), diags
		}

		target, ok := byName[head.Target()]
		if !ok {
			diags.AddError("failed to get repository HEAD", fmt.Sprintf("HEAD points to unknown reference '%s' on '%s'", head.Target(), repoURL))
			return "", diags
		}
		head = target
	}

	diags.AddError("failed to get repository HEAD", fmt.Sprintf("too many symbolic references on '%s'", repoURL))
	return "", diags
}
