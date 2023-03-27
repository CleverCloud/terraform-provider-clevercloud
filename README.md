# Terraform Provider Hashicups

This repo is a companion repo to the [Call APIs with Terraform Providers](https://learn.hashicorp.com/collections/terraform/providers) Learn collection.

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

> By using [Terraform Cloud](https://app.terraform.io/) features, you have to share your [ClerverCloud](https://www.clever-cloud.com/) credential with a third party company with non European data hosting.
>
> If you store your credentials outside of the European union, you will lose your data sovereignty
