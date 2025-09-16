# Metabase

Manage [Metabase](https://www.metabase.com/) product.

See [Metabase product specification](https://www.clever.cloud/developers/doc/addons/metabase/).

## Example usage

### Basic

```terraform
resource "clevercloud_metabase" "my_metabase" {
    name = "my-metabase"
    region = "par"
    plan = "base"
}
```
