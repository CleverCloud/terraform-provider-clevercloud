# FrankenPHP

FrankenPHP is a modern PHP application server, written in Go. It gives superpowers to your PHP apps thanks to its stunning features: Early Hints, worker mode, real-time capabilities, automatic HTTPS, HTTP/2, and HTTP/3 support...

## Links

- [FrankenPHP Official Website](https://frankenphp.dev/)
- [CleverCloud FrankenPHP Documentation](https://www.clever.cloud/developers/doc/applications/frankenphp/)


## Argument Reference

### Generic arguments

* `name` - (Required) Name of the FrankenPHP.
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


### Specific arguments

* `dev_dependencies` - (Optional) Set to true to install dev dependencies. Default: false

## Attribute Reference

* `id` - Generated unique identifier.
* `name` - Name of the instance.
* `deploy_url` - Git url for deployments.
