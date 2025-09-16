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
