Retrieves information about a specific PostgreSQL backup from Clever Cloud.

The backup can be identified in three ways:
- By UUID for exact match
- By date to find the most recent backup before (or at) that date
- By using "latest" to get the most recent backup

You can fetch backup list and associated UUID with a clever-tools command.

Ex: `clever database backups postgresql_9629c177-4ef0-4b61-9ca6-b67c2db35e3b`

## Example Usage

### Lookup by UUID

```terraform
data "clevercloud_postgresql_backup" "by_uuid" {
  postgresql_id = "postgresql_9629c177-4ef0-4b61-9ca6-b67c2db35e3b"
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

## Notes

- Backups are created automatically every 24 hours for PostgreSQL addons
- Newly created PostgreSQL addons will not have backups until 24 hours after creation
- The `download_url` is a signed S3 URL that expires. Always retrieve a fresh datasource before downloading a backup
- When using a date selector, the datasource returns the most recent backup created before or at the specified date
