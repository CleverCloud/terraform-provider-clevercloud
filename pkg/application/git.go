package application

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.clever-cloud.dev/client"
)

func gitDeploy(ctx context.Context, d Deployment, cc *client.Client, cleverRemote string) diag.Diagnostics {
	diags := diag.Diagnostics{}
	cleverRemote = strings.Replace(cleverRemote, "git+ssh", "https", 1) + ".git" // switch protocol
	fs := memory.NewStorage()
	cloneOpts := &git.CloneOptions{
		URL:        d.Repository,
		RemoteName: "origin",
		Progress:   os.Stdout,
	}

	r, err := git.CloneContext(ctx, fs, nil, cloneOpts)
	if err != nil {
		diags.AddError("failed to clone repository", err.Error())
		return diags
	}

	remoteOpts := &config.RemoteConfig{
		Name: "clever",
		URLs: []string{cleverRemote}, // for fetch and push
	}

	remote, err := r.CreateRemote(remoteOpts)
	if err != nil {
		diags.AddError("failed to add clever remote", err.Error())
		return diags
	}

	token, secret := cc.Oauth1UserCredentials()
	auth := &http.BasicAuth{Username: token, Password: secret}

	pushOptions := &git.PushOptions{
		RemoteName: "clever",
		Force:      true,
		Progress:   os.Stdout,
		Auth:       auth,
	}
	if d.Commit != nil {
		// can be
		// refs/heads/[BRANCH]
		// or
		// [COMMIT_SHA]

		// We need to check if provided ref exists (several issues with main/master)
		_, err = r.Storer.Reference(plumbing.ReferenceName(*d.Commit))
		if err != nil && err != plumbing.ErrReferenceNotFound {
			diags.AddError("failed to get reference", err.Error())
			return diags
		}

		hasCommitErr := r.Storer.HasEncodedObject(plumbing.NewHash(*d.Commit))
		if hasCommitErr != nil && hasCommitErr != plumbing.ErrObjectNotFound {
			diags.AddError("failed to get commit", hasCommitErr.Error())
			return diags
		}

		if err == plumbing.ErrReferenceNotFound && hasCommitErr == plumbing.ErrObjectNotFound {
			diags.AddError("unknown reference", fmt.Sprintf("commit or reference %s not found", *d.Commit))
			return diags
		}

		refStr := fmt.Sprintf("%s:%s", *d.Commit, plumbing.Master)
		tflog.Debug(ctx, "refspec", map[string]any{"ref": refStr})

		ref := config.RefSpec(refStr)
		if err := ref.Validate(); err != nil {
			diags.AddError("failed to build ref spec to push", err.Error())
			return diags
		}

		pushOptions.RefSpecs = []config.RefSpec{ref}
	} else {
		pushOptions.RefSpecs = []config.RefSpec{
			config.RefSpec(fmt.Sprintf("%s:%s", plumbing.HEAD, plumbing.Master)),
		}
	}

	tflog.Debug(ctx, "pushing...", map[string]any{
		"options": fmt.Sprintf("%+v", pushOptions),
	})

	err = remote.PushContext(ctx, pushOptions)
	if err != nil /*&& err != git.NoErrAlreadyUpToDate*/ {
		diags.AddError("failed to push to clever remote", err.Error())
		return diags
	}

	return diags
}
