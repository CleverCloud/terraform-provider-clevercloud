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

  # Tags applied to every taggable resource, merged with each resource's own `tags`.
  default_tags = ["managed-by:terraform", "env:demo"]
}

resource "clevercloud_postgresql" "PG1" {
  name = "PG1"
  plan = "dev"
  region = "par"

  # Merged with the provider `default_tags`; the effective set is exposed as `tags_all`.
  tags = ["team:data"]
}

resource "clevercloud_nodejs" "node1" {
  name = "myNodeApp"
	region = "par"
	min_instance_count = 1
	max_instance_count = 2
	smallest_flavor = "XS"
	biggest_flavor = "M"
	node_version = "14.0"
	tags = ["team:backend"]
}

resource "clevercloud_cellar" "cellar1" {
  name = "cellar1"
  region = "par"

  # A resource with no `tags` still receives the provider `default_tags` via `tags_all`.
}

resource "clevercloud_cellar_bucket" "bucket1" {
  id = "bucket1"
  cellar_id = clevercloud_cellar.cellar1.id
}

output "host" {
  value = clevercloud_postgresql.PG1.host
}

output "ID" {
  value = clevercloud_postgresql.PG1.id
}

# Effective tags: the resource `tags` merged with the provider `default_tags`.
# => ["env:demo", "managed-by:terraform", "team:data"]
output "postgresql_tags_all" {
  value = clevercloud_postgresql.PG1.tags_all
}
