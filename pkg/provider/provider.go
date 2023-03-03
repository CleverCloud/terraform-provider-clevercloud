package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"go.clever-cloud.dev/client"
)

type Provider struct {
	version      string
	cc           *client.Client
	Organisation string
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &Provider{version: version}
	}
}
