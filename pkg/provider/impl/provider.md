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