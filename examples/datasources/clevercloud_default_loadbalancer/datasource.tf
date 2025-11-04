# Example: Using the default load balancer datasource

# Configure the Clever Cloud provider
provider "clevercloud" {
  # organisation is set via ORGANISATION environment variable
  # or can be set here: organisation = "orga_xxxx"
}

# Create a Docker application
resource "clevercloud_docker" "example" {
  name               = "my-app"
  region             = "par"
  min_instance_count = 1
  max_instance_count = 2
  smallest_flavor    = "XS"
  biggest_flavor     = "M"
}

# Fetch the default load balancer information for the application
data "clevercloud_default_loadbalancer" "example" {
  application_id = clevercloud_docker.example.id
}

# Output the load balancer DNS information
output "loadbalancer_cname" {
  description = "CNAME record for the load balancer"
  value       = data.clevercloud_default_loadbalancer.example.cname
}

output "loadbalancer_ips" {
  description = "List of IP addresses for the load balancer"
  value       = data.clevercloud_default_loadbalancer.example.servers
}

output "loadbalancer_name" {
  description = "Name/region of the load balancer"
  value       = data.clevercloud_default_loadbalancer.example.name
}

# Example: Using the load balancer IPs in a DNS configuration
# This could be used with a DNS provider like Cloudflare, AWS Route53, etc.
#
# resource "cloudflare_record" "app_a_records" {
#   count   = length(data.clevercloud_default_loadbalancer.example.servers)
#   zone_id = var.cloudflare_zone_id
#   name    = "myapp"
#   value   = data.clevercloud_default_loadbalancer.example.servers[count.index]
#   type    = "A"
#   ttl     = 300
# }
