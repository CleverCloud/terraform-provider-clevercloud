terraform {
  required_providers {
    clevercloud = {
      source = "CleverCloud/clevercloud"
    }
  }
}

provider "clevercloud" {
  # Configuration will be read from environment variables or config file
}

# Create an FSBucket
resource "clevercloud_fsbucket" "example" {
  name   = "example-bucket"
  region = "par"
}

# Upload a single local file
action "clevercloud_fsbucket_upload" "single_file" {
  fsbucket_id = clevercloud_fsbucket.example.id

  file {
    local_path  = "file://${path.module}/files/config.json"
    remote_path = "/config/config.json"
  }
}

# Upload from HTTP(S) URL
action "clevercloud_fsbucket_upload" "from_url" {
  fsbucket_id = clevercloud_fsbucket.example.id

  # Download Bootstrap CSS from CDN
  file {
    local_path  = "https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css"
    remote_path = "/static/css/bootstrap.min.css"
  }

  # Download jQuery from CDN
  file {
    local_path  = "https://code.jquery.com/jquery-3.7.0.min.js"
    remote_path = "/static/js/jquery.min.js"
  }
}

# Upload multiple local files
action "clevercloud_fsbucket_upload" "multiple_files" {
  fsbucket_id = clevercloud_fsbucket.example.id

  file {
    local_path  = "file://${path.module}/files/README.md"
    remote_path = "/README.md"
  }

  file {
    local_path  = "file://${path.module}/files/LICENSE"
    remote_path = "/LICENSE"
  }

  file {
    local_path  = "file://${path.module}/files/assets/logo.png"
    remote_path = "/public/logo.png"
  }
}

# Upload a directory (automatically recursive)
action "clevercloud_fsbucket_upload" "upload_dist" {
  fsbucket_id = clevercloud_fsbucket.example.id

  # Directories are automatically uploaded recursively
  file {
    local_path  = "file://${path.module}/dist/"
    remote_path = "/app/"
  }
}

# Mixed upload: local files, directories, and HTTP sources
action "clevercloud_fsbucket_upload" "deploy_app" {
  fsbucket_id = clevercloud_fsbucket.example.id

  # Local configuration file
  file {
    local_path  = "file://${path.module}/config.production.json"
    remote_path = "/config.json"
  }

  # Local static assets directory (automatically recursive)
  file {
    local_path  = "file://${path.module}/public/"
    remote_path = "/static/"
  }

  # Local build artifacts (automatically recursive)
  file {
    local_path  = "file://${path.module}/build/"
    remote_path = "/app/"
  }
}
