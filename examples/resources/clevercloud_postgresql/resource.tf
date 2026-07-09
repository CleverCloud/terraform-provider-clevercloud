resource "clevercloud_postgresql" "postgresql_database" {
  name   = "postgresql_database"
  plan   = "dev"
  region = "par"

  # Merged with the provider-level `default_tags`; the effective set is exposed as `tags_all`.
  tags = ["team:data", "env:prod"]
}