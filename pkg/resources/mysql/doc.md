# Mysql

Manage [Mysql](https://www.mysql.org/) product.

See [product specification](https://www.clever.cloud/developers/doc/addons/mysql/).

## Example usage

### Basic

```terraform
resource "clevercloud_mysql" "my_mysql" {
    name = "my-mysql"
    plan = "base"
    region = "par"
}
```

### Advanced

```terraform
resource "clevercloud_mysql" "my_mysql2" {
    name = "my-mysql-2"
    plan = "XXL"
    region = "par"
    version = "8.4"
    backup = true
}
```
