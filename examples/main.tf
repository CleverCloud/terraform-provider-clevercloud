terraform {
  required_providers {
    clevercloud = {
      source  = "CleverCloud/clevercloud"
    }
  }
  required_version = ">= 1.1.0"
}

variable "organisation" {
  type = string
  nullable = false
}

provider "clevercloud" {
  organisation = var.organisation
}

data "clevercloud_datasource" "myorg" {}

resource "clevercloud_postgresql" "PG1" {
  name = "PG1"
  plan = "dev"
  region = "par"
}

resource "clevercloud_nodejs" "node1" {
  name = "myNodeApp"
  description = ""
	region = "par"
	min_instance_count = 1
	max_instance_count = 2
	smallest_flavor = "XS"
	biggest_flavor = "M"
  environment = {
    PROD = true
  }
  dependencies = [ "${clevercloud_postgresql.PG1.id}" ]

  package_manager = "yarn2"
  dev_dependencies = true
}

output "host" {
  value = clevercloud_postgresql.PG1.host
}

output "ID" {
  value = clevercloud_postgresql.PG1.id
}
