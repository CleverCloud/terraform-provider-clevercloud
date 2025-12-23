---
page_title: "clevercloud_postgresql_backup Data Source - terraform-provider-clevercloud"
subcategory: ""
description: |-
  Retrieves information about a specific PostgreSQL backup from Clever Cloud.
---

# clevercloud_postgresql_backup (Data Source)

Retrieves information about a specific PostgreSQL backup from Clever Cloud.

The backup can be identified in three ways:
- By UUID for exact match
- By date to find the most recent backup before (or at) that date
- By using "latest" to get the most recent backup

## Example Usage

### Lookup by UUID

```terraform
data "clevercloud_postgresql_backup" "by_uuid" {
  postgresql_id = "postgresql_a1b2c3d4"
  selector      = "550e8400-e29b-41d4-a716-446655440000"
}

output "backup_url" {
  value = data.clevercloud_postgresql_backup.by_uuid.download_url
}
```

### Lookup by Date

```terraform
data "clevercloud_postgresql_backup" "by_date" {
  postgresql_id = clevercloud_postgresql.mydb.id
  selector      = "2025-12-23T10:00:00Z"
}

output "latest_backup_before_date" {
  value = {
    backup_id     = data.clevercloud_postgresql_backup.by_date.id
    creation_date = data.clevercloud_postgresql_backup.by_date.creation_date
    download_url  = data.clevercloud_postgresql_backup.by_date.download_url
  }
}
```

### Lookup Latest Backup

```terraform
data "clevercloud_postgresql_backup" "latest" {
  postgresql_id = clevercloud_postgresql.mydb.id
  selector      = "latest"
}

# Download the latest backup
resource "null_resource" "download_backup" {
  provisioner "local-exec" {
    command = "curl -o backup.dump '${data.clevercloud_postgresql_backup.latest.download_url}'"
  }
}
```

## Schema

### Required

- `postgresql_id` (String) The PostgreSQL addon ID. Accepts both formats: `postgresql_xxx` (real ID) or `addon_xxx` (addon ID)
- `selector` (String) One of three options:
  - A backup UUID for exact match (e.g., `550e8400-e29b-41d4-a716-446655440000`)
  - An ISO8601 date string to find the most recent backup before this date (e.g., `2025-12-23T10:00:00Z`)
  - The string `latest` to get the most recent backup

### Read-Only

- `id` (String) The backup UUID
- `download_url` (String) The signed S3 URL to download the backup archive. **Note:** This URL expires after 2 hours. Retrieve fresh datasource before downloading.
- `creation_date` (String) The ISO8601 timestamp when the backup was created
- `delete_at` (String) The ISO8601 timestamp when the backup will be deleted

## Notes

- Backups are created automatically every 24 hours for PostgreSQL addons
- Newly created PostgreSQL addons will not have backups until 24 hours after creation
- The `download_url` is a signed S3 URL that expires. Always retrieve a fresh datasource before downloading a backup
- When using a date selector, the datasource returns the most recent backup created before or at the specified date
