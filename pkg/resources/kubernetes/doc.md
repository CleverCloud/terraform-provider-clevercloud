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

### Kubernetes Cluster with node autoscaling

Node auto-provisioning, powered by [Karpenter](https://karpenter.sh/), creates and
deletes nodes on demand instead of relying on statically sized node groups.

```terraform
resource "clevercloud_kubernetes" "my_cluster" {
  name                  = "my-kubernetes-cluster"
  node_autoprovisioning = true
}
```

Enabling this attribute only installs Karpenter in the cluster. Nothing is
provisioned until you deploy `NodePool` and `CleverNodeClass` custom resources
yourself: they are Kubernetes objects, managed with `kubectl` or a Kubernetes
provider, not Clever Cloud resources, and this provider does not manage them.

Removing the attribute from your configuration is equivalent to setting it to
`false`, which uninstalls Karpenter. The API refuses to uninstall it while any
`NodePool`, `NodeClaim`, `NodeOverlay` or `CleverNodeClass` still exists, so
delete them first.

Installing and removing Karpenter takes time, and the API only reports the
feature once the installation is complete, so `terraform apply` waits for it. A
refresh that happens during that window can show a diff which disappears once the
installation completes. Should the installation or the removal fail, the whole
cluster is left in a failed state and has to be resumed on the Clever Cloud side
before it accepts any other change.

Karpenter requires Kubernetes 1.34 or later, and the API does not check it for
you. It is also mutually exclusive with the cluster wide `autoscalingEnabled`
feature, which this provider does not expose: a cluster carrying it is refused
until that feature is disabled.
