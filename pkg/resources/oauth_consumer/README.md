# OAuth Consumer Resource

This resource manages OAuth consumers for Clever Cloud organizations.

## Files Structure

- `oauth_consumer.go` - Resource definition and metadata
- `schema.go` - Terraform schema with all fields and validation
- `crud.go` - CRUD operations (Create, Read, Update, Delete)
- `helpers.go` - Helper functions for rights conversion between Set and API format
- `oauth_consumer_test.go` - Acceptance tests
- `doc.md` - Documentation template for tfplugindocs

## API Integration

The resource uses the following API endpoints:

- **POST** `/v2/organisations/{orgID}/consumers` - Create consumer
- **GET** `/v2/organisations/{orgID}/consumers/{key}` - Get consumer details
- **GET** `/v2/organisations/{orgID}/consumers/{key}/secret` - Get consumer secret
- **PUT** `/v2/organisations/{orgID}/consumers/{key}` - Update consumer
- **DELETE** `/v2/organisations/{orgID}/consumers/{key}` - Delete consumer

API implementation is in [pkg/tmp/oauth_consumer.go](../../tmp/oauth_consumer.go).

## Fields

### Required
- `name` - Consumer name
- `base_url` - Base URL of the application
- `website_url` - Website URL
- `rights` - Set of OAuth permissions (forces replacement on change)

### Optional
- `description` - Consumer description
- `logo_url` - Logo/picture URL

### Computed (read-only)
- `id` - OAuth consumer key (client ID)
- `secret` - OAuth consumer secret (client secret) - **sensitive**

**Note**: The organisation is taken from the provider configuration, not from the resource.

## Available Rights

- `access_organisations`
- `access_organisations_bills`
- `access_organisations_consumption_statistics`
- `access_organisations_credit_count`
- `access_personal_information`
- `manage_organisations`
- `manage_organisations_applications`
- `manage_organisations_members`
- `manage_organisations_services`
- `manage_personal_information`
- `manage_ssh_keys`

## Testing

Run acceptance tests:

```bash
CC_ORGANISATION=orga_xxxx-xxxx-xxxx-xxxx TF_ACC=1 go test -v ./pkg/resources/oauth_consumer/... -timeout 30m
```

Tests include:
- Basic consumer creation and update
- Consumer with all available rights
- Secret retrieval and sensitivity

## Test Cleanup (Sweeper)

A sweeper is configured to automatically clean up test OAuth consumers (those with names starting with `tf-test`).

Run the sweeper:

```bash
CC_ORGANISATION=orga_xxxx-xxxx-xxxx-xxxx make sweep SWEEPARGS="-sweep-run=clevercloud_oauth_consumer"
```

Or use the interactive script:

```bash
./scripts/sweep.sh
```

The sweeper will:
- List all OAuth consumers in the organization
- Delete only those with names starting with `tf-test`
- Skip consumers that are already deleted
- Report the number of cleaned resources and any errors
