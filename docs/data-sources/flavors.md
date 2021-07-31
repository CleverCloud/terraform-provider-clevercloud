---
page_title: "Clever Cloud: clevercloud_flavors"
description: |-
  Get the list of application instance flavors.
---

# clevercloud_flavors

Use this data source to get application instance flavors.

## Example Usage

```hcl
data "clevercloud_flavors" "available" {
}
```

## Attributes Reference

In addition to all above arguments, the following attributes are exported:

- `flavors` - The list of flavors.
  - `name` - The flavor name.
  - `mem` - The flavor memory quantity in MB.
  - `memory` - The flavor memory details.
    - `unit` - The flavor memory unit value.
    - `value` - the flavor memory value.
    - `formatted` - The flavor human readable memory.
  - `cpus` - The flavor CPU quantity.
  - `gpus` - The flavor GPU quantity.
  - `disk` - The flavor disk quantity.
  - `price` - The flavor price per hour.
  - `price_id` - The flavor price ID.
  - `available` - The flavor availability.
  - `microservice` - The flavor microservice capability.
  - `machine_learning` - The flavor machine learning capability.
  - `nice` - The flavor nice factor.
  - `rbd_image` - The flavor RBD image.
