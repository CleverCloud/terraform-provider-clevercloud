Manage [Keycloak](https://www.keycloak.org/) product.

See [Keycloak product specification](https://www.clever-cloud.com/developers/doc/addons/keycloak/).

## Realms Management

The `realms` attribute allows you to create additional Keycloak realms beyond the default `master` realm.

**Important notes:**

- Realms can only be **added**, not removed once created
- Realm names must contain only alphanumeric characters, underscores and hyphens
- When you add new realms to an existing Keycloak instance, the underlying Java application will be restarted automatically
- Removing realms from the configuration will not delete them from Keycloak, they will remain in the state

**Example:**

```hcl
resource "clevercloud_keycloak" "example" {
  name   = "my-keycloak"
  region = "par"
  realms = ["production", "staging", "dev"]
}
```
