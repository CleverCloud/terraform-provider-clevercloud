Manage an Addon Provider for the Clever Cloud Marketplace.

This resource allows you to register your service as an addon provider on the Clever Cloud marketplace. Once created, your addon will be available for provisioning by Clever Cloud users.

See [Marketplace API documentation](https://www.clever-cloud.com/developers/doc/marketplace/) for more details.

## Important Notes

- **Unique Provider ID**: The `provider_id` must be unique across all Clever Cloud add-ons. Choose a distinctive identifier that won't conflict with existing providers on the platform.
- **Security**: The `password` and `sso_salt` attributes are sensitive and must be at least 35 characters long.

## Example Usage

### Basic addon provider

```hcl
resource "clevercloud_addon_provider" "my_service" {
  provider_id = "my-awesome-service"
  name        = "My Awesome Service"

  config_vars = [
    "MY_SERVICE_URL",
    "MY_SERVICE_API_KEY",
    "MY_SERVICE_SECRET"
  ]

  password = "very-secure-random-password-at-least-35-chars-long-12345678"
  sso_salt = "very-secure-random-sso-salt-at-least-35-chars-long-87654321"

  production_base_url = "https://api.myservice.com/clevercloud/resources"
  production_sso_url  = "https://api.myservice.com/clevercloud/sso/login"

  test_base_url = "https://staging.myservice.com/clevercloud/resources"
  test_sso_url  = "https://staging.myservice.com/clevercloud/sso/login"
}
```

### Addon provider with features and plans

```hcl
resource "clevercloud_addon_provider" "database_service" {
  provider_id = "my-database-service"
  name        = "My Database Service"

  config_vars = [
    "MY_DATABASE_SERVICE_URL",
    "MY_DATABASE_SERVICE_DATABASE",
    "MY_DATABASE_SERVICE_USERNAME",
    "MY_DATABASE_SERVICE_PASSWORD"
  ]

  password = "very-secure-random-password-at-least-35-chars-long-12345678"
  sso_salt = "very-secure-random-sso-salt-at-least-35-chars-long-87654321"

  production_base_url = "https://api.mydbservice.com/clevercloud/resources"
  production_sso_url  = "https://api.mydbservice.com/clevercloud/sso/login"

  test_base_url = "https://staging.mydbservice.com/clevercloud/resources"
  test_sso_url  = "https://staging.mydbservice.com/clevercloud/sso/login"

  # Define features that can vary per plan
  feature {
    name = "disk_size"
    type = "FILESIZE"
  }

  feature {
    name = "connection_limit"
    type = "NUMBER"
  }

  feature {
    name = "backup_enabled"
    type = "BOOLEAN"
  }

  # Define pricing plans with their feature values
  plan {
    name  = "Free"
    slug  = "free"
    price = 0

    features {
      name  = "disk_size"
      value = "512"  # 512 MB
    }

    features {
      name  = "connection_limit"
      value = "5"
    }

    features {
      name  = "backup_enabled"
      value = "false"
    }
  }

  plan {
    name  = "Starter"
    slug  = "starter"
    price = 9.99

    features {
      name  = "disk_size"
      value = "5120"  # 5 GB
    }

    features {
      name  = "connection_limit"
      value = "25"
    }

    features {
      name  = "backup_enabled"
      value = "true"
    }
  }

  plan {
    name  = "Professional"
    slug  = "pro"
    price = 49.99

    features {
      name  = "disk_size"
      value = "51200"  # 50 GB
    }

    features {
      name  = "connection_limit"
      value = "100"
    }

    features {
      name  = "backup_enabled"
      value = "true"
    }
  }
}
```
