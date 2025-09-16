# Docker

Manage [Docker](https://www.docker.com/) applications.

See [Docker product specification](https://www.clever.cloud/developers/doc/applications/docker).

## Example usage

### Basic

```terraform
resource "clevercloud_docker" "my_docker" {
  name = "my-docker"
  region = "par"
  min_instance_count = 1
  max_instance_count = 2
  smallest_flavor = "XS"
  biggest_flavor = "M"
}
```

### Advanced

```terraform
resource "clevercloud_docker" "my_docker2" {
    name = "my-docker2"
    description = "My docker container"
    region = "par"
    min_instance_count = 1
    max_instance_count = 2
    smallest_flavor = "XS"
    biggest_flavor = "M"
    dockerfile = "Dockerfile"
    registry_url = "https://registry.gitlab.example.com"
    registry_user = "registry_username_deployment"
    registry_password = "changeme"
    vhosts = [
        "example.com",
        "www.example.com",
    ]
}
```
