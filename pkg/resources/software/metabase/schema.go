package metabase

import (
	"context"
	_ "embed"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg/resources"
)

type Metabase struct {
	ID     types.String `tfsdk:"id"`
	Tags   types.Set    `tfsdk:"tags"`
	Name   types.String `tfsdk:"name"`
	Region types.String `tfsdk:"region"`
	Host   types.String `tfsdk:"host"`
}

//go:embed doc.md
var resourceMetabaseDoc string

func (r ResourceMetabase) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:             0,
		MarkdownDescription: resourceMetabaseDoc,
		Attributes: map[string]schema.Attribute{
			"tags": resources.TagsAttribute,
			"id":   schema.StringAttribute{Computed: true, MarkdownDescription: "Generated unique identifier", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"name": schema.StringAttribute{Required: true, MarkdownDescription: "Name of the service"},
			"region": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("par"),
				MarkdownDescription: "Geographical region where the data will be stored",
			},
			"host": schema.StringAttribute{Computed: true, MarkdownDescription: "Metabase host, used to connect to"},
		},
	}
}
