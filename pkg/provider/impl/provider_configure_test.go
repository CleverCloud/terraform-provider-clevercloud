package impl

import (
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"regexp"
	"testing"
)

var protoV6Provider = map[string]func() (tfprotov6.ProviderServer, error){
	"clevercloud": providerserver.NewProtocol6WithError(New("test")()),
}

func TestProvider_ConfigureEmptyOrganisation(t *testing.T) {
	providerBlock := `
provider "clevercloud" {}

// dummy resource to trigger provider configuration
resource "clevercloud_cellar" "test" {
  name = "test"
}
`
	expectedError := regexp.MustCompile("Organisation should be set")

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: protoV6Provider,
		Steps: []resource.TestStep{{
			Config: providerBlock,
			SkipFunc: func() (bool, error) {
				return false, nil
			},
			ExpectError: expectedError,
		}},
	})
}
