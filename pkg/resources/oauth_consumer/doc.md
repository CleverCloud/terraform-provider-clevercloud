Manage OAuth consumers for Clever Cloud organizations.

OAuth consumers allow third-party applications to access Clever Cloud APIs on behalf of users. This resource manages the consumer credentials and permissions.

## Example Usage

```terraform
provider "clevercloud" {
  organisation = "orga_xxxx-xxxx-xxxx-xxxx-xxxx"
}

resource "clevercloud_oauth_consumer" "my_app" {
  name        = "My Application"
  description = "OAuth consumer for my application"
  base_url    = "https://api.myapp.com"
  logo_url    = "https://myapp.com/logo.png"
  website_url = "https://myapp.com"

  rights = [
    "access_organisations",
    "manage_organisations_applications",
  ]
}

# Access the generated credentials
output "oauth_key" {
  value = clevercloud_oauth_consumer.my_app.id
}

output "oauth_secret" {
  value     = clevercloud_oauth_consumer.my_app.secret
  sensitive = true
}
```

## Available Rights

The following rights can be granted to an OAuth consumer:

- `access_organisations` - View organization information
- `access_organisations_bills` - View organization bills
- `access_organisations_consumption_statistics` - View consumption statistics
- `access_organisations_credit_count` - View credit count
- `access_personal_information` - View personal information
- `manage_organisations` - Manage organization settings
- `manage_organisations_applications` - Manage applications
- `manage_organisations_members` - Manage organization members
- `manage_organisations_services` - Manage services
- `manage_personal_information` - Manage personal information
- `manage_ssh_keys` - Manage SSH keys
