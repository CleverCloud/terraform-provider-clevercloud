terraform {
  required_providers {
    clevercloud = {
      source = "hashicorp.com/gaelreyrol/clevercloud"
      version = "0.1.0"
    }
  }
}

provider "clevercloud" {}

module "self" {
  source = "./self"
}

output "current_self" {
  value = module.self.current
}
