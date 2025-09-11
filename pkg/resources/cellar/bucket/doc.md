# Cellar

Manage [Cellar Buckets](https://www.clever.cloud/developers/doc/addons/cellar/).

See [Cellar product specification](https://www.clever.cloud/developers/doc/addons/cellar/).


## Example usage

A `clevercloud_cellar` resource is needed to use a `clevercloud_cellar_bucket`.

### Basic

```terraform
resource "clevercloud_cellar_bucket" "my_bucket" {
    name = "my bucket"
    cellar_id = "cellar_b6643eb2-1e55-4f87-aa0a-b24882cebfef"
}
```

```terraform
resource "clevercloud_cellar_bucket" "my_bucket" {
    name = "my bucket"
    cellar_id = clevercloud_cellar.my_cellar.id
}
```
