# Keycloak

Manage [Keycloak](https://www.keycloak.org/) product.

See [Keycloak product specification](https://www.clever.cloud/developers/doc/addons/keycloak/).

## Example usage

### Basic

```terraform
resource "clevercloud_keycloak" "my_keycloak" {
  name = "my-keycloak"
  region = "par"
  plan = "base"
}
```

## Argument Reference

* `name` - (Required) Name of the Keycloak.
* `region` - (Optional) Geographical region where the data will be stored. Defaults to `par`.
* `plan` - (Optional) The plan size. Defaults to `base`.

## Attribute Reference

* `id` - Generated unique identifier.
* `name` - Name of the instance.
* `host` - Keycloak endpoint.
