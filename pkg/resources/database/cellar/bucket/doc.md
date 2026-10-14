Manage [Cellar Buckets](https://www.clever.cloud/developers/doc/addons/cellar/).

See [Cellar product specification](https://www.clever.cloud/developers/doc/addons/cellar/).

## Import

A bucket is imported by its Cellar ID and its name, separated by a `/`. The Cellar ID is the
real ID (`cellar_…`) or the add-on ID (`addon_…`), whichever `cellar_id` holds in your configuration:

```terraform
import {
  to = clevercloud_cellar_bucket.my_bucket
  id = "cellar_xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx/my-bucket"
}
```

```bash
terraform import clevercloud_cellar_bucket.my_bucket cellar_xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx/my-bucket
```
