# FS Bucket

## FS Bucket Resource

This resource allows you to create and manage an FS Bucket on Clever Cloud.

FS Bucket is a managed object storage solution that provides S3-compatible storage.

See [FS Bucket product specification](https://www.clever.cloud/developers/doc/addons/fs-bucket/).

## Example usage

### Basic

```terraform
resource "clevercloud_fsbucket" "my_fsbucket" {
  name = "my-fsbucket"
  region = "par"
}
```

