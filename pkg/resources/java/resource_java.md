# Manage [Java](https://www.java.com/en/) applications.

See [Java product](https://www.clever-cloud.com/doc/getting-started/by-language/java/) specification.

## Example usage

```terraform
resource "clevercloud_java_war" "myapp" {
	name = "tf-myapp"
	region = "par"
	min_instance_count = 1
	max_instance_count = 2
	smallest_flavor = "XS"
	biggest_flavor = "M"
    deployment {
        repository = "https://github.com/..."
    }
}
```
