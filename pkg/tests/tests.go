package tests

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"go.clever-cloud.com/terraform-provider/pkg/provider/impl"
)

var ProtoV6Provider = map[string]func() (tfprotov6.ProviderServer, error){
	"clevercloud": providerserver.NewProtocol6WithError(impl.New("test")()),
}

var ORGANISATION = os.Getenv("ORGANISATION")

func ExpectOrganisation(t *testing.T) func() {
	return func() {
		if ORGANISATION == "" {
			t.Fatalf("ORGANISATION env var is not set")
		}
	}
}
