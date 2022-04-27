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

output "host" {
  value = clevercloud_postgresql.PG1.host
}

output "ID" {
  value = clevercloud_postgresql.PG1.id
}
