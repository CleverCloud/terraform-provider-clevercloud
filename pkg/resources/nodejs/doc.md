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

## Argument Reference

### Generic arguments

* `name` - (Required) Name of the Nodejs.
* `region` - (Optional) Geographical region where the data will be stored. Defaults to `par`.
* `smallest_flavor` (String) Smallest instance flavor
* `biggest_flavor` (String) Biggest instance flavor, if different from smallest, enable auto-scaling
* `max_instance_count` (Number) Maximum instance count, if different from min value, enable auto-scaling
* `min_instance_count` (Number) Minimum instance count
* `env` - (Optional) Environment variables.
* `dependencies` - (Optional) Addon IDs to link to.
* `vhosts` - (Optional) Custom domain names. If empty, a test domain name will be generated.
* `build_flavor` - (Optional) If set, use a build instance of the size provided.
* `sticky_sessions` - (Optional) If set to true, when horizontal scalability is enabled, a user is always served by the same scaler. Some frameworks or technologies require this option. Default: false
* `redirect_https` - (Optional) If set to true, any non secured HTTP request to this application will be redirected to HTTPS with a 301 Moved Permanently status code. Default: true
* `cancel_on_push` - (Optional) A "git push" will cancel any ongoing deployment and start a new one with the last available commit.


### Specific arguments

None.

## Attribute Reference

* `id` - Generated unique identifier.
* `name` - Name of the instance.
* `deploy_url` - Git url for deployments.
