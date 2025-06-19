package impl

import (
	"github.com/go-git/go-git/v5/plumbing/transport/http"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"go.clever-cloud.dev/client"
)

type Provider struct {
	version      string
	cc           *client.Client
	gitAuth      *http.BasicAuth
	organization string
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &Provider{version: version}
	}
}

func (p *Provider) Organization() string {
	return p.organization
}
func (p *Provider) Client() *client.Client {
	return p.cc
}

func (p *Provider) GitAuth() *http.BasicAuth {
	return p.gitAuth
}
