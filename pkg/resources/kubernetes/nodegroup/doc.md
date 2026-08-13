Manage node groups of a [Clever Kubernetes Engine](https://www.clever.cloud/developers/doc/kubernetes/) cluster.

See [Kubernetes product specification](https://www.clever.cloud/developers/doc/kubernetes/).

## Example Usage

```terraform
resource "clevercloud_kubernetes" "my_cluster" {
  name = "my-kubernetes-cluster"
}

resource "clevercloud_kubernetes_nodegroup" "workers" {
  kubernetes_id = clevercloud_kubernetes.my_cluster.id
  name          = "workers"
  flavor        = "S"
  size          = 2
}
```
