Retrieves information about the default load balancer for a Clever Cloud application.

It help user configure DNS entries to reach their applications

## Example Usage

```hcl
data "clevercloud_default_loadbalancer" "example" {
  application_id  = "app_2b29643f-ae97-4de8-95da-795b009469e5"
}

# Use the load balancer information
output "lb_cname" {
  value = data.clevercloud_default_loadbalancer.example.cname
}

output "lb_ips" {
  value = data.clevercloud_default_loadbalancer.example.servers
}
```
