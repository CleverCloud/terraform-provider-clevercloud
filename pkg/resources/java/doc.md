# Java

Manage [Java](https://www.java.com/en/) applications.

See [Java product specification](https://www.clever.cloud/developers/doc/applications/java/).

## Example usage

### Basic

```terraform
resource "clevercloud_java_war" "myapp" {
    name = "tf-myapp"
    region = "par"
    min_instance_count = 1
    max_instance_count = 2
    smallest_flavor = "XS"
    biggest_flavor = "M"
}
```

### Advanced

```terraform
resource "clevercloud_java_war" "myapp" {
    name = "tf-myapp"
    region = "par"
    min_instance_count = 1
    max_instance_count = 2
    smallest_flavor = "XS"
    biggest_flavor = "M"
    dependencies = [
        "addon_bcc1d486-90f2-4e89-892d-38dbd8f7bc32"
    ]
    deployment {
        repository = "https://github.com/..."
    }
}
```


## Argument Reference

### Generic arguments

* `name` - (Required) Name of the Java.
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

None.

## Attribute Reference

* `id` - Generated unique identifier.
* `name` - Name of the instance.
* `deploy_url` - Git url for deployments.
