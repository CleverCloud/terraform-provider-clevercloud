> Action used to reboot an application

This action wait for the end of reboot (when new instances are ready)

Exemple:

```hcl
terraform {
  required_version = ">= 1.14.0"

  required_providers {
    clevercloud = {
      source  = "CleverCloud/clevercloud"
      version = "1.2.0"
    }
  }
}

provider "clevercloud" {
  organisation = "orga_xxx"
}

action "clevercloud_application_reboot" "restart_php" {
  config {
    application_id = "app_16247e01-849e-4z95-b5ca-be883e849562"
  }
}
```

### Manual trigger

```sh
terraform apply -invoke action.clevercloud_application_reboot.restart_php
```