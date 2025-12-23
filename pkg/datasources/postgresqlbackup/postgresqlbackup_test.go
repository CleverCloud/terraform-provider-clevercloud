package postgresqlbackup_test

import (
	"regexp"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/tests"
)

const (
	// This is a dedicated PostgreSQL addon kept for testing purposes only
	// It has backups created automatically (24+ hours old)
	testPostgreSQLID = "postgresql_71ee0b87-de5d-4e7a-a079-d123e87bf67d"
)

func TestAccDataSourcePostgreSQLBackup_latest(t *testing.T) {
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)

	dataBlock := helper.NewDataRessource(
		"clevercloud_postgresql_backup",
		"latest",
		helper.SetKeyValues(map[string]any{
			"postgresql_id": testPostgreSQLID,
			"selector":      "latest",
		}))

	config := providerBlock.Append(dataBlock).String()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		PreCheck:                 tests.ExpectOrganisation(t),
		Steps: []resource.TestStep{{
			ResourceName: "clevercloud_postgresql_backup.by_date",
			Config:       config,
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(
					"data.clevercloud_postgresql_backup.latest",
					tfjsonpath.New("id"),
					knownvalue.StringRegexp(regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)),
				),
				statecheck.ExpectKnownValue(
					"data.clevercloud_postgresql_backup.latest",
					tfjsonpath.New("download_url"),
					knownvalue.StringRegexp(regexp.MustCompile(`^https://`)),
				),
				statecheck.ExpectKnownValue(
					"data.clevercloud_postgresql_backup.latest",
					tfjsonpath.New("creation_date"),
					knownvalue.StringRegexp(regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}`)),
				),
				statecheck.ExpectKnownValue(
					"data.clevercloud_postgresql_backup.latest",
					tfjsonpath.New("deletion_date"),
					knownvalue.StringRegexp(regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}`)),
				),
			},
		}},
	})
}

func TestAccDataSourcePostgreSQLBackup_byDate(t *testing.T) {
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)

	dataBlock := helper.NewDataRessource("clevercloud_postgresql_backup", "by_date",
		helper.SetKeyValues(map[string]any{
			"postgresql_id": testPostgreSQLID,
			"selector":      time.Now().Add(-48 * time.Hour),
		}))

	config := providerBlock.Append(dataBlock).String()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		PreCheck:                 tests.ExpectOrganisation(t),
		Steps: []resource.TestStep{{
			ResourceName: "clevercloud_postgresql_backup.by_date",
			Config:       config,
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(
					"data.clevercloud_postgresql_backup.by_date",
					tfjsonpath.New("id"),
					knownvalue.StringRegexp(regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)),
				),
				statecheck.ExpectKnownValue(
					"data.clevercloud_postgresql_backup.by_date",
					tfjsonpath.New("download_url"),
					knownvalue.StringRegexp(regexp.MustCompile(`^https://`)),
				),
				statecheck.ExpectKnownValue(
					"data.clevercloud_postgresql_backup.by_date",
					tfjsonpath.New("creation_date"),
					knownvalue.StringRegexp(regexp.MustCompile(`^2025-12-2[0-3]T`)), // Should be 20, 21, 22, or 23
				),
			},
		},
		},
	})
}

func TestAccDataSourcePostgreSQLBackup_byUUID(t *testing.T) {
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)

	dataBlock := helper.NewDataRessource("clevercloud_postgresql_backup", "by_uuid",
		helper.SetKeyValues(map[string]any{
			"postgresql_id": testPostgreSQLID,
			"selector":      "596538c3-286c-474b-aa90-c09832ce13f1",
		}))

	config := providerBlock.Append(dataBlock).String()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		PreCheck:                 tests.ExpectOrganisation(t),
		Steps: []resource.TestStep{{
			ResourceName: "clevercloud_postgresql_backup.by_uuid",
			Config:       config,
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(
					"data.clevercloud_postgresql_backup.by_uuid",
					tfjsonpath.New("id"),
					knownvalue.StringRegexp(regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)),
				),
				statecheck.ExpectKnownValue(
					"data.clevercloud_postgresql_backup.by_uuid",
					tfjsonpath.New("download_url"),
					knownvalue.StringRegexp(regexp.MustCompile(`^https://`)),
				),
				statecheck.ExpectKnownValue(
					"data.clevercloud_postgresql_backup.by_uuid",
					tfjsonpath.New("creation_date"),
					knownvalue.StringRegexp(regexp.MustCompile(`^2025-12-2[0-3]T`)), // Should be 20, 21, 22, or 23
				),
			},
		},
		},
	})
}

func TestAccDataSourcePostgreSQLBackup_notFound(t *testing.T) {
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)

	dataBlock := helper.NewDataRessource(
		"clevercloud_postgresql_backup",
		"not_found",
		helper.SetKeyValues(map[string]any{
			"postgresql_id": testPostgreSQLID,
			"selector":      "2020-01-01T00:00:00Z",
		}))

	config := providerBlock.Append(dataBlock).String()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		PreCheck:                 tests.ExpectOrganisation(t),
		Steps: []resource.TestStep{{
			ResourceName: "clevercloud_postgresql_backup.by_date",
			Config:       config,
			ExpectError:  regexp.MustCompile(`No backup found`),
		}},
	})
}
