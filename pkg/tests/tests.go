package tests

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"go.clever-cloud.com/terraform-provider/pkg/provider/impl"
)

// ProtoV6Provider builds a fresh provider instance per factory call so parallel tests
// don't share mutable provider configuration (e.g. default_tags). Baking a single
// impl.New("test")() instance here would let one test's provider config leak into another.
var ProtoV6Provider = map[string]func() (tfprotov6.ProviderServer, error){
	"clevercloud": func() (tfprotov6.ProviderServer, error) {
		return providerserver.NewProtocol6WithError(impl.New("test")())()
	},
}

var ORGANISATION = os.Getenv("ORGANISATION")

func ExpectOrganisation(t *testing.T) func() {
	return func() {
		if ORGANISATION == "" {
			t.Fatalf("ORGANISATION env var is not set")
		}
	}
}
