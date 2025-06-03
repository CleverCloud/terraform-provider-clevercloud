# Manage [Play2](https://www.playframework.com/) applications.

See [Play2 product](https://www.clever-cloud.com/developers/guides/play-framework-2/) specification.

## Example usage

### Basic

```terraform
resource "clevercloud_play2" "myapp" {
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
resource "clevercloud_play2" "myapp" {
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
