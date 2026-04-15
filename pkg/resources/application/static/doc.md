Manage Static applications.

Lightweight static-file runtime (no Apache). Ideal for sites built by static site generators (Hugo, Jekyll, Astro, MkDocs) or frontend frameworks exporting a static bundle (Vite, Next.js `output: 'export'`, Nuxt `generate`). For legacy sites needing `.htaccess` or Apache modules, use `clevercloud_static_apache` instead.

See [Static product specification](https://www.clever.cloud/developers/doc/deploy/application/static/).

## Example usage

### Basic

```terraform
resource "clevercloud_static" "myapp" {
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
resource "clevercloud_static" "myapp" {
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
