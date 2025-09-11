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
