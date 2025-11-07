Manage Static with Apache applications.

See [Static with Apache product specification](https://www.clever.cloud/developers/doc/deploy/application/static/).

## Example usage

### Basic

```terraform
resource "clevercloud_static" "myapp" {
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
resource "clevercloud_static" "myapp" {
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

**Note**: For deploying from private GitHub repositories, see the [private repository deployment guide](https://registry.terraform.io/providers/CleverCloud/clevercloud/latest/docs#applications-private-repository-deployment).
