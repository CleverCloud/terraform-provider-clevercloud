package application

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.clever-cloud.dev/client"
)

func gitDeploy(ctx context.Context, d Deployment, cc *client.Client, cleverRemote string) diag.Diagnostics {
	diags := diag.Diagnostics{}

	cleverRemote = strings.Replace(cleverRemote, "git+ssh", "https", 1) // switch protocol

	cloneOpts := &git.CloneOptions{
		URL:        d.Repository,
		RemoteName: "origin",
		Progress:   os.Stdout,
	}

	r, err := git.CloneContext(ctx, memory.NewStorage(), nil, cloneOpts)
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
		RemoteURL:  cleverRemote,
		Force:      true,
		Progress:   os.Stdout,
		Auth:       auth,
	}
	if d.Commit != nil {
		// refs/heads/[BRANCH]
		// [COMMIT]
		refStr := fmt.Sprintf("%s:refs/heads/master", *d.Commit)
		tflog.Debug(ctx, "refspec", map[string]any{"ref": refStr})
		ref := config.RefSpec(refStr)
		if err := ref.Validate(); err != nil {
			diags.AddError("failed to build ref spec to push", err.Error())
			return diags
		}

		pushOptions.RefSpecs = []config.RefSpec{ref}
	} else {
		pushOptions.RefSpecs = []config.RefSpec{
			config.RefSpec("main:refs/heads/master"),
		}
	}

	tflog.Debug(ctx, "pushing...", map[string]any{
		"options": fmt.Sprintf("%+v", pushOptions),
	})

	err = remote.PushContext(ctx, pushOptions)
	if err != nil && err != git.NoErrAlreadyUpToDate {
		diags.AddError("failed to push to clever remote", err.Error())
		return diags
	}

	return diags
}
