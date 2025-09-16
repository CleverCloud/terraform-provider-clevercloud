# Network Group

Manage any [Network Group](https://www.clever.cloud/developers/doc/develop/network-groups/).

See [Network Groups product specification](https://www.clever.cloud/developers/doc/develop/network-groups/).

## Example usage

### Basic

```
resource "clevercloud_networkgroup" "my_networkgroup" {
    name = "my_ng"
}
```

### Advanced

```
resource "clevercloud_networkgroup" "my_networkgroup2" {
    name = "my_ng"
    description = "My networkgroup"
    tags = [
        "secure",
        "myng",
        "databases",
    ]
}
```
