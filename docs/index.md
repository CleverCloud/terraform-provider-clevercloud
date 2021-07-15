---
page_title: "Provider: Clever Cloud"
description: |-
  The Clever Cloud provider is used to manage Clever Cloud resources. The provider needs to be configured with the proper credentials before it can be used.
---

# Clever Cloud Provider

The Clever Cloud provider is used to manage Clever Cloud resources.
The provider needs to be configured with the proper credentials before it can be used.

Use the navigation to the left to read about the available resources.

## Example

You can test this config by creating a `test.tf` and run terraform commands from this directory:

- Initialize a Terraform working directory: `terraform init`
- Generate and show the execution plan: `terraform plan`
- Build the infrastructure: `terraform apply`

```hcl
terraform {
  required_providers {
    clevercloud = {
      source = "clevercloud/clevercloud"
    }
  }
}

provider "clevercloud" {}
```

## Authentication

The Clever Cloud authentication is based on OAuth 1.0 which require a **token**, and a **secret**.

You can generate a token using our CLI with the command `clever login`.

The Clever Cloud provider offers three ways of providing these credentials.
The following methods are supported, in this priority order:

- [Clever Cloud Provider](#clever-cloud-provider)
  - [Example](#example)
  - [Authentication](#authentication)
    - [Environment variables](#environment-variables)
    - [Static credentials](#static-credentials)
    - [Shared configuration file](#shared-configuration-file)
  - [Arguments Reference](#arguments-reference)

### Environment variables

You can provide your credentials via the `CLEVER_TOKEN`, `CLEVER_SECRET` environment variables.

Example:

```hcl
provider "clevercloud" {}
```

Usage:

```bash
$ export CLEVER_TOKEN="my-token"
$ export CLEVER_SECRET="my-secret"
$ terraform plan
```

### Static credentials

Static credentials can be provided by adding `token` and `secret` attributes using variables in the Clever Cloud provider block:

Example:

```hcl
variable "clevercloud_token" {
  type = string
}

variable "clevercloud_secret" {
  type = string
}

provider "clevercloud" {
  token = var.clevercloud_token
  secret = var.clevercloud_secret
}
```

### Shared configuration file

It is a configuration file shared with the [Clever Cloud CLI](https://www.clever-cloud.com/doc/getting-started/cli/).
Its default location is `$HOME/.config/clever-cloud` (`%USERPROFILE%/.config/clever-cloud` on Windows).
If it fails to detect credentials inline, or in the environment, Terraform will check this file.

## Arguments Reference

In addition to [generic provider arguments](https://www.terraform.io/docs/configuration/providers.html) (e.g. `alias` and `version`), the following arguments are supported in the Clever Cloud provider block:

| Provider Argument | [Environment Variables](#environment-variables) | Description                                                                                                                             | Mandatory |
|-------------------|-------------------------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------|-----------|
| `token`           | `CLEVER_TOKEN`                                  | Clever Cloud OAuth token                                                                                                                | ✅         |
| `secret`          | `CLEVER_SECRET`                                 | Clever Cloud OAuth secret                                                                                                              | ✅         |
