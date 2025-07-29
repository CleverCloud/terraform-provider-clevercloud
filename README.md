# Terraform Clever Cloud Provider

The Clever Cloud Provider allows Terraform to manage Clever Cloud resources.

## Build provider

Run the following command to build the provider

```shell
$ go build -o terraform-provider-clevercloud
```

## Test sample configuration

First, build and install the provider.

```shell
$ make install
```

Then, navigate to the `examples` directory.

```shell
$ cd examples
```

Run the following command to initialize the workspace and apply the sample configuration.

```shell
$ terraform init && terraform apply
```

## Testing dev build

```shell
$ make install
```

```hcl
provider_installation {
    dev_overrides {
        "CleverCloud/clevercloud" = "/home/[USER]/.terraform.d/plugins/registry.terraform.io/CleverCloud/clevercloud/dev/linux_amd64"
    }
    
    direct{}
}
```

