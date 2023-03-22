package provider

import "go.clever-cloud.dev/client"

type Provider interface {
	Organization() string
	Client() *client.Client
}
