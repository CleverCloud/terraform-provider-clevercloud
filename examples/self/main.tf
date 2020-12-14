terraform {
  required_providers {
    clevercloud = {
      source = "hashicorp.com/gaelreyrol/clevercloud"
      version = "0.1.0"
    }
  }
}

data "clevercloud_self" "current" {}

# Returns self
output "current" {
  value = data.clevercloud_self.current
}
