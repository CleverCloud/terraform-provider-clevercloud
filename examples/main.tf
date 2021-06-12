terraform {
  required_providers {
    clevercloud = {
      source = "clevercloud/clevercloud"
      version = "0.1.0"
    }
  }
}

provider "clevercloud" {}

data "clevercloud_self" "current" {}
data "clevercloud_zones" "available" {}
data "clevercloud_flavors" "available" {}

output "self_current" {
  value = data.clevercloud_self.current
}

output "available_zones" {
  value = data.clevercloud_zones.available
}

output "available_flavors" {
  value = data.clevercloud_flavors.available
}
