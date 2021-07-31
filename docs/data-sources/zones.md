---
page_title: "Clever Cloud: clevercloud_zones"
description: |-
  Get the list of the zones list.
---

# clevercloud_zones

Use this data source to get the zones list.

## Example Usage

```hcl
data "clevercloud_zones" "available" {
}
```

## Attributes Reference

In addition to all above arguments, the following attributes are exported:

- `zones` - The list of zones.
  - `name` - The zone name.
  - `internal` - The zone internal value.
  - `corresponding_region` - The zone corresponding region.