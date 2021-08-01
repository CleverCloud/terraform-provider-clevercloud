terraform {
  required_providers {
    clevercloud = {
      source = "local/clevercloud/clevercloud"
      version = "0.1.0"
    }
  }
}

variable "clevercloud_token" {
  type = string
}

variable "clevercloud_secret" {
  type = string
}

variable "clevercloud_consumer_key" {
  type = string
}

variable "clevercloud_consumer_secret" {
  type = string
}

provider "clevercloud" {
  token = var.clevercloud_token
  secret = var.clevercloud_secret
  consumer_key = var.clevercloud_consumer_key
  consumer_secret = var.clevercloud_consumer_secret
}

data "clevercloud_self" "current" {}
# data "clevercloud_zones" "available" {}
# data "clevercloud_flavors" "available" {}
# data "clevercloud_addon_providers" "available" {}
# data "clevercloud_instances" "available" {}

# data "clevercloud_application" "test" {
#   id = "app_6a43221f-f0a9-4e10-98d6-5fbe6c7cb296"
# }

# data "clevercloud_addon" "config_test" {
#   id = "addon_b2c1423b-05b7-417e-af20-ef2a2c77cc77"
# }

# data "clevercloud_addon" "cellar_test" {
#   id = "addon_aea9c52e-4d5b-49a9-90ab-0b510fdd7b86"
# }

# data "clevercloud_addon" "postgresql_test" {
#   id = "addon_a3c22ce4-6ce1-4586-bbe3-2b76ea015352"
# }

output "self_current" {
  value = data.clevercloud_self.current
}

# output "available_zones" {
#   value = data.clevercloud_zones.available
# }

# output "available_flavors" {
#   value = data.clevercloud_flavors.available
# }

# output "available_addons" {
#   value = data.clevercloud_addon_providers.available
# }

# output "available_instances" {
#   value = data.clevercloud_instances.available
# }

# output "test_application" {
#   value = data.clevercloud_application.test
# }

# output "test_config_addon" {
#   value = data.clevercloud_addon.config_test
# }

# output "test_cellar_addon" {
#   value = data.clevercloud_addon.cellar_test
# }

# output "test_postgresql_addon" {
#   value = data.clevercloud_addon.postgresql_test
# }
