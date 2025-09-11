# Nodejs

Manage [Node.js](https://nodejs.org/) & [Bun](https://bun.sh/) applications.

See [Node.js & Bun product specification](https://www.clever.cloud/developers/doc/applications/nodejs/).

## Example usage

### Basic

```terraform
resource "clevercloud_nodejs" "my_nodejs" {
    name = "my nodejs"
    region = "par
    max_instance_count = 1
    min_instance_count = 2
    smallest_flavor = "M"
    biggest_flavor = "L"
}
```

### Advanced

```terraform
resource "clevercloud_nodejs" "my_nodejs2" {
    name = "my nodejs in ${var.region}"
    region = var.region
    max_instance_count = var.app_max_instances
    min_instance_count = var.app_min_instances
    smallest_flavor = var.app_nodejs_flavour
    biggest_flavor = var.app_nodejs_flavour
    dependencies = [
        clevercloud_postgresql.postgresql_database.id,
        clevercloud_mysql.mysql_database.id,
        clevercloud_redis.redis_database.id,
        clevercloud_mongodb.mongodb_database.id
    ]
    environment = {
        "CC_NODE_BUILD_TOOL": "yarn-berry"
        "NPM_TOKEN": "00000000-0000-0000-0000-000000000000"
    }
}
```
