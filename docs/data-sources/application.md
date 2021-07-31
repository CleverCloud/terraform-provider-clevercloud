---
page_title: "Clever Cloud: clevercloud_application"
description: |-
  Get information on an application.
---

# clevercloud_application

Use this data source to get application information based on its ID.

## Example Usage

```hcl
data "clevercloud_application" "foobar_application" {
  id = "app_a3c22ce4-6ce1-4586-bbe3-2b76ea015352"
}
```

## Argument Reference

- `id` - The application ID.
- `organization_id` - The ID of the organization that the application belongs to (optionnal).

## Attributes Reference

In addition to all above arguments, the following attributes are exported:

- `name` - The application name.
- `description` - The application description.
- `zone` - The zone where the application is located.
- `instance` - The application instance information.
  - `type` - The application instance type (e.g. Node.js, PHP, Ruby etc...).
  - `version` - The application instance version.
  - `variant` - The application instance variant.
    - `id` - The variant ID.
    - `name` - The variant name.
    - `slug` - The variant slug.
    - `logo` - The variant logo.
    - `deploy_type` - The variant deploy type.
  - `min_instances` - The application minimum instances for scaling.
  - `max_instances` - The application maximum instances for scaling.
  - `max_allowed_instances` - The maximum instances allowed.
  - `min_flavor` - The application minimum flavor for scaling.
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
  - `max_flavor` - The application maximum flavor for scaling.
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
  - `default_env` - The instance default environnement variables
  - `lifetime` - The instance lifetime type.
  - `instance_version` - The instance version (`type`_`version`).
- `deployment` - The application deployment information
  - `type` - The deployment type.
  - `url` - The deployment URL.
  - `http_url` - The deployment HTTP URL.
  - `repo_state` - The deployment repository state.
  - `shutdownable` - The deployment ability to be shutdown.
- `deploy_url` - The application deployment URL.
- `vhosts` - The application domains.
  - `fqdn` - The domain FQDN.
- `archived` - The application archived status.
- `sticky_sessions` - The application "Sticky sessions" setting.
- `homogeneous` - The application "Homogeneous" setting.
- `cancel_on_push` - The application "Cancel on push" setting.
- `force_https` - The application "Force HTTPS" setting.
- `seperate_build` - The application "Seperate build" setting.
- `build_flavor` - The application build flavor.
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
- `state` - The application state.
- `commit_id` - The application git commit ID.
- `branch` - The application git branch.
- `webhook_url` - The application webhook URL.
- `webhook_secret` - The application webhook secret.
- `appliance` - The application appliance type.
- `favorite` - The application favorite status.
- `owner_id` - The application real owner ID.
