## FSBucket Resource

This resource allows you to create and manage an FSBucket on Clever Cloud.

FSBucket is a managed object storage solution that provides S3-compatible storage.

See [FS Bucket product specification](https://www.clever.cloud/developers/doc/addons/fs-bucket/).

## Example Usage

```terraform
resource "clevercloud_fsbucket" "my_fsbucket" {
  name = "my-fsbucket"
  region = "par"
}
```

## Argument Reference

* `name` - (Required) Name of the FSBucket.
* `region` - (Optional) Geographical region where the data will be stored. Defaults to `par`.

## Attribute Reference

* `id` - Generated unique identifier.
* `host` - FSBucket FTP endpoint.
* `ftp_username` - FTP username used to authenticate.
* `ftp_password` - FTP password used to authenticate.
