---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "clevercloud_php Resource - terraform-provider-clevercloud"
description: |-
  Manage PHP https://www.php.net/ applications.
  See PHP product https://www.clever-cloud.com/doc/getting-started/by-language/php/ specification.
---

# clevercloud_php (Resource)

Manage [PHP](https://www.php.net/) applications.

See [PHP product](https://www.clever-cloud.com/doc/getting-started/by-language/php/) specification.



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `biggest_flavor` (String) Biggest intance flavor, if different from smallest, enable autoscaling
- `max_instance_count` (Number) Maximum instance count, if different from min value, enable autoscaling
- `min_instance_count` (Number) Minimum instance count
- `name` (String) Application name
- `smallest_flavor` (String) Smallest instance flavor

### Optional

- `additional_vhosts` (List of String, Deprecated) Add custom hostname in addition to the default one, see [documentation](https://www.clever-cloud.com/doc/administrate/domain-names/)
- `app_folder` (String) Folder in which the application is located (inside the git repository)
- `build_flavor` (String) Use dedicated instance with given flavor for build step
- `dependencies` (Set of String) A list of application or addons requires to run this application.
Can be either app_xxx or postgres_yyy ID format
- `deployment` (Block, Optional) (see [below for nested schema](#nestedblock--deployment))
- `description` (String) Application description
- `dev_dependencies` (Boolean) Install development dependencies
- `environment` (Map of String, Sensitive) Environment variables injected into the application
- `hooks` (Block, Optional) (see [below for nested schema](#nestedblock--hooks))
- `php_version` (String) PHP version (Default: 8)
- `redirect_https` (Boolean) Redirect client from plain to TLS port
- `redis_sessions` (Boolean) Use a linked Redis instance to store sessions (Default: false)
- `region` (String) Geographical region where the database will be deployed
- `sticky_sessions` (Boolean) Enable sticky sessions, use it when your client sessions are instances scoped
- `vhosts` (Set of String) Add custom hostname, see [documentation](https://www.clever-cloud.com/doc/administrate/domain-names/)
- `webroot` (String) Define the DocumentRoot of your project (default: ".")

### Read-Only

- `deploy_url` (String) Git URL used to push source code
- `id` (String) Unique identifier generated during application creation

<a id="nestedblock--deployment"></a>
### Nested Schema for `deployment`

Optional:

- `commit` (String) Support multiple syntax like `refs/heads/[BRANCH]` or `[COMMIT]`, in most of the case, you can use `refs/heads/master`
- `repository` (String) The repository URL to deploy, can be 'https://...', 'file://...'


<a id="nestedblock--hooks"></a>
### Nested Schema for `hooks`

Optional:

- `post_build` (String) [CC_POST_BUILD_HOOK](https://www.clever-cloud.com/doc/develop/build-hooks/#post-build-cc_post_build_hook)
- `pre_build` (String) [CC_PRE_BUILD_HOOK](https://www.clever-cloud.com/doc/develop/build-hooks/#pre-build-cc_pre_build_hook)
- `pre_run` (String) [CC_PRE_RUN_HOOK](https://www.clever-cloud.com/doc/develop/build-hooks/#pre-run-cc_pre_run_hook)
- `run_failed` (String) [CC_RUN_FAILED_HOOK](https://www.clever-cloud.com/doc/develop/build-hooks/#run-succeeded-cc_run_succeeded_hook-or-failed-cc_run_failed_hook)
- `run_succeed` (String) [CC_RUN_SUCCEEDED_HOOK](https://www.clever-cloud.com/doc/develop/build-hooks/#run-succeeded-cc_run_succeeded_hook-or-failed-cc_run_failed_hook)
