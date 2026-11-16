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

The flavor and the cluster of a node group cannot change. Changing `flavor` or
`kubernetes_id` replaces the node group: Terraform deletes it, waits for its
nodes to be drained and removed, then creates the new one. The workloads lose
that capacity in between. The name of a node group is unique in its cluster, so
`create_before_destroy` cannot be used to avoid the gap unless the name changes
too.

Node groups have a fixed size. For nodes created and deleted on demand, enable
node autoscaling on the cluster with the `node_autoprovisioning` attribute of
[`clevercloud_kubernetes`](kubernetes). Both can coexist: a node group carries the
baseline capacity while auto-provisioned nodes absorb the peaks.

Leave node groups created by Karpenter under its control. The provider refuses
to read, update or delete a node group whose API labels contain a nonempty
`karpenter.sh/nodepool` value. It checks the current labels before updates and
deletions even when Terraform refresh is disabled. Fixed node groups without
this label remain manageable when `node_autoprovisioning` is enabled.

If such a group is already in Terraform state, remove its resource configuration
and use a `removed` block with `lifecycle { destroy = false }`, or run
`terraform state rm`, to stop tracking it without deleting it. Do not remove the
Karpenter label to bypass this protection. The check relies on labels exposed by
the API and cannot detect groups whose Karpenter labels have been removed.
