Manage Static with Apache applications.

Runs your static files behind an Apache HTTPD server. Supports `.htaccess`, rewrite rules, and other Apache modules. For a lighter runtime without Apache (ideal for sites built by Hugo, Jekyll, Astro, Vite, Next.js export, MkDocs...), use `clevercloud_static` instead.

See [Static with Apache product specification](https://www.clever.cloud/developers/doc/deploy/application/static/).

## Migrating from `clevercloud_static` (before v1.12.0)

In releases before v1.12.0, the `clevercloud_static` resource actually managed Static-with-Apache apps. It has been renamed to `clevercloud_static_apache` so the `clevercloud_static` name can be reused for the lightweight Static runtime.

Existing users should add a `moved` block to their configuration so Terraform re-tags the state without recreating the app:

```terraform
moved {
  from = clevercloud_static.myapp
  to   = clevercloud_static_apache.myapp
}

resource "clevercloud_static_apache" "myapp" {
  # same configuration as before
}
```

Run `terraform plan` to confirm `No changes`, then `terraform apply`. No downtime, no redeploy.

## Example usage

### Basic

```terraform
resource "clevercloud_static_apache" "myapp" {
	name = "tf-myapp"
	region = "par"
	min_instance_count = 1
	max_instance_count = 2
	smallest_flavor = "XS"
	biggest_flavor = "M"
}
```

### Advanced

```terraform
resource "clevercloud_static_apache" "myapp" {
    name = "tf-myapp"
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
}
```

**Note**: For deploying from private GitHub repositories, see the [private repository deployment guide](https://registry.terraform.io/providers/CleverCloud/clevercloud/latest/docs#applications-private-repository-deployment).
