package provider

import (
	"sync"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"go.clever-cloud.dev/client"
)

type Provider struct {
	cc           *client.Client
	Organisation string
	configured   sync.Once
}

func New() tfsdk.Provider {
	return &Provider{}
}
