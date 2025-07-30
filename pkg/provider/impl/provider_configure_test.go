package impl_test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"go.clever-cloud.com/terraform-provider/pkg/tests"
)

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
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		Steps: []resource.TestStep{{
			Config: providerBlock,
			SkipFunc: func() (bool, error) {
				return false, nil
			},
			ExpectError: expectedError,
		}},
	})
}
