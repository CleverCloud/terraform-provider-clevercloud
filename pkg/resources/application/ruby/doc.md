# Ruby resource

Manage a Ruby application on Clever Cloud.

## Example Usage

```terraform
# Simple Ruby application
resource "clevercloud_ruby" "simple_ruby_app" {
  name               = "my-ruby-app"
  region             = "par"
  min_instance_count = 1
  max_instance_count = 2
  smallest_flavor    = "XS" 
  biggest_flavor     = "M"
  ruby_version       = "3.3"
}

# Ruby on Rails application with database and custom configuration
resource "clevercloud_ruby" "rails_app" {
  name               = "my-rails-app"
  region             = "par"
  min_instance_count = 1
  max_instance_count = 4
  smallest_flavor    = "XS"
  biggest_flavor     = "L"
  build_flavor       = "XL"
  
  ruby_version       = "3.3.1"
  rails_env          = "production"
  rake_goals         = "db:migrate,assets:precompile"
  enable_sidekiq     = true
  sidekiq_files      = "./config/sidekiq.yml"
  
  environment = {
    SECRET_KEY_BASE = "your-secret-key"
    DATABASE_URL    = "postgresql://..."
  }
  
  deployment {
    repository = "https://github.com/your-org/your-rails-app.git"
    commit     = "main"
  }
  
  depends_on = [clevercloud_postgresql.db]
}

# Ruby application with custom server and static files
resource "clevercloud_ruby" "custom_ruby_app" {
  name               = "custom-ruby-app"
  region             = "par"
  min_instance_count = 2
  max_instance_count = 6
  smallest_flavor    = "S"
  biggest_flavor     = "XL"
  
  ruby_version        = "3.2"
  rackup_server      = "unicorn"
  rack_env           = "production"
  static_files_path  = "public"
  static_url_prefix  = "/assets"
  
  enable_gzip_compression = true
  nginx_read_timeout     = 600
  
  hooks {
    pre_build  = "bundle config set --local deployment true"
    post_build = "bundle exec rake assets:precompile"
  }
}
```