terraform {
  required_providers {
    clevercloud = {
      source  = "CleverCloud/clevercloud"
      version = "~> 0.1"
    }
  }
}

provider "clevercloud" {
  # Configure your organisation ID
}

# Create a simple Node.js application
resource "clevercloud_nodejs" "app" {
  name         = "drain-example-app"
  region       = "par"
  min_instance_count = 1
  max_instance_count = 1
  smallest_flavor    = "XS"
  biggest_flavor     = "XS"
}

# Example HTTP drain
resource "clevercloud_drain_http" "example" {
  resource_id = clevercloud_nodejs.app.id
  kind        = "LOG"
  url         = "https://httpbin.org/post"
}

# Example Syslog UDP drain
resource "clevercloud_drain_syslog_udp" "syslog" {
  resource_id = clevercloud_nodejs.app.id
  kind        = "LOG"
  url         = "udp://logs.example.com:514"
}

# Example Datadog drain (requires valid API key)
resource "clevercloud_drain_datadog" "datadog" {
  resource_id = clevercloud_nodejs.app.id
  kind        = "LOG"
  url         = "https://http-intake.logs.datadoghq.com/v1/input/YOUR_API_KEY"
}

# Example New Relic drain (requires valid API key)
resource "clevercloud_drain_newrelic" "newrelic" {
  resource_id = clevercloud_nodejs.app.id
  kind        = "LOG"
  url         = "https://log-api.newrelic.com/log/v1"
  api_key     = "YOUR_NEW_RELIC_API_KEY"
}

# Example Elasticsearch drain
resource "clevercloud_drain_elasticsearch" "elasticsearch" {
  resource_id      = clevercloud_nodejs.app.id
  kind             = "LOG"
  url              = "https://your-elasticsearch.com"
  username         = "elastic"
  password         = "your-password"
  index            = "clever-cloud-logs"
  tls_verification = "DEFAULT"
}

# Output the drain IDs
output "http_drain_id" {
  value = clevercloud_drain_http.example.id
}

output "datadog_drain_id" {
  value = clevercloud_drain_datadog.datadog.id
}