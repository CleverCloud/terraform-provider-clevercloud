package provider

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const orgBlock = `
data "%s" "%s" {}
`

func TestAccOrganisation_basic(t *testing.T) {
	org := os.Getenv("ORGANISATION")
	fullName := "data.clevercloud_organisation.org"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			if org == "" {
				t.Fatalf("missing ORGANISATION env var")
			}
		},
		ProtoV6ProviderFactories:  protoV6Provider,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{{
			Config: fmt.Sprintf(providerBlock, org) +
				fmt.Sprintf(orgBlock, "clevercloud_organisation", "org"),
			Check: resource.ComposeTestCheckFunc(
				resource.TestMatchResourceAttr(fullName, "id", regexp.MustCompile(`^(user|org)_.*`)),
				resource.TestCheckResourceAttrSet(fullName, "name"),
			),
		}},
	})
}
