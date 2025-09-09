# Ruby Resource

Provides a Ruby application resource.

## Example Usage

```hcl
resource "clevercloud_ruby" "my_ruby_app" {
  name             = "my-ruby-app"
  region           = "par"
  
  # Instance sizing
  smallest_flavor  = "nano"
  biggest_flavor   = "nano"
  
  # Scaling
  min_instance_count = 1
  max_instance_count = 1
  
  # Ruby-specific configuration
  ruby_version = "3.1"
  bundler_file = "Gemfile"
  
  # Environment variables
  environment = {
    PORT = "8080"
  }
  
  # Custom domain names
  vhosts = [
    "my-ruby-app.example.com"
  ]
  
  # Force HTTPS redirection
  redirect_https = true
}
```

## Argument Reference

* `name` - (Required) The name of the application.
* `region` - (Required) The region where the application will be deployed.
* `smallest_flavor` - (Required) The smallest instance flavor to use.
* `biggest_flavor` - (Required) The biggest instance flavor to use.
* `min_instance_count` - (Required) The minimum number of instances.
* `max_instance_count` - (Required) The maximum number of instances.
* `ruby_version` - (Optional) The Ruby version to use. Default is "3.1".
* `bundler_file` - (Optional) The path to the Gemfile. Default is "Gemfile".
* `environment` - (Optional) A map of environment variables.
* `vhosts` - (Optional) A list of custom domain names.
* `redirect_https` - (Optional) Whether to force HTTPS redirection. Default is true.
* `description` - (Optional) A description for the application.
* `deployment` - (Optional) A deployment configuration block.
* `dependencies` - (Optional) A list of addon dependencies.

## Attribute Reference

* `id` - The ID of the application.
* `deploy_url` - The Git URL to deploy to.

## Import

Ruby applications can be imported using the application ID:

```
terraform import clevercloud_ruby.my_ruby_app app_12345678-90ab-cdef-1234-567890abcdef
```
