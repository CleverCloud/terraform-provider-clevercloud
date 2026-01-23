# Kubernetes Resource

Provides a Kubernetes cluster managed by Clever Cloud.

## Example Usage

```hcl
resource "clevercloud_kubernetes" "my_cluster" {
  name       = "my-kubernetes-cluster"
}

## Argument Reference

* `name` - (Required) Name of the Kubernetes cluster.

## Attribute Reference

* `id` - Unique identifier for the Kubernetes cluster.
* `kubeconfig` - Kubernetes configuration file content for accessing the cluster.

## Import

Kubernetes clusters can be imported using the ID, e.g.,

```
$ terraform import clevercloud_kubernetes.my_cluster addon_12345
```
