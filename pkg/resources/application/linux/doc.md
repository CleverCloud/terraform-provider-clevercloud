Manage [Linux](https://www.clever-cloud.com/developers/doc/applications/linux/) applications.

See [Linux runtime](https://www.clever-cloud.com/developers/doc/applications/linux/).

## Example usage

### Basic

```terraform
resource "clevercloud_linux" "myapp" {
	name = "tf-myapp"
	region = "par"
	min_instance_count = 1
	max_instance_count = 2
	smallest_flavor = "XS"
	biggest_flavor = "M"
	run_command = "./start.sh"
}
```

### Advanced

```terraform
resource "clevercloud_linux" "myapp" {
    name = "tf-myapp"
    region = "par"
    min_instance_count = 1
    max_instance_count = 2
    smallest_flavor = "XS"
    biggest_flavor = "M"
    run_command = "./start.sh"
    build_command = "make build"
    makefile = "Makefile.custom"
    mise_file_path = "./tools/mise.toml"
    disable_mise = true
    dependencies = [
        "addon_bcc1d486-90f2-4e89-892d-38dbd8f7bc32"
    ]
    deployment {
        repository = "https://github.com/..."
    }
}
```
