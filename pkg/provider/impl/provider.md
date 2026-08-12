CleverCloud provider allow you to interact with CleverCloud platform.

## Dedicated OAuth consumer

If you want to use a dedicated OAuth consumer and the acording user tokens, 
be sure the next rights are granted

```
# EN
Access my personal information
Access my organizations
Manage my organizations
Manage my organizations's applications
Manage my organizations's add-ons

# FR
Accéder à mes informations personnelles
Accéder à mes organisations
Gérer mes organisations
Gérer les applications de mes organisations
Gérer les add-ons de mes organisations
```

## Applications: private repository deployment

To deploy from a private GitHub repository, you need to generate a Personal Access Token (PAT) that will be used for authentication.

### Creating a GitHub Personal Access Token

1. Navigate to GitHub Settings → Developer settings → [Personal access tokens](https://github.com/settings/tokens)
2. [Create a fine-grained token](https://github.com/settings/personal-access-tokens/new) with read access to repository contents
3. **Best practice**: Limit the token to only the specific repositories you want to deploy (use "Only select repositories" instead of "All repositories")

![Example PAT configuration](./pat_example.png)

For detailed instructions, refer to [GitHub's documentation on creating a fine-grained personal access token](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/managing-your-personal-access-tokens#creating-a-fine-grained-personal-access-token)

### Using the Token in Terraform

Once the token is generated, add the `authentication_basic` attribute to the `deployment` block of your application resource:

```hcl
resource "clevercloud_nodejs" "my_app" {
  # ... other configuration ...

  deployment {
    repository = "https://github.com/OWNER/REPO.git"
    authentication_basic = "USER:PAT_TOKEN"
  }
}
```

Where:
- `USER` is the GitHub username of the person who created the token
- `PAT_TOKEN` is the Personal Access Token generated in the previous step

## Store the Terraform state on Cellar

A [Cellar](https://www.clever.cloud/developers/doc/addons/cellar/) bucket can store your Terraform state through the [S3 backend](https://developer.hashicorp.com/terraform/language/backend/s3). The backend must exist before `terraform init`, so use a Cellar add-on and a bucket created beforehand (from the [Console](https://console.clever-cloud.com/) or the CLI):

```terraform
terraform {
  backend "s3" {
    bucket = "my-terraform-state-bucket"
    key    = "my-project.tfstate"
    region = "us-east-1"
    endpoints = {
      s3 = "https://cellar-c2.services.clever-cloud.com"
    }

    use_path_style              = true
    skip_region_validation      = true
    skip_credentials_validation = true
    skip_metadata_api_check     = true
    skip_requesting_account_id  = true
  }
}
```

Credentials are read from the standard AWS environment variables, filled with the Cellar add-on ones:

```bash
export AWS_ACCESS_KEY_ID="$CELLAR_ADDON_KEY_ID"
export AWS_SECRET_ACCESS_KEY="$CELLAR_ADDON_KEY_SECRET"
```

~> Recent Terraform versions upload the state with a trailing checksum
(`x-amz-content-sha256: STREAMING-UNSIGNED-PAYLOAD-TRAILER`), a feature Cellar does not support yet:
every state write fails with `Error: Failed to save state [...] api error XAmzContentSHA256Mismatch`
(reads are not affected). The `skip_s3_checksum` backend option has no effect on this behaviour.
Until Cellar supports trailing checksums, disable them through the AWS SDK environment variable:

```bash
export AWS_REQUEST_CHECKSUM_CALCULATION=when_required
```