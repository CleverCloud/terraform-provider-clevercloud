# PHP

Manage [PHP with Apache](https://www.php.net/) applications.

See [PHP with Apache product specification](https://www.clever.cloud/developers/doc/applications/php/).

Note: It is currently not possible to deploy through FTP when deploying with terraform.

## Example usage

### Basic

```terraform
resource "clevercloud_php" "my_php" {
    name = "tf-php"
    region = "par"
    min_instance_count = 1
    max_instance_count = 2
    smallest_flavor = "XS"
    biggest_flavor = "M"
}
```

### Advanced

```terraform
resource "clevercloud_php" "my_php2" {
    name = "tf-frankenphp2"
    description = "My website example.com in php"
    region = "par"
    min_instance_count = 1
    max_instance_count = 2
    smallest_flavor = "XS"
    biggest_flavor = "M"
    dependencies = [
        "addon_bcc1d486-90f2-4e89-892d-38dbd8f7bc32"
    ]
    deployment {
        repository = "https://github.com/..."
    }
    php_version = "8.4"
    redirect_https = true
    vhosts = [
        "www.example.com",
        "example.com",
    ]
}
```
