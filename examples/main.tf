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

resource "clevercloud_postgresql" "PG1" {
  name = "PG1"
  plan = "dev"
  region = "par"
}

resource "clevercloud_nodejs" "node1" {
  name = "myNodeApp"
	region = "par"
	min_instance_count = 1
	max_instance_count = 2
	smallest_flavor = "XS"
	biggest_flavor = "M"
}

resource "clevercloud_cellar" "cellar1" {
  name = "cellar1"
  region = "par"
}

resource "clevercloud_cellar_bucket" "bucket1" {
  id = "bucket1"
  cellar_id = cellar1.id
}

output "host" {
  value = clevercloud_postgresql.PG1.host
}

output "ID" {
  value = clevercloud_postgresql.PG1.id
}
