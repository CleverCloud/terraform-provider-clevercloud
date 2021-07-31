---
page_title: "Clever Cloud: clevercloud_addon_providers"
description: |-
  Get the list of available add-on providers.
---

# clevercloud_addon_providers

Use this data source to get available add-ons providers.

## Example Usage

```hcl
data "clevercloud_addon_providers" "available" {
}
```

## Argument Reference

- `organization_id` - The ID of the organization, to get available add-on providers to this organization (optionnal).

## Attributes Reference

In addition to all above arguments, the following attributes are exported:

- `addon_providers` - The list of available add-on providers.
  - `id` - The add-on provider id.
  - `name` - The add-on provider name.
  - `website` - The add-on provider website.
  - `short_desc` - The add-on provider short description.
  - `long_desc` - The add-on provider description.
  - `status` - The add-on provider status.
  - `can_upgrade` - The add-on provider ability to be upgrade to another plan.
  - `regions` - The add-on provider available regions.
  - `plans` - The add-on provider plans.
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
  - `features` - The add-on plan features (CPU, memory, storage etc ...).
    - `name` - The add-on provider feature name.
    - `name_code` - The add-on feature slugged name.
    - `type` - The add-on feature value type (number, boolean, bytes etc ...).
