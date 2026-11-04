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

### Kubernetes Cluster with explicit options

```terraform
resource "clevercloud_kubernetes" "my_cluster" {
  name               = "my-kubernetes-cluster"
  version            = "1.36"
  persistent_storage = true
  tags               = ["env:prod"]

  topology = {
    type               = "DEDICATED_COMPUTE"
    flavor             = "S"
    replication_factor = 3
  }
}
```

The available versions, and the flavors and replication factors each topology
accepts, are decided by the platform: the provider checks them during
`terraform plan` and lists the accepted values when one is refused.

The `DISTRIBUTED` topology sizes each control plane component instead:

```terraform
resource "clevercloud_kubernetes" "my_cluster" {
  name = "my-kubernetes-cluster"

  topology = {
    type = "DISTRIBUTED"
    components = {
      apiserver                = { flavor = "S", replication_factor = 3 }
      controller_manager       = { flavor = "XS", replication_factor = 2 }
      scheduler                = { flavor = "XS", replication_factor = 2 }
      node_group_operator      = { flavor = "2XS", replication_factor = 2 }
      cloud_controller_manager = { flavor = "2XS", replication_factor = 2 }
    }
  }
}
```

Changing the topology **replaces the cluster**, the API cannot change it on an
existing one.

Changing `version` moves the cluster to that version in place, the control plane
first and then every node group, and `terraform apply` waits for the end of the
rollout. Should a node refuse to drain, the operation stops and leaves that node
cordoned. The API also accepts an older version, but Kubernetes does not officially
support downgrades: check that your workloads and stored objects allow it first.

`persistent_storage` installs the CSI driver that provides persistent volumes. The
API leaves it disabled unless it is requested, so set it to `true` if your workloads
use them. Enabling or disabling it reconciles the cluster, and `terraform apply`
waits for it.

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
