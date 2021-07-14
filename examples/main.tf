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
data "clevercloud_zones" "available" {}
data "clevercloud_flavors" "available" {}

data "clevercloud_application" "cc_deploy_button" {
  id = "app_ab73990b-79b0-49f4-aeb6-76aa72b780a0"
}

output "self_current" {
  value = data.clevercloud_self.current
}

output "available_zones" {
  value = data.clevercloud_zones.available
}

output "available_flavors" {
  value = data.clevercloud_flavors.available
}

output "deploy_button" {
  value = data.clevercloud_application.cc_deploy_button
}
