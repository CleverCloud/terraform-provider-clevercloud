Manage [Clever Kubernetes Engine](https://www.clever.cloud/developers/doc/kubernetes/) clusters.

See [Kubernetes product specification](https://www.clever.cloud/developers/doc/kubernetes/).

## Example Usage

### Kubernetes Cluster

```terraform
resource "clevercloud_kubernetes" "my_cluster" {
  name = "my-kubernetes-cluster"
}
```

### Kubernetes Cluster with a Node Group

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
