# MongoDB

Manage [MongoDB](https://www.mongodb.com/) product.

See [product specification](https://www.clever.cloud/developers/doc/addons/mongodb/).

## Example usage

### Basic

```terraform
resource "clevercloud_mongodb" "my_mongodb" {
    name = "my-mongodb"
    plan = "base"
    region = "par"
}
```
