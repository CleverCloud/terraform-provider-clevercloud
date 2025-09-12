# Materia KV

Manage [Materia KV](https://www.clever-cloud.com/materia/materia-kv/) product.

See [Materia KV product specification](https://www.clever.cloud/developers/doc/addons/materia-kv/).

## Example usage

### Basic

```terraform
resource "clevercloud_materia_kv" "my_materiakv" {
  name = "my-materiakv"
  region = "par"
  plan = "base"
}
```

## Argument Reference

* `name` - (Required) Name of the Materia KV instance.
* `region` - (Optional) Geographical region where the data will be stored. Defaults to `par`.

## Attribute Reference

* `id` - Generated unique identifier.
* `name` - Name of the instance.
* `host` - Materia KV endpoint.
* `port` - Materia KV port.
* `token` - Materia KV authentication token.
