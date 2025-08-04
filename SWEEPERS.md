# Test Sweepers

This document describes how to use test sweepers to clean up test resources in the Clever Cloud provider.

## Overview

Sweepers are automated cleanup functions that remove test resources left behind after acceptance tests. They are implemented in [pkg/tests/sweepers.go](pkg/tests/sweepers.go).

## Available Sweepers

The following sweepers are available:

1. **clevercloud_networkgroup** - Cleans up test network groups
2. **clevercloud_kubernetes** - Cleans up test Kubernetes clusters
3. **clevercloud_application** - Cleans up test applications (nodejs, php, java, python, golang, rust, scala, etc.)
4. **clevercloud_addon** - Cleans up test addons (postgresql, mysql, redis, pulsar, cellar, mongodb, keycloak, metabase, materiakv, otoroshi)

## Sweep Order

Sweepers are executed in the following order to respect dependencies:

1. Addons (depends on applications being present)
2. Applications (depends on network groups)
3. Kubernetes clusters (depends on network groups)
4. Network groups

## Resource Selection

Sweepers only delete resources that:
- Have names starting with `tf-test` prefix
- Are in the organization specified by the `ORGANISATION` environment variable

This ensures that only test resources created by the acceptance tests are deleted.

## Usage

### Running All Sweepers

To run all sweepers for a specific region:

```bash
ORGANISATION=orga_xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx make sweep
```

Or using the Go test command directly:

```bash
ORGANISATION=orga_xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx \
go test ./... -sweep=par -sweep-run=clevercloud
```

### Running Specific Sweepers

To run a specific sweeper:

```bash
# Clean up only network groups
ORGANISATION=orga_xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx \
go test ./... -sweep=par -sweep-run=clevercloud_networkgroup

# Clean up only Kubernetes clusters
ORGANISATION=orga_xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx \
go test ./... -sweep=par -sweep-run=clevercloud_kubernetes

# Clean up only applications
ORGANISATION=orga_xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx \
go test ./... -sweep=par -sweep-run=clevercloud_application

# Clean up only addons
ORGANISATION=orga_xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx \
go test ./... -sweep=par -sweep-run=clevercloud_addon
```

### Region Parameter

The `-sweep` flag specifies the region. While Clever Cloud has multiple regions (par, rbx, wsw, etc.), the sweeper uses the API to list and delete resources across all regions within the specified organization, so the region parameter is primarily used for organization purposes.

## Requirements

Before running sweepers:

1. **Authentication**: Ensure you have valid Clever Cloud credentials configured:
   - `CLEVER_TOKEN` and `CLEVER_SECRET` environment variables, or
   - `~/.config/clever-cloud` configuration file

2. **Organization**: Set the `ORGANISATION` environment variable to your test organization ID

Example:

```bash
export ORGANISATION=orga_xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
export CLEVER_TOKEN=your_token_here
export CLEVER_SECRET=your_secret_here
```

## Makefile Integration

You can add a sweep target to your Makefile:

```makefile
.PHONY: sweep
sweep:
	@echo "Running sweepers to clean up test resources..."
	@go test ./... -v -sweep=par -sweep-run=clevercloud -timeout 30m
```

Then run:

```bash
ORGANISATION=orga_xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx make sweep
```

## Logging

Sweepers produce detailed logs showing:
- Which resources are being deleted
- Resources that are skipped (not matching the `tf-test` prefix)
- Errors encountered during deletion
- Summary of swept resources and errors

Example output:

```
[INFO] Sweeping networkgroups in organization: orga_12345678-1234-1234-1234-123456789012
[INFO] Deleting networkgroup: tf-test-ng-abc123 (ng_xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx)
[INFO] Swept 5 networkgroups (errors: 0)
[INFO] Sweeping Kubernetes clusters in organization: orga_12345678-1234-1234-1234-123456789012
[INFO] Deleting Kubernetes cluster: tf-test-k8s-cluster (kubernetes_xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx) - Status: ACTIVE
[INFO] Swept 3 Kubernetes clusters (errors: 0)
[INFO] Sweeping applications in organization: orga_12345678-1234-1234-1234-123456789012
[INFO] Deleting application: tf-test-node-xyz789 (app_xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx)
[INFO] Swept 10 applications (errors: 0)
[INFO] Sweeping addons in organization: orga_12345678-1234-1234-1234-123456789012
[INFO] Deleting addon: tf-test-pg-def456 (postgresql_xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx) - provider: postgresql-addon
[INFO] Swept 8 addons (errors: 0)
```

## Error Handling

Sweepers handle errors gracefully:
- Resources already marked for deletion are skipped
- "Not found" errors are ignored (resource already deleted)
- Other errors are logged but don't stop the sweeper
- A final count of errors is reported

## Safety Features

1. **Prefix Matching**: Only resources with `tf-test` prefix are deleted
2. **Organization Scoping**: Only resources in the specified organization are affected
3. **Error Tolerance**: Errors during deletion don't cause the sweeper to abort
4. **Dependency Order**: Resources are deleted in the correct order to respect dependencies

## Implementation Details

The sweepers are implemented using the [terraform-plugin-testing](https://github.com/hashicorp/terraform-plugin-testing) framework:

- Registered in `init()` function in [pkg/tests/sweepers.go](pkg/tests/sweepers.go)
- Use the Clever Cloud Go client for API calls
- Follow Terraform's sweeper best practices

## Troubleshooting

### Sweeper doesn't find any resources

Check that:
1. The `ORGANISATION` environment variable is set correctly
2. You have valid authentication credentials
3. Test resources actually exist with the `tf-test` prefix

### Permission errors

Ensure your API credentials have sufficient permissions to:
- List resources in the organization
- Delete resources in the organization

### Resources not being deleted

Check the logs for specific error messages. Common issues:
- Resources might be in use or have dependencies
- API rate limiting
- Network connectivity issues

## See Also

- [Terraform Plugin Testing - Sweepers](https://developer.hashicorp.com/terraform/plugin/testing/acceptance-tests/sweepers)
- [Project README](README.md)
- [CLAUDE.md - Development Guide](CLAUDE.md)
