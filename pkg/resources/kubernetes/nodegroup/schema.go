package nodegroup

import (
	"context"
	_ "embed"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type KubernetesNodegroup struct {
	ID           types.String `tfsdk:"id"`
	KubernetesID types.String `tfsdk:"kubernetes_id"`
	Name         types.String `tfsdk:"name"`
	Flavor       types.String `tfsdk:"flavor"`
	Size         types.Int64  `tfsdk:"size"`
}

type KubernetesNodegroupIdentity struct {
	ID types.String `tfsdk:"id"`
}

//go:embed doc.md
var resourceKubernetesNodegroupDoc string

func (r ResourceKubernetesNodegroup) IdentitySchema(_ context.Context, req resource.IdentitySchemaRequest, res *resource.IdentitySchemaResponse) {
	res.IdentitySchema = identityschema.Schema{
		Attributes: map[string]identityschema.Attribute{
			"id": identityschema.StringAttribute{RequiredForImport: true},
		},
	}
}

func (r ResourceKubernetesNodegroup) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:             0,
		MarkdownDescription: resourceKubernetesNodegroupDoc,
		Attributes: map[string]schema.Attribute{
			"id":            schema.StringAttribute{Computed: true, MarkdownDescription: "Generated unique identifier", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"kubernetes_id": schema.StringAttribute{Required: true},
			"name":          schema.StringAttribute{Required: true, MarkdownDescription: "Name of the node group"},
			"flavor":        schema.StringAttribute{Required: true, Description: "flavor of nodes"},
			"size":          schema.Int64Attribute{Required: true, Description: "count of nodes"}, // add validator with min=0, max=16
		},
	}
}
