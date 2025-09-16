# FrankenPHP

FrankenPHP is a modern PHP application server, written in Go. It gives superpowers to your PHP apps thanks to its stunning features: Early Hints, worker mode, real-time capabilities, automatic HTTPS, HTTP/2, and HTTP/3 support...

## Links

- [FrankenPHP Official Website](https://frankenphp.dev/)
- [CleverCloud FrankenPHP Documentation](https://www.clever.cloud/developers/doc/applications/frankenphp/)

## Example usage

### Basic

```terraform
resource "clevercloud_frankenphp" "my_frankenphp" {
    name = "tf-frankenphp"
    region = "par"
    min_instance_count = 1
    max_instance_count = 2
    smallest_flavor = "XS"
    biggest_flavor = "M"
}
```

### Advanced


```terraform
resource "clevercloud_frankenphp" "my_frankenphp2" {
    name = "tf-frankenphp2"
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
