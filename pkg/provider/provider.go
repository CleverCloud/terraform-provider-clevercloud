package provider

import (
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"go.clever-cloud.dev/client"
	"go.clever-cloud.dev/sdk"
)

type Provider interface {
	Organization() string
	Client() *client.Client
	SDK() sdk.SDK
	GitAuth() *http.BasicAuth
	IsNetwrkgroupsDisabled() bool
}
