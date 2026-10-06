package elasticsearch_cluster

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

// mapping.go holds every API-to-state mapper for clevercloud_elasticsearch_cluster.
// See CONTRIBUTING.md § "API → state mapping".

func versionFromAPI(v tmp.ElasticsearchVersion) types.Object {
	ver := Version{
		Major: pkg.FromI(v.Major),
		Minor: pkg.FromI(v.Minor),
		Patch: pkg.FromI(v.Patch),
	}
	obj, _ := types.ObjectValueFrom(context.Background(), versionAttrTypes, ver)
	return obj
}

// FromCluster maps the cluster view.
func (c *ElasticsearchCluster) FromCluster(ctx context.Context, api *tmp.ElasticsearchCluster, diags *diag.Diagnostics) *ElasticsearchCluster {
	if c == nil || api == nil {
		return c
	}

	c.ID = pkg.FromStr(api.ID)
	c.Name = pkg.FromStr(api.Name)
	c.Username = pkg.FromStr(api.Username)
	c.NetworkGroupID = pkg.FromStr(api.NetworkGroupID)
	c.Version = versionFromAPI(api.Version)
	c.NodeCount = pkg.FromI(int64(len(api.Nodes)))

	// The plan is only echoed back per node; every node shares the same one.
	// Preserve the configured value while the nodes are not listed yet.
	for _, node := range api.Nodes {
		if node.Plan != nil && node.Plan.Name != "" {
			c.Plan = pkg.FromStr(node.Plan.Name)
			break
		}
	}

	return c
}

// FromCredentials maps the credentials call.
//
// The API answers them empty while the cluster boots, and the password is never
// returned again afterwards, so an empty value must leave state alone rather
// than clear a credential the practitioner still needs.
func (c *ElasticsearchCluster) FromCredentials(ctx context.Context, api *tmp.ElasticsearchCredentials, diags *diag.Diagnostics) *ElasticsearchCluster {
	if c == nil || api == nil {
		return c
	}

	if api.Username != "" {
		c.Username = pkg.FromStr(api.Username)
	}
	if api.Password != "" {
		c.Password = pkg.FromStr(api.Password)
	}

	return c
}
