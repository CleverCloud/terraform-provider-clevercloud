Manage [Mysql](https://www.mysql.org/) product.

See [product specification](https://www.clever.cloud/developers/doc/addons/mysql/).


## Argument Reference

### Generic arguments

* `name` - (Required) Name of the Mysql.
* `region` - (Optional) Geographical region where the data will be stored. Defaults to `par`.
* `smallest_flavor` (String) Smallest instance flavor
* `biggest_flavor` (String) Biggest instance flavor, if different from smallest, enable auto-scaling
* `max_instance_count` (Number) Maximum instance count, if different from min value, enable auto-scaling
* `min_instance_count` (Number) Minimum instance count
* `env` - (Optional) Environment variables.
* `dependencies` - (Optional) Addon IDs to link to.
* `vhosts` - (Optional) Custom domain names. If empty, a test domain name will be generated.
* `build_flavor` - (Optional) If set, use a build instance of the size provided.

### Specific arguments

None.

## Attribute Reference

* `id` - Generated unique identifier.
* `name` - Name of the instance.
* `deploy_url` - Git url for deployments.
