# Terraform Clever Cloud Provider

The Clever Cloud Provider allows Terraform to manage Clever Cloud resources.

## Prerequisites

- Go 1.23.7 or later
- [golangci-lint](https://golangci-lint.run/welcome/install/) for code linting
- [tfplugindocs](https://github.com/hashicorp/terraform-plugin-docs) for documentation generation

## Authentication

To use this provider, you need to configure your Clever Cloud credentials. You can obtain these from your [Clever Cloud console](https://console.clever-cloud.com) or [clever-tools](https://github.com/clevercloud/clever-tools).

Set the following environment variables:

```bash
export CLEVER_TOKEN="your-clever-cloud-token"
export CLEVER_SECRET="your-clever-cloud-secret"
export ORGANISATION="your-organization-id"  # Optional, for multi-org accounts
```

## Build provider

Run the following command to build the provider

```shell
$ go build -o terraform-provider-clevercloud
```

## Build and install the provider

Run the following command to build and install the provider in the current user's terraform plugin directory

```shell
$ make
```

To produce an unstripped binary (for debugging purpose), override `LDFLAGS` with an empty value:

```shell
$ make LDFLAGS=""
```

## Examples

The [`examples`](./examples) directory contains sample Terraform configurations demonstrating how to use the provider:

- **Basic usage** (`main.tf`): Shows how to create PostgreSQL databases, Node.js applications, and Cellar storage
- **Individual resources** (`resources/`): Detailed examples for specific resource types

To test the examples:

1. Build and install the provider:
   ```shell
   make install
   ```

2. Navigate to the examples directory:
   ```shell
   cd examples
   ```

3. Set required variables:
   ```shell
   export TF_VAR_organisation="your-org-id"
   ```

4. Initialize and apply:
   ```shell
   terraform init && terraform apply
   ```

## Testing dev build

First, build and install the provider:

```shell
$ make install
```

Then, create or modify the file `~/.terraformrc` with the following configuration:

```hcl
provider_installation {
    dev_overrides {
        "CleverCloud/clevercloud" = "/home/[USER]/.terraform.d/plugins/registry.terraform.io/CleverCloud/clevercloud/dev/linux_amd64"
    }

    direct{}
}
```

## Documentation

Full documentation for all resources and data sources is available in the [`/docs`](./docs) directory. The documentation is automatically generated from the provider schema and includes:

- [Provider configuration](./docs/index.md)
- Resource documentation for all supported services (Docker, Node.js, PostgreSQL, etc.)
- Usage examples and configuration options

To regenerate the documentation after making changes:

```shell
make docs
```

