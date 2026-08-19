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

## Applications: deployment and the commit attribute

The `deployment` block drives how Terraform deploys your application, and its `commit` attribute reflects the commit currently running on it.

| configuration | behaviour |
|---|---|
| no `deployment` block | you deploy outside of Terraform (CLI, console, CI...), the provider never reconciles anything deployment-related and never shows a diff |
| block without `commit` | the repository HEAD is deployed on create/update; `commit` is computed with the running hash, and a deployment done outside of Terraform silently updates it on the next refresh |
| `commit = "<hash>"` | the commit is pinned: a deployment done outside of Terraform shows up as a diff and the next apply re-deploys the pinned hash |
| `commit = "refs/heads/..."` | the reference is resolved and deployed; the value is kept as-is in the state |
| `commit = "github_hook"` | deployments are delegated to GitHub, the provider never pushes nor reconciles the running commit |

### Known limitations

- **Import**: `terraform import` does not populate the `deployment` block (the provider cannot tell whether you want to manage deployments with Terraform). Add the block to your configuration to start tracking the running commit — be aware that the first apply then triggers a deployment.
- **Git references**: when `commit` holds a reference (`refs/heads/...`), the running commit is not reported and deployments done outside of Terraform are not detected (a local reference cannot be compared to the running hash without cloning the repository).
- **Switching from a reference to the computed commit**: removing `commit = "refs/heads/..."` from the configuration keeps the reference in the state; re-create the resource (or set an explicit hash once) to switch to the computed behaviour.
- **Repository HEAD moves, nothing else changes**: `terraform plan` does not clone the repository, so a new commit on your branch does not show up as a diff by itself; the push happens on the next apply that carries a change.
- **State freshness on update**: when an update deploys a new HEAD while `commit` is not configured, the state reports the new hash after the next refresh (Terraform requires the applied value to match the planned one).

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