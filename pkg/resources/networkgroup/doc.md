Manage any [Network Group](https://www.clever.cloud/developers/doc/develop/network-groups/).

See [Network Groups product specification](https://www.clever.cloud/developers/doc/develop/network-groups/).

~> Network groups cannot be updated in place: any change to the effective tag set (the resource `tags`, or the provider-level `default_tags` unless `ignore_default_tags` is set) **replaces the network group**, dropping its members during recreation. Set `ignore_default_tags = true` to shield a network group from provider-level tag changes.
