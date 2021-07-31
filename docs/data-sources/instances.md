---
page_title: "Clever Cloud: clevercloud_instances"
description: |-
  Get the list of application instances.
---

# clevercloud_instances

Use this data source to get application instances.

## Example Usage

```hcl
data "clevercloud_instances" "available" {
}
```

## Attributes Reference

In addition to all above arguments, the following attributes are exported:

- `instances` - The list of application instances.
  - `name` - The application instance name.
  - `description` - The application instance description.
  - `type` - The application instance type (e.g. Node.js, PHP, Ruby etc...).
  - `version` - The application instance version.
  - `variant` - The application instance variant.
    - `id` - The variant ID.
    - `name` - The variant name.
    - `slug` - The variant slug.
    - `logo` - The variant logo.
    - `deploy_type` - The variant deploy type.
  - `enabled` - The application instance enabled value.
  - `coming_soon`: - the application instance coming soon value.
  - `max_instances` - The application maximum instances for scaling.
  - `tags` - The application instances tags.
  - `deployments` - The application instance deployments. 
  - `flavors` - The application instance available flavors.
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
  - `default_flavor` - The application instance default flavor.
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
  - `build_flavor` - The application instance build flavor.
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
