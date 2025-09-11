# Postgresql

Manage [PostgreSQL](https://www.postgresql.org/) product.

See [product specification](https://www.clever.cloud/developers/doc/addons/postgresql/).

## Example usage

### Basic

```terraform
resource "clevercloud_postgresql" "postgresql_database" {
    name   = "postgresql_database"
    plan   = "dev"
    region = "par"
}
```

### Advanced

```terraform
resource "clevercloud_postgresql" "postgresql_database" {
    name   = "postgresql_database"
    plan   = "dev"
    region = "par"
}
```
