# FS Bucket Resource

This resource allows you to create and manage an FS Bucket on Clever Cloud.

FS Bucket is a managed object storage solution that provides S3-compatible storage.

See [FS Bucket product specification](https://www.clever.cloud/developers/doc/addons/fs-bucket/).

## Example usage

```terraform
resource "clevercloud_fsbucket" "my_fsbucket" {
  name = "my-fsbucket"
  region = "par"
}
```

## Argument Reference

* `name` - (Required) Name of the FS Bucket.
* `region` - (Optional) Geographical region where the data will be stored. Defaults to `par`.

## Attribute Reference

* `id` - Generated unique identifier.
* `host` - FS Bucket FTP endpoint.
* `ftp_username` - FTP username used to authenticate.
* `ftp_password` - FTP password used to authenticate.
