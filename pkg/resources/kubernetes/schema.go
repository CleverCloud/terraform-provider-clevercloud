package kubernetes

import (
	"context"
	_ "embed"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type Kubernetes struct {
	ID types.String `tfsdk:"id"`

	Name           types.String `tfsdk:"name"`
	Region         types.String `tfsdk:"region"`
	Version        types.String `tfsdk:"version"`
	APIServerURL   types.String `tfsdk:"apiserver_url"`
	NetworkgroupID types.String `tfsdk:"networkgroup_id"`
	KubeConfig     types.String `tfsdk:"kubeconfig"`
}

//go:embed doc.md
var resourceKubernetesDoc string

func (r ResourceKubernetes) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:             0,
		MarkdownDescription: resourceKubernetesDoc,
		Attributes: map[string]schema.Attribute{
			"id":              schema.StringAttribute{Computed: true, MarkdownDescription: "Generated unique identifier", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"name":            schema.StringAttribute{Required: true, MarkdownDescription: "Name of the Kubernetes cluster"},
			"region":          schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "Geographical region where the cluster will be deployed", Default: stringdefault.StaticString("par")},
			"version":         schema.StringAttribute{Computed: true, MarkdownDescription: "Kubernetes version"},
			"kubeconfig":      schema.StringAttribute{Computed: true, MarkdownDescription: "Kubernetes configuration file content for accessing the cluster"},
			"apiserver_url":   schema.StringAttribute{Computed: true, MarkdownDescription: "Kubernetes APIServer URL"},
			"networkgroup_id": schema.StringAttribute{Computed: true, MarkdownDescription: "NetworkgGroup ID"},
		},
	}
}

// https://developer.hashicorp.com/terraform/plugin/framework/resources/state-upgrade#implementing-state-upgrade-support
func (r ResourceKubernetes) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{}
}
