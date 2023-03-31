package application

import (
	"context"
	"os"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"go.clever-cloud.dev/client"
)

func gitDeploy(ctx context.Context, d Deployment, cc *client.Client, cleverRemote string, diags diag.Diagnostics) {
	cleverRemote = strings.Replace(cleverRemote, "git+ssh", "https", 1) // switch protocol

	cloneOpts := &git.CloneOptions{
		URL:        d.Repository,
		RemoteName: "origin",
		//Depth:      1,
		Progress: os.Stdout,
	}

	r, err := git.CloneContext(ctx, memory.NewStorage(), nil, cloneOpts)
	if err != nil {
		diags.AddError("failed to clone repository", err.Error())
		return
	}

	remoteOpts := &config.RemoteConfig{
		Name: "clever",
		URLs: []string{cleverRemote}, // for fetch and push
	}

	remote, err := r.CreateRemote(remoteOpts)
	if err != nil {
		diags.AddError("failed to add clever remote", err.Error())
		return
	}

	token, secret := cc.Oauth1UserCredentials()
	auth := &http.BasicAuth{Username: token, Password: secret}

	pushOptions := &git.PushOptions{
		RemoteName: "clever",
		RemoteURL:  cleverRemote,
		Force:      true,
		Progress:   os.Stdout,
		// TODO: deploy right branch/tag/commit
		/*RefSpecs: []config.RefSpec{
			"master:master",
		},*/
		Auth:   auth,
		Atomic: true,
	}
	err = remote.Push(pushOptions)
	if err != nil {
		diags.AddError("failed to push to clever remote", err.Error())
		return
	}
}
