package impl

import (
	"github.com/go-git/go-git/v5/plumbing/transport/http"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"go.clever-cloud.dev/client"
	"go.clever-cloud.dev/sdk"
)

type Provider struct {
	version                string
	cc                     *client.Client
	gitAuth                *http.BasicAuth
	organization           string
	isNetwrkgroupsDisabled bool
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

func (p *Provider) SDK() sdk.SDK {
	return sdk.NewSDK(sdk.WithClient(p.Client()))
}

func (p *Provider) GitAuth() *http.BasicAuth {
	return p.gitAuth
}

func (p *Provider) IsNetwrkgroupsDisabled() bool {
	return p.isNetwrkgroupsDisabled
}
