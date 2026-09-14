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
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

func versionFromAPI(v tmp.ElasticsearchVersion) types.Object {
	ver := Version{
		Major: pkg.FromI(v.Major),
		Minor: pkg.FromI(v.Minor),
		Patch: pkg.FromI(v.Patch),
	}
	obj, _ := types.ObjectValueFrom(context.Background(), versionAttrTypes, ver)
	return obj
}

func stateFromAPI(cluster *tmp.ElasticsearchCluster, state *ElasticsearchCluster) {
	state.ID = pkg.FromStr(cluster.ID)
	state.Name = pkg.FromStr(cluster.Name)
	state.Username = pkg.FromStr(cluster.Username)
	state.NetworkGroupID = pkg.FromStr(cluster.NetworkGroupID)
	state.Version = versionFromAPI(cluster.Version)
	state.NodeCount = pkg.FromI(int64(len(cluster.Nodes)))

	// The plan is only echoed back per node; every node shares the same one.
	// Preserve the configured value while the nodes are not listed yet.
	for _, node := range cluster.Nodes {
		if node.Plan != nil && node.Plan.Name != "" {
			state.Plan = pkg.FromStr(node.Plan.Name)
			break
		}
	}
}

func versionToAPI(ctx context.Context, obj types.Object, diags *diag.Diagnostics) *tmp.ElasticsearchVersionRequest {
	av := &tmp.ElasticsearchVersionRequest{}

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

func validateVersionAgainstAvailable(requested *tmp.ElasticsearchVersionRequest, available []tmp.ElasticsearchVersion) string {
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

func validatePlanAgainstAvailable(requested string, available []tmp.ElasticsearchPlan) string {
	names := make([]string, len(available))
	for i, p := range available {
		if p.Name == requested {
			return ""
		}
		names[i] = p.Name
	}
	return fmt.Sprintf("plan %q is not available, supported plans: %s", requested, strings.Join(names, ", "))
}

// applyCredentials copies the connection details returned by /credentials into
// the state, only overriding fields the endpoint actually populates.
func applyCredentials(c *tmp.ElasticsearchCredentials, state *ElasticsearchCluster) {
	if c.Username != "" {
		state.Username = pkg.FromStr(c.Username)
	}
	if c.Password != "" {
		state.Password = pkg.FromStr(c.Password)
	}
}

// fetchEndpoint resolves the cluster endpoint: the cluster has no public
// address, it is only reachable through its network group, as the member
// domain name registered there.
func (r *ResourceElasticsearchCluster) fetchEndpoint(ctx context.Context, networkGroupID, clusterID string) (string, error) {
	res := r.SDK.V4().Networkgroups().Organisations().Ownerid(r.Organization()).
		Networkgroups().Networkgroupid(networkGroupID).
		Members().Listnetworkgroupmembers(ctx)
	if res.HasError() {
		return "", res.Error()
	}

	for _, member := range *res.Payload() {
		if member.ID == clusterID {
			return member.DomainName, nil
		}
	}

	return "", fmt.Errorf("cluster %s is not a member of network group %s", clusterID, networkGroupID)
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

		availableRes := tmp.FetchElasticsearchAvailableVersions(ctx, r.Client())
		if availableRes.HasError() {
			res.Diagnostics.AddError("failed to fetch available Elasticsearch versions", availableRes.Error().Error())
			return
		}

		if msg := validateVersionAgainstAvailable(version, *availableRes.Payload()); msg != "" {
			res.Diagnostics.AddError("Invalid Elasticsearch version", msg)
		}
	}

	if !plan.Plan.IsNull() && !plan.Plan.IsUnknown() {
		plansRes := tmp.FetchElasticsearchAvailablePlans(ctx, r.Client())
		if plansRes.HasError() {
			res.Diagnostics.AddError("failed to fetch available Elasticsearch plans", plansRes.Error().Error())
			return
		}

		if msg := validatePlanAgainstAvailable(plan.Plan.ValueString(), *plansRes.Payload()); msg != "" {
			res.Diagnostics.AddError("Invalid Elasticsearch plan", msg)
		}
	}
}

func (r *ResourceElasticsearchCluster) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// remove when GA
	resp.Diagnostics.AddWarning(
		"Elasticsearch cluster product is in alpha",
		"It is not meant for production workloads: it can break or be reset at any time, use it at your own risks. "+
			"The organisation also needs to be allow-listed by the support to create clusters.",
	)

	plan := helper.PlanFrom[ElasticsearchCluster](ctx, req.Plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	body := tmp.WannabeElasticsearchCluster{
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

	res := tmp.CreateElasticsearchClusterWithRetry(ctx, r.Client(), r.Organization(), body)
	if res.HasError() {
		resp.Diagnostics.AddError("failed to create elasticsearch cluster", res.Error().Error())
		return
	}

	cluster := res.Payload()
	clusterID := cluster.ID
	stateFromAPI(cluster, &plan)

	// Poll until the cluster is deployed and its connection details are
	// populated. The password lives on a dedicated /credentials endpoint and
	// the endpoint is the cluster's network group member domain.
	for range 60 {
		if cluster.DeploymentStatus == tmp.ElasticsearchClusterDeployed && connectionReady(&plan) {
			break
		}

		time.Sleep(10 * time.Second)

		getRes := tmp.GetElasticsearchCluster(ctx, r.Client(), r.Organization(), clusterID)
		if getRes.HasError() {
			tflog.Debug(ctx, "ElasticsearchCluster poll error, retrying...", map[string]any{"error": getRes.Error().Error()})
			continue
		}
		cluster = getRes.Payload()
		stateFromAPI(cluster, &plan)
		tflog.Debug(ctx, "ElasticsearchCluster polling", map[string]any{"status": cluster.DeploymentStatus})

		if cluster.DeploymentStatus != tmp.ElasticsearchClusterDeployed || cluster.NetworkGroupID == "" {
			continue
		}

		credsRes := tmp.GetElasticsearchClusterCredentials(ctx, r.Client(), r.Organization(), clusterID)
		if credsRes.HasError() {
			tflog.Debug(ctx, "ElasticsearchCluster credentials not ready, retrying...", map[string]any{"error": credsRes.Error().Error()})
			continue
		}
		applyCredentials(credsRes.Payload(), &plan)

		endpoint, err := r.fetchEndpoint(ctx, cluster.NetworkGroupID, clusterID)
		if err != nil {
			tflog.Debug(ctx, "ElasticsearchCluster endpoint not ready, retrying...", map[string]any{"error": err.Error()})
			continue
		}
		plan.Endpoint = pkg.FromStr(endpoint)
	}

	if cluster.DeploymentStatus != tmp.ElasticsearchClusterDeployed || !connectionReady(&plan) {
		resp.Diagnostics.AddError(
			"elasticsearch cluster provisioning timeout",
			fmt.Sprintf("cluster was not deployed with its connection details (endpoint, username, password) after 10 minutes, last status: %q", cluster.DeploymentStatus),
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

	res := tmp.GetElasticsearchCluster(ctx, r.Client(), r.Organization(), state.ID.ValueString())
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

	credsRes := tmp.GetElasticsearchClusterCredentials(ctx, r.Client(), r.Organization(), state.ID.ValueString())
	if credsRes.HasError() {
		resp.Diagnostics.AddError("failed to read elasticsearch cluster credentials", credsRes.Error().Error())
		return
	}
	applyCredentials(credsRes.Payload(), &state)

	if cluster.NetworkGroupID != "" {
		endpoint, err := r.fetchEndpoint(ctx, cluster.NetworkGroupID, cluster.ID)
		if err != nil {
			resp.Diagnostics.AddError("failed to read elasticsearch cluster endpoint", err.Error())
			return
		}
		state.Endpoint = pkg.FromStr(endpoint)
	}

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

	res := tmp.DeleteElasticsearchCluster(ctx, r.Client(), r.Organization(), state.ID.ValueString())
	if res.HasError() && !res.IsNotFoundError() {
		resp.Diagnostics.AddError("failed to delete elasticsearch cluster", res.Error().Error())
		return
	}

	resp.State.RemoveResource(ctx)
}
