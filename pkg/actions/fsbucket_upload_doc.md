# FSBucket Upload Action

Upload files and directories to a Clever Cloud FSBucket using FTP.

This action uploads local files, directories, or downloads from HTTP(S) URLs to an FSBucket addon. It automatically retrieves the FTP credentials from the FSBucket addon and handles the upload process. **Directories are automatically detected and uploaded recursively.**

## Example Usage

### Upload a single local file

```hcl
resource "clevercloud_fsbucket" "my_bucket" {
  name   = "my-fsbucket"
  region = "par"
}

action "clevercloud_fsbucket_upload" "upload_file" {
  fsbucket_id = clevercloud_fsbucket.my_bucket.id

  file {
    local_path  = "file://${path.module}/files/my-file.txt"
    remote_path = "/uploads/my-file.txt"
  }
}
```

### Upload from HTTP(S) URL

```hcl
action "clevercloud_fsbucket_upload" "download_and_upload" {
  fsbucket_id = clevercloud_fsbucket.my_bucket.id

  file {
    local_path  = "https://example.com/archive.zip"
    remote_path = "/downloads/archive.zip"
  }

  file {
    local_path  = "https://cdn.example.com/style.css"
    remote_path = "/assets/style.css"
  }
}
```

### Upload multiple local files

```hcl
action "clevercloud_fsbucket_upload" "upload_files" {
  fsbucket_id = clevercloud_fsbucket.my_bucket.id

  file {
    local_path  = "file://README.md"
    remote_path = "/README.md"
  }

  file {
    local_path  = "file://config.json"
    remote_path = "/config/config.json"
  }

  file {
    local_path  = "file://assets/logo.png"
    remote_path = "/public/logo.png"
  }
}
```

### Upload a directory (automatically recursive)

```hcl
action "clevercloud_fsbucket_upload" "upload_directory" {
  fsbucket_id = clevercloud_fsbucket.my_bucket.id

  # Directories are automatically uploaded recursively
  file {
    local_path  = "file://dist/"
    remote_path = "/app/"
  }
}
```

### Mixed uploads (files, directories, and HTTP sources)

```hcl
action "clevercloud_fsbucket_upload" "deploy" {
  fsbucket_id = clevercloud_fsbucket.my_bucket.id

  # Upload a local config file
  file {
    local_path  = "file://config.production.json"
    remote_path = "/config.json"
  }

  # Upload entire local build directory (automatically recursive)
  file {
    local_path  = "file://build/"
    remote_path = "/app/"
  }

  # Download and upload from HTTP
  file {
    local_path  = "https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css"
    remote_path = "/static/css/bootstrap.min.css"
  }
}
```

## Attributes

- `fsbucket_id` (String, Required) - The ID of the FSBucket addon to upload to. Must be a valid FSBucket ID starting with `bucket_`.

## Blocks

### `file` Block

The `file` block defines a file or directory to upload. You can specify multiple `file` blocks to upload multiple files/directories.

- `local_path` (String, Required) - The source to upload from. Supports the following URL schemes:
  - `file://` - Local filesystem path (files or directories)
  - `http://` - Download from HTTP URL
  - `https://` - Download from HTTPS URL

- `remote_path` (String, Required) - The destination path in the FSBucket where the file(s) will be stored.

## Supported URL Schemes

### file:// - Local Filesystem

Upload files or directories from the local filesystem where Terraform is running. **Directories are automatically detected and uploaded recursively** - you don't need to specify any additional parameter.

```hcl
# Single file
file {
  local_path  = "file:///absolute/path/to/file.txt"
  remote_path = "/file.txt"
}

# Relative path
file {
  local_path  = "file://./relative/path/file.txt"
  remote_path = "/file.txt"
}

# Directory - automatically uploaded recursively
file {
  local_path  = "file://./dist/"
  remote_path = "/app/"
}
```

**How it works:**
- If `local_path` points to a **file**, the file is uploaded
- If `local_path` points to a **directory**, all files in the directory are uploaded recursively, preserving the directory structure

### http:// and https:// - Download from URL

Download content from HTTP(S) URLs and upload to FSBucket. Useful for deploying assets from CDNs or downloading release artifacts.

```hcl
file {
  local_path  = "https://github.com/user/repo/releases/download/v1.0.0/app.zip"
  remote_path = "/releases/app.zip"
}
```

**Notes for HTTP(S) URLs**:
- Content is streamed directly from the URL to the FSBucket (no local temporary file)
- HTTP requests timeout after 5 minutes
- Only HTTP 200 OK responses are accepted
- Only single files can be downloaded (no directory support)

## Notes

- The action automatically creates parent directories in the remote path if they don't exist.
- FTP credentials are automatically retrieved from the FSBucket addon environment variables.
- The upload uses FTP protocol on port 21.
- Directories are automatically detected and uploaded recursively
- When uploading directories, the directory structure is preserved in the remote location.
- All URL schemes must be explicitly specified (file://, http://, or https://).
- For HTTP(S) downloads, ensure the URLs are publicly accessible or include authentication in the URL if needed.
