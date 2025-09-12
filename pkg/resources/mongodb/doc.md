# Mongodb

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

## Argument Reference

### Generic arguments

* `name` - (Required) Name of the MongoDB.
* `region` - (Optional) Geographical regio
* `plan` - (Optional) Plan size. Default: `base`

### Specific arguments

None.

## Attribute Reference

* `id` - Generated unique identifier.
* `name` - Name of the instance.
* `host` - MongoDB host
* `port` - MongoDB port
* `user` - MongoDB username
* `password` - MongoDB password
