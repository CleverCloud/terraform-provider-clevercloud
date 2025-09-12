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


## Argument Reference

### Generic arguments

* `name` - (Required) Name of the docker.
* `region` - (Optional) Geographical region where the data will be stored. Defaults to `par`.
* `smallest_flavor` (String) Smallest instance flavor
* `biggest_flavor` (String) Biggest instance flavor, if different from smallest, enable auto-scaling
* `max_instance_count` (Number) Maximum instance count, if different from min value, enable auto-scaling
* `min_instance_count` (Number) Minimum instance count
* `env` - (Optional) Environment variables.
* `dependencies` - (Optional) Addon IDs to link to.
* `vhosts` - (Optional) Custom domain names. If empty, a test domain name will be generated.
* `sticky_sessions` - (Optional) If set to true, when horizontal scalability is enabled, a user is always served by the same scaler. Some frameworks or technologies require this option. Default: false
* `redirect_https` - (Optional) If set to true, any non secured HTTP request to this application will be redirected to HTTPS with a 301 Moved Permanently status code. Default: true

### Specific arguments

* `dockerfile` - (Optional) The name of the Dockerfile to build. Shortcut for `CC_DOCKERFILE`
* `container_port` - (Optional) Set to custom HTTP port if your Docker container runs on custom port. Shortcut for `CC_DOCKER_EXPOSED_HTTP_PORT`.
* `container_port_tcp` - (Optional) Set to custom TCP port if your Docker container runs on custom port. Shortcut for `CC_DOCKER_EXPOSED_TCP_PORT`
* `ipv6_cidr` - (Optional) Activate the support of IPv6 with an IPv6 subnet in the docker daemon. Shortcut for `CC_DOCKER_FIXED_CIDR_V6`
* `registry_url` - (Optional) The server of your private registry. Shortcut for `CC_DOCKER_LOGIN_SERVER`
* `registry_user` - (Optional) The username to login to a private registry. Shortcut for `CC_DOCKER_LOGIN_USERNAME`
* `registry_password` - (Optional) The password of your username. Shortcut for `CC_DOCKER_LOGIN_PASSWORD`
* `daemon_socket_mount` - (Optional) Set to true to access the host Docker socket from inside your container. Shortcut for `CC_MOUNT_DOCKER_SOCKET`


## Attribute Reference

* `id` - Generated unique identifier.
* `name` - Name of the instance.
* `deploy_url` - Git url for deployments.
