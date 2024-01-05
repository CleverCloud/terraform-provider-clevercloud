resource "clevercloud_postgresql" "postgresql_database" {
  name   = "postgresql_database"
  plan   = "dev"
  region = "par"
}