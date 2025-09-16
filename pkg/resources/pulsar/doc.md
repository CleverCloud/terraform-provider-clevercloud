# Pulsar

Manage [Pulsar](https://www.pulsar.org/) product.

See [Pulsar product specification](https://www.clever.cloud/developers/doc/addons/pulsar/).

## Example usage

### Basic

```terraform
resource "clevercloud_pulsar" "my_pulsar" {
    name = "my-pulsar"
    region = "par"
}
```
