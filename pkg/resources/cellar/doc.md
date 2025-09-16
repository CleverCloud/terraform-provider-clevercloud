# Cellar

Manage [Cellar](https://www.clever.cloud/developers/doc/addons/cellar/) product.

See [Cellar product specification](https://www.clever.cloud/developers/doc/addons/cellar/).

## Example usage

### Basic

```terraform
resource "clevercloud_cellar" "my_cellar" {
    name = "my-cellar"
}
```

### Advanced

```terraform
resource "clevercloud_cellar" "my_cellar2" {
    name = "my-cellar"
    description = "My cellar"
    region = "par
}
```
