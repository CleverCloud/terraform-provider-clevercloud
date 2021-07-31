---
page_title: "Clever Cloud: clevercloud_addon"
description: |-
  Get information on an addon.
---

# clevercloud_addon

Use this data source to get addon information based on its ID.

## Example Usage

```hcl
data "clevercloud_addon" "foobar_addon" {
  id = "addon_a3c22ce4-6ce1-4586-bbe3-2b76ea015352"
}
```

## Argument Reference

- `id` - The add-on id starting with `addon_` prefix.
- `organization_id` - The ID of the organization that the add-on belongs to (optionnal).

## Attributes Reference

In addition to all above arguments, the following attributes are exported:

- `name` - The add-on name.
- `real_id` - The add-on real id with the specific addon prefix e.g. `postgresql_`.
- `region` - The region where is located the add-on.
- `config_keys` - The environnement variables exposed by the add-on when linked to an application.
- `provider_info` - The add-on provider information.
  - `id` - The add-on provider id.
  - `name` - The add-on provider name.
  - `website` - The add-on provider website.
  - `short_desc` - The add-on provider short description.
  - `long_desc` - The add-on provider description.
  - `status` - The add-on provider status.
  - `can_upgrade` - The add-on provider ability to be upgrade to another plan.
  - `regions` - The add-on provider available regions.
- `plan` - The current add-on plan.
  - `id` - The add-on plan id.
  - `name` - The add-on plan name.
  - `slug` - The add-on plan slug.
  - `features` - The add-on plan features (CPU, memory, storage etc ...).
    - `name` - The add-on feature name.
    - `name_code` - The add-on feature slugged name.
    - `type` - The add-on feature value type (number, boolean, bytes etc ...).
    - `computable_value` - The add-on feature value.
    - `value` - The add-on feature human readable value.
  - `price` - The add-on plan price.
  - `zones` - The add-on plan available zones.
- `creation_date` - The add-on creation date.
