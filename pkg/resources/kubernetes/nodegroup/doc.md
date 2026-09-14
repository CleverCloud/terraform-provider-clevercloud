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

Node groups have a fixed size. For nodes created and deleted on demand, enable
node autoscaling on the cluster with the `node_autoprovisioning` attribute of
[`clevercloud_kubernetes`](kubernetes). Both can coexist: a node group carries the
baseline capacity while auto-provisioned nodes absorb the peaks. Do not manage a
node provisioned by Karpenter with a `clevercloud_kubernetes_nodegroup` resource.
