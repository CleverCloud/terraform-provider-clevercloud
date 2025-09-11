Manage any add-on through the [addon-provider API](https://www.clever.cloud/developers/doc/marketplace/#add-on-provider-requests)


List of available providers:

* [Mailpace](https://www.clever.cloud/developers/doc/addons/mailpace/)


## Argument Reference

### Generic arguments

* `name` - (Required) Name of the addon.
* `region` - (Optional) Geographical region where the data will be stored. Defaults to `par`.

### Specific arguments

* `third_party_provider` - (Required) Name of a provider

## Attribute Reference

* `id` - Generated unique identifier.
* `name` - Name of the instance.
