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
