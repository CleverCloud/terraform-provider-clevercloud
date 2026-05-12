package elasticsearch_cluster

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.dev/client"
)

const basePath = "/v4/elasticsearch/organisations/%s/clusters"

type apiVersion struct {
	Major int64 `json:"major"`
	Minor int64 `json:"minor"`
	Patch int64 `json:"patch"`
}

type apiVersionRequest struct {
	Major *int64 `json:"major"`
	Minor *int64 `json:"minor"`
	Patch *int64 `json:"patch"`
}

type apiCreateRequest struct {
	Name           string             `json:"name"`
	Version        *apiVersionRequest `json:"version"`
	NumberOfNodes  int64              `json:"numberOfNodes"`
	NodeCPU        int64              `json:"nodeCPU"`
	NodeMemoryMB   int64              `json:"nodeMemoryMB"`
	NodeDiskMB     int64              `json:"nodeDiskMB"`
	NetworkGroupID *string            `json:"networkGroupId,omitempty"`
}

type apiNode struct {
	ID       string  `json:"id"`
	CPU      int64   `json:"cpu"`
	MemoryMB float64 `json:"memoryMB"`
	DiskMB   float64 `json:"diskMB"`
}

type apiClusterResponse struct {
	ID             string     `json:"id"`
	Name           string     `json:"name"`
	Endpoint       string     `json:"endpoint"`
	Username       string     `json:"username"`
	Password       string     `json:"password"`
	Nodes          []apiNode  `json:"nodes"`
	Version        apiVersion `json:"version"`
	NetworkGroupID string     `json:"networkGroupId"`
}

func versionFromAPI(v apiVersion) types.Object {
	ver := Version{
		Major: pkg.FromI(v.Major),
		Minor: pkg.FromI(v.Minor),
		Patch: pkg.FromI(v.Patch),
	}
	obj, _ := types.ObjectValueFrom(context.Background(), versionAttrTypes, ver)
	return obj
}

func stateFromAPI(cluster *apiClusterResponse, state *ElasticsearchCluster) {
	state.ID = pkg.FromStr(cluster.ID)
	state.Name = pkg.FromStr(cluster.Name)
	state.Endpoint = pkg.FromStr(cluster.Endpoint)
	state.Username = pkg.FromStr(cluster.Username)
	state.Password = pkg.FromStr(cluster.Password)
	state.NetworkGroupID = pkg.FromStr(cluster.NetworkGroupID)
	state.Version = versionFromAPI(cluster.Version)
	state.NodeCount = pkg.FromI(int64(len(cluster.Nodes)))

	if len(cluster.Nodes) > 0 {
		node := cluster.Nodes[0]
		state.CPUCount = pkg.FromI(node.CPU)
		state.MemorySize = pkg.FromI(int64(node.MemoryMB))
		state.DiskSize = pkg.FromI(int64(node.DiskMB))
	}
}

func versionToAPI(ctx context.Context, obj types.Object, diags *diag.Diagnostics) *apiVersionRequest {
	av := &apiVersionRequest{}

	if obj.IsNull() || obj.IsUnknown() {
		return av
	}

	var v Version
	d := obj.As(ctx, &v, basetypes.ObjectAsOptions{})
	diags.Append(d...)
	if diags.HasError() {
		return av
	}

	pkg.IfIsSetI(v.Major, func(i int64) { av.Major = &i })
	pkg.IfIsSetI(v.Minor, func(i int64) { av.Minor = &i })
	pkg.IfIsSetI(v.Patch, func(i int64) { av.Patch = &i })
	return av
}

const versionsPath = "/v4/elasticsearch/organisations/%s/versions"

func (r *ResourceElasticsearchCluster) fetchAvailableVersions(ctx context.Context) ([]apiVersion, error) {
	path := fmt.Sprintf(versionsPath, r.Organization())
	res := client.Get[[]apiVersion](ctx, r.esClient(), path)
	if res.HasError() {
		return nil, res.Error()
	}
	return *res.Payload(), nil
}

func validateVersionAgainstAvailable(requested *apiVersionRequest, available []apiVersion) string {
	if requested == nil {
		return ""
	}

	for _, v := range available {
		if requested.Major != nil && *requested.Major != v.Major {
			continue
		}
		if requested.Minor != nil && *requested.Minor != v.Minor {
			continue
		}
		if requested.Patch != nil && *requested.Patch != v.Patch {
			continue
		}
		return ""
	}

	parts := []string{}
	if requested.Major != nil {
		parts = append(parts, fmt.Sprintf("%d", *requested.Major))
	}
	if requested.Minor != nil {
		parts = append(parts, fmt.Sprintf("%d", *requested.Minor))
	}
	if requested.Patch != nil {
		parts = append(parts, fmt.Sprintf("%d", *requested.Patch))
	}

	avail := make([]string, len(available))
	for i, v := range available {
		avail[i] = fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
	}

	return fmt.Sprintf(
		"version %s is not available, supported versions: %s",
		strings.Join(parts, "."),
		strings.Join(avail, ", "),
	)
}

func clusterPath(orgID string) string {
	return fmt.Sprintf(basePath, orgID)
}

func clusterIDPath(orgID, clusterID string) string {
	return fmt.Sprintf(basePath+"/%s", orgID, clusterID)
}

// esClient returns a client targeting the Elasticsearch API.
// TODO: remove once the API is served from the main API domain
func (r *ResourceElasticsearchCluster) esClient() *client.Client {
	endpoint := os.Getenv("CC_ES_API_ENDPOINT")
	if endpoint == "" {
		return r.Client()
	}
	return client.New(
		client.WithEndpoint(endpoint),
		client.WithAuthenticator(r.Client().Authenticator()),
	)
}

func (r *ResourceElasticsearchCluster) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, res *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() {
		return
	}

	plan := helper.From[ElasticsearchCluster](ctx, req.Plan, &res.Diagnostics)
	if res.Diagnostics.HasError() {
		return
	}

	if plan.Version.IsNull() || plan.Version.IsUnknown() {
		return
	}

	version := versionToAPI(ctx, plan.Version, &res.Diagnostics)
	if res.Diagnostics.HasError() {
		return
	}

	available, err := r.fetchAvailableVersions(ctx)
	if err != nil {
		res.Diagnostics.AddError("failed to fetch available Elasticsearch versions", err.Error())
		return
	}

	if msg := validateVersionAgainstAvailable(version, available); msg != "" {
		res.Diagnostics.AddError("Invalid Elasticsearch version", msg)
	}
}

func (r *ResourceElasticsearchCluster) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	plan := helper.PlanFrom[ElasticsearchCluster](ctx, req.Plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	body := apiCreateRequest{
		Name:          plan.Name.ValueString(),
		Version:       versionToAPI(ctx, plan.Version, &resp.Diagnostics),
		NumberOfNodes: plan.NodeCount.ValueInt64(),
		NodeCPU:       plan.CPUCount.ValueInt64(),
		NodeMemoryMB:  plan.MemorySize.ValueInt64(),
		NodeDiskMB:    plan.DiskSize.ValueInt64(),
	}
	pkg.IfIsSetStr(plan.NetworkGroupID, func(s string) { body.NetworkGroupID = &s })

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "ElasticsearchCluster CREATE", map[string]any{"name": body.Name})

	res := client.Post[apiClusterResponse](ctx, r.esClient(), clusterPath(r.Organization()), body)
	if res.HasError() {
		resp.Diagnostics.AddError("failed to create elasticsearch cluster", res.Error().Error())
		return
	}

	cluster := res.Payload()
	stateFromAPI(cluster, &plan)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Read Elasticsearch cluster information
func (r *ResourceElasticsearchCluster) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	state := helper.StateFrom[ElasticsearchCluster](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.ID.ValueString() == "" {
		resp.State.RemoveResource(ctx)
		return
	}

	tflog.Debug(ctx, "ElasticsearchCluster READ", map[string]any{"id": state.ID.ValueString()})

	res := client.Get[apiClusterResponse](ctx, r.esClient(), clusterIDPath(r.Organization(), state.ID.ValueString()))
	if res.IsNotFoundError() {
		resp.State.RemoveResource(ctx)
		return
	}
	if res.HasError() {
		resp.Diagnostics.AddError("failed to read elasticsearch cluster", res.Error().Error())
		return
	}

	cluster := res.Payload()
	stateFromAPI(cluster, &state)

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Update Elasticsearch cluster
func (r *ResourceElasticsearchCluster) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("update not supported", "elasticsearch cluster does not support in-place updates")
}

// Delete Elasticsearch cluster
func (r *ResourceElasticsearchCluster) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	state := helper.StateFrom[ElasticsearchCluster](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "ElasticsearchCluster DELETE", map[string]any{"id": state.ID.ValueString()})

	res := client.Delete[client.Nothing](ctx, r.esClient(), clusterIDPath(r.Organization(), state.ID.ValueString()))
	if res.HasError() && !res.IsNotFoundError() {
		resp.Diagnostics.AddError("failed to delete elasticsearch cluster", res.Error().Error())
		return
	}

	resp.State.RemoveResource(ctx)
}
