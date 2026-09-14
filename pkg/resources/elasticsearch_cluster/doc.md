Manage a dedicated Elasticsearch cluster on Clever Cloud.

~> **Alpha product**: not meant for production workloads, it can break or be
reset at any time. Creating clusters requires the organisation to be
allow-listed by the support.

The cluster is not exposed publicly: it is only reachable from applications
attached to its network group (`networkgroup_id`), at the host given by
`endpoint`.

Clusters have at least 3 nodes and cannot be updated in place: changing any
attribute recreates the cluster.
