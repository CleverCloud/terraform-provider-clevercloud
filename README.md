# Terraform Provider Clever Cloud

## ToDo

- [ ] Data Sources
  - [ ] Application
  - [ ] Addon
  - [ ] User / Organization (self)
  - [ ] Addon providers
  - [ ] 
  - [ ] Flavors
  - [ ] Zones
- [ ] Resources
  - [ ] Application
  - [ ] Addon

## Build provider

Run the following command to build the provider

```shell
make build
```

## Test sample configuration

First, build and install the provider.

```shell
make install
```

Then, run the following command to initialize the workspace and apply the sample configuration.

```shell
terraform init && terraform apply
```