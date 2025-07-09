package application

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.clever-cloud.dev/client"
)

func gitDeploy(ctx context.Context, d Deployment, cc *client.Client, cleverRemote string) diag.Diagnostics {
	var diags diag.Diagnostics

	for range 5 {
		diags = _gitDeploy(ctx, d, cc, cleverRemote)
		if !diags.HasError() {
			break
		}

		time.Sleep(3 * time.Second)
	}
	return diags
}

func _gitDeploy(ctx context.Context, d Deployment, cc *client.Client, cleverRemote string) diag.Diagnostics {
	cleverRemote = strings.Replace(cleverRemote, "git+ssh", "https", 1) // switch protocol

	repo, diags := OpenOrClone(ctx, d.Repository, d.Commit)
	if diags.HasError() {
		return diags
	}

	currentRef, err := repo.Head()
	if err != nil {
		diags.AddError("failed to get current ref", err.Error())
		return diags
	}

	remoteOpts := &config.RemoteConfig{
		Name: "tf-clever",
		URLs: []string{cleverRemote, cleverRemote}, // for fetch and push
	}

	if err := repo.DeleteRemote("tf-clever"); err == nil {
		diags.AddWarning("a remote was set on this repository, it will be deleted", "remote = tf-clever")
	}

	remote, err := repo.CreateRemote(remoteOpts)
	if err != nil {
		diags.AddError("failed to add clever remote", err.Error())
		return diags
	}

	pushOptions := &git.PushOptions{
		RemoteName: "tf-clever",
		Force:      true,
		Progress:   os.Stdout,
		Auth:       d.CleverGitAuth,
		RefSpecs: []config.RefSpec{
			config.RefSpec(fmt.Sprintf("%s:%s", currentRef.Name(), plumbing.Master)),
		},
	}
	if d.Commit != nil {
		refNameOrCommit := *d.Commit
		var refSpec config.RefSpec

		// can be
		// refs/heads/[BRANCH]
		// or
		// [COMMIT_SHA] a397296e135b24e682a011e31f8e15f2fa8a5a0e

		if IsSHA1(refNameOrCommit) {
			commit, err := repo.CommitObject(plumbing.NewHash(refNameOrCommit))
			if err == plumbing.ErrObjectNotFound {
				diags.AddError("requested commit not found", fmt.Sprintf("no commit '%s'", refNameOrCommit))
				return diags
			}
			if err != nil {
				diags.AddError("failed to look for commit", err.Error())
				return diags
			}

			refSpec = config.RefSpec(fmt.Sprintf("%s:%s", commit.Hash.String(), plumbing.Master))
		} else {
			if !strings.HasPrefix(refNameOrCommit, "refs/") {
				refNameOrCommit = "refs/heads/" + refNameOrCommit
			}

			// We need to check if provided ref exists (several issues with main/master)
			ref, err := repo.Storer.Reference(plumbing.ReferenceName(refNameOrCommit))
			if err == plumbing.ErrReferenceNotFound {
				diags.AddError("requested reference not found", fmt.Sprintf("no reference named '%s'", refNameOrCommit))
				return diags
			}
			if err != nil {
				diags.AddError("failed to get reference", err.Error())
				return diags
			}

			refSpec = config.RefSpec(fmt.Sprintf("%s:%s", ref.Hash().String(), plumbing.Master))
		}

		if err := refSpec.Validate(); err != nil {
			diags.AddError("failed to build ref spec to push", err.Error())
			return diags
		}

		pushOptions.RefSpecs = []config.RefSpec{refSpec}
	}

	tflog.Debug(ctx, "pushing...", map[string]any{
		"options": fmt.Sprintf("%+v", pushOptions),
	})

	err = remote.PushContext(ctx, pushOptions)
	if err != nil {
		if err == git.NoErrAlreadyUpToDate {
			diags.AddWarning("Git push rejected", "repository is already up-to-date")
		} else {
			diags.AddError("failed to push to clever remote", err.Error())
		}
	}

	return diags
}

func IsSHA1(s string) bool {
	h, err := hex.DecodeString(s)
	return err == nil && len(h) == sha1.Size
}

func OpenOrClone(ctx context.Context, repoUrl string, commit *string) (*git.Repository, diag.Diagnostics) {
	if strings.HasPrefix(repoUrl, "file://") {
		return open(repoUrl)
	}

	return clone(ctx, repoUrl, commit)
}

func open(repoUrl string) (*git.Repository, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	repo, err := git.PlainOpen(strings.TrimPrefix(repoUrl, "file://"))
	if err != nil {
		diags.AddError("failed to open repository", fmt.Sprintf("cannot open '%s': %s", repoUrl, err.Error()))
		return nil, diags
	}

	return repo, diags
}

func clone(ctx context.Context, repoUrl string, commit *string) (*git.Repository, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	fs := memory.NewStorage()
	wt := memfs.New()

	cloneOpts := &git.CloneOptions{
		URL:        repoUrl,
		RemoteName: "origin",
		Progress:   os.Stdout,
	}

	if commit != nil && strings.HasPrefix(*commit, "refs/") {
		cloneOpts.ReferenceName = plumbing.ReferenceName(*commit)
	}

	r, err := git.CloneContext(ctx, fs, wt, cloneOpts)
	if err != nil {
		diags.AddError("failed to clone repository", err.Error())
		return nil, diags
	}

	return r, diags
}
