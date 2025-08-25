# Clever Cloud Drains Example

This example demonstrates how to configure log drains for Clever Cloud applications to forward logs to various external destinations.

## Available Drain Types

- **HTTP**: Forward logs to any HTTP endpoint
- **Syslog TCP/UDP**: Send logs via standard syslog protocol
- **Datadog**: Forward logs to Datadog log management
- **New Relic**: Send logs to New Relic log platform
- **Elasticsearch**: Index logs in an Elasticsearch cluster
- **OVH**: Send logs to OVH Logs Data Platform

## Usage

1. Configure your Clever Cloud provider with your organization ID
2. Create or reference an existing application
3. Configure drains for your desired destinations
4. Apply the configuration

```bash
terraform init
terraform plan
terraform apply
```

## Configuration Notes

### Sensitive Data
- API keys and passwords should be managed securely using environment variables or Terraform variables
- Consider using `terraform.tfvars` files for sensitive configuration (add to .gitignore)

### Testing Drains
- HTTP drains can be tested using services like httpbin.org or webhook.site
- Syslog drains can be tested with local syslog servers or cloud services
- External service drains require valid accounts and API keys

### Log Types
All drains support three types of logs:
- `LOG`: Application logs (default)
- `ACCESSLOGS`: HTTP access logs  
- `AUDITLOG`: Audit logs

## Example with Variables

Create a `terraform.tfvars` file:

```hcl
datadog_api_key = "your-datadog-api-key"
newrelic_api_key = "your-newrelic-api-key"
elasticsearch_password = "your-elasticsearch-password"
```

Then reference in main.tf:

```hcl
variable "datadog_api_key" {
  type = string
  sensitive = true
}

resource "clevercloud_drain_datadog" "logs" {
  resource_id = clevercloud_nodejs.app.id
  kind        = "LOG"
  url         = "https://http-intake.logs.datadoghq.com/v1/input/${var.datadog_api_key}"
}
```