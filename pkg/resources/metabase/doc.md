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

## Argument Reference

### Generic arguments

* `name` - (Required) Name of the Metabase.
* `region` - (Optional) Geographical region where the data will be stored. Defaults to `par`.
* `plan` - (Optional) The plan size. Defaults to `base`.

### Specific arguments

None.

## Attribute Reference

* `id` - Generated unique identifier.
* `name` - Name of the instance.
* `host` - Metabase URL.
