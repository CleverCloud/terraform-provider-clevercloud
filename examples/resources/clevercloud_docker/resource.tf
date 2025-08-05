resource "clevercloud_docker" "docker_instance" {
  name = "docker_instance"
  region = "par"

  # horizontal scaling
  min_instance_count = 1
  max_instance_count = 2

  # vertical scaling
  smallest_flavor = "XS"
  biggest_flavor  = "M"
}
