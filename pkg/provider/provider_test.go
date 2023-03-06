package provider

import (
	_ "embed"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// Common variables used by different acceptance tests

//go:embed provider_test_block.tf
var providerBlock string

var protoV6Provider = map[string]func() (tfprotov6.ProviderServer, error){
	"clevercloud": providerserver.NewProtocol6WithError(New("test")()),
}
