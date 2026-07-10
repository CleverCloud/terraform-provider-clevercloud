package elasticsearch_cluster

import (
	"context"
	"fmt"
	"strings"
	"time"

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
	Plan           string             `json:"plan"`
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
	Nodes          []apiNode  `json:"nodes"`
	Plan           string     `json:"plan"`
	Version        apiVersion `json:"version"`
	NetworkGroupID string     `json:"networkGroupId"`
}

// apiCredentials is returned by the dedicated /credentials endpoint. The
// password is no longer served on the cluster GET.
type apiCredentials struct {
	Endpoint string `json:"endpoint"`
	Username string `json:"username"`
	Password string `json:"password"`
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
	state.NetworkGroupID = pkg.FromStr(cluster.NetworkGroupID)
	state.Version = versionFromAPI(cluster.Version)
	state.NodeCount = pkg.FromI(int64(len(cluster.Nodes)))

	// The API may not echo the plan back; preserve the configured value if so.
	if cluster.Plan != "" {
		state.Plan = pkg.FromStr(cluster.Plan)
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

const versionsPath = "/v4/elasticsearch/versions"

func (r *ResourceElasticsearchCluster) fetchAvailableVersions(ctx context.Context) ([]apiVersion, error) {
	res := client.Get[[]apiVersion](ctx, r.Client(), versionsPath)
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

const plansPath = "/v4/elasticsearch/plans"

type apiPlan struct {
	Name     string `json:"name"`
	CPU      int64  `json:"cpu"`
	MemoryMB int64  `json:"memoryMB"`
	DiskMB   int64  `json:"diskMB"`
}

func (r *ResourceElasticsearchCluster) fetchAvailablePlans(ctx context.Context) ([]apiPlan, error) {
	res := client.Get[[]apiPlan](ctx, r.Client(), plansPath)
	if res.HasError() {
		return nil, res.Error()
	}
	return *res.Payload(), nil
}

func validatePlanAgainstAvailable(requested string, available []apiPlan) string {
	names := make([]string, len(available))
	for i, p := range available {
		if p.Name == requested {
			return ""
		}
		names[i] = p.Name
	}
	return fmt.Sprintf("plan %q is not available, supported plans: %s", requested, strings.Join(names, ", "))
}

func clusterPath(orgID string) string {
	return fmt.Sprintf(basePath, orgID)
}

func clusterIDPath(orgID, clusterID string) string {
	return fmt.Sprintf(basePath+"/%s", orgID, clusterID)
}

func credentialsPath(orgID, clusterID string) string {
	return fmt.Sprintf(basePath+"/%s/credentials", orgID, clusterID)
}

func (r *ResourceElasticsearchCluster) fetchCredentials(ctx context.Context, clusterID string) (*apiCredentials, error) {
	res := client.Get[apiCredentials](ctx, r.Client(), credentialsPath(r.Organization(), clusterID))
	if res.HasError() {
		return nil, res.Error()
	}
	return res.Payload(), nil
}

// applyCredentials copies the connection details returned by /credentials into
// the state, only overriding fields the endpoint actually populates.
func applyCredentials(c *apiCredentials, state *ElasticsearchCluster) {
	if c.Endpoint != "" {
		state.Endpoint = pkg.FromStr(c.Endpoint)
	}
	if c.Username != "" {
		state.Username = pkg.FromStr(c.Username)
	}
	if c.Password != "" {
		state.Password = pkg.FromStr(c.Password)
	}
}

func connectionReady(s *ElasticsearchCluster) bool {
	return s.Endpoint.ValueString() != "" && s.Username.ValueString() != "" && s.Password.ValueString() != ""
}

func (r *ResourceElasticsearchCluster) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, res *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() {
		return
	}

	plan := helper.From[ElasticsearchCluster](ctx, req.Plan, &res.Diagnostics)
	if res.Diagnostics.HasError() {
		return
	}

	if !plan.Version.IsNull() && !plan.Version.IsUnknown() {
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

	if !plan.Plan.IsNull() && !plan.Plan.IsUnknown() {
		plans, err := r.fetchAvailablePlans(ctx)
		if err != nil {
			res.Diagnostics.AddError("failed to fetch available Elasticsearch plans", err.Error())
			return
		}

		if msg := validatePlanAgainstAvailable(plan.Plan.ValueString(), plans); msg != "" {
			res.Diagnostics.AddError("Invalid Elasticsearch plan", msg)
		}
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
		Plan:          plan.Plan.ValueString(),
	}
	pkg.IfIsSetStr(plan.NetworkGroupID, func(s string) { body.NetworkGroupID = &s })

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "ElasticsearchCluster CREATE", map[string]any{"name": body.Name})

	res := client.Post[apiClusterResponse](ctx, r.Client(), clusterPath(r.Organization()), body)
	if res.HasError() {
		resp.Diagnostics.AddError("failed to create elasticsearch cluster", res.Error().Error())
		return
	}

	cluster := res.Payload()
	clusterID := cluster.ID
	stateFromAPI(cluster, &plan)

	// Poll until connection details are populated (the cluster may still be
	// provisioning). The password lives on a dedicated /credentials endpoint.
	for range 60 {
		if connectionReady(&plan) {
			break
		}

		time.Sleep(10 * time.Second)

		getRes := client.Get[apiClusterResponse](ctx, r.Client(), clusterIDPath(r.Organization(), clusterID))
		if getRes.HasError() {
			tflog.Debug(ctx, "ElasticsearchCluster poll error, retrying...", map[string]any{"error": getRes.Error().Error()})
			continue
		}
		cluster = getRes.Payload()
		stateFromAPI(cluster, &plan)

		creds, err := r.fetchCredentials(ctx, clusterID)
		if err != nil {
			tflog.Debug(ctx, "ElasticsearchCluster credentials not ready, retrying...", map[string]any{"error": err.Error()})
			continue
		}
		applyCredentials(creds, &plan)
	}

	if !connectionReady(&plan) {
		resp.Diagnostics.AddError(
			"elasticsearch cluster provisioning timeout",
			"connection details (endpoint, username, password) were not available after 10 minutes",
		)
		return
	}

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

	res := client.Get[apiClusterResponse](ctx, r.Client(), clusterIDPath(r.Organization(), state.ID.ValueString()))
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

	creds, err := r.fetchCredentials(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("failed to read elasticsearch cluster credentials", err.Error())
		return
	}
	applyCredentials(creds, &state)

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

	res := client.Delete[client.Nothing](ctx, r.Client(), clusterIDPath(r.Organization(), state.ID.ValueString()))
	if res.HasError() && !res.IsNotFoundError() {
		resp.Diagnostics.AddError("failed to delete elasticsearch cluster", res.Error().Error())
		return
	}

	resp.State.RemoveResource(ctx)
}
