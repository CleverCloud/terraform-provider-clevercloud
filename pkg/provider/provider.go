package provider

import (
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"go.clever-cloud.dev/client"
)

type Provider interface {
	Organization() string
	Client() *client.Client
	GitAuth() *http.BasicAuth
	IsNetwrkgroupsDisabled() bool
	// DefaultTags returns the tags configured at provider level, which are merged
	// into every taggable resource's own tags.
	DefaultTags() []string
}
