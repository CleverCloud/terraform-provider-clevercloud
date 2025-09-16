# Redis

Manage [Redis](https://redis.io/) product.

See [Redis product specification](https://www.clever.cloud/developers/doc/addons/redis/).

## Example usage

### Basic

```terraform
resource "clevercloud_redis" "my_redis" {
    name = "my-redis"
    plan = "base"
    region = "par"
}
```
