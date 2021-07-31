---
page_title: "Clever Cloud: clevercloud_self"
description: |-
  Get information on current user.
---

# clevercloud_zones

Use this data source to get information on current user.

## Example Usage

```hcl
data "clevercloud_self" "current" {
}
```

## Attributes Reference

In addition to all above arguments, the following attributes are exported:

- `name` - The user name.
- `email` - The user email address.
- `email_validated` - The user email validation.
- `phone` - The user phone number.
- `address` - The user address.
- `city` - The user city.
- `zip_code` - The user city zip code.
- `country` - The user country.
- `avatar` - The user avatar URL.
- `creation_date` - The user creation date.
- `language` - The user language.
- `oauth_apps` - The user OAuth applications.
- `admin` - The user admin capability.
- `can_pay` - The user payment capability.
- `preferred_mfa` - The user MFA setting.
- `has_password` - The user password setting.