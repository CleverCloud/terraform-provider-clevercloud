package networkgroup

import (
	"context"
	_ "embed"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"go.clever-cloud.com/terraform-provider/pkg"
)

type Networkgroup struct {
	ID                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	Description       types.String `tfsdk:"description"`
	Tags              types.Set    `tfsdk:"tags"`
	TagsAll           types.Set    `tfsdk:"tags_all"`
	IgnoreDefaultTags types.Bool   `tfsdk:"ignore_default_tags"`
	Network           types.String `tfsdk:"network"`
}

//go:embed doc.md
var resourcePostgresqlDoc string

func (r ResourceNG) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:             0,
		MarkdownDescription: resourcePostgresqlDoc,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{Computed: true, MarkdownDescription: "Generated unique identifier"},
			"name": schema.StringAttribute{Required: true, MarkdownDescription: "Name of the network group", Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
				validateLabel,
			}},
			"description": schema.StringAttribute{Optional: true, MarkdownDescription: "Description of the network group"},
			"tags": schema.SetAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "Tags of the network group. Network groups cannot be updated in place, so changing tags replaces the resource.",
				PlanModifiers:       []planmodifier.Set{setplanmodifier.RequiresReplace()},
			},
			"tags_all": schema.SetAttribute{ElementType: types.StringType, Computed: true, MarkdownDescription: "All tags applied to the network group: the resource `tags` merged with the provider-level `default_tags` (unless `ignore_default_tags` is set)"},
			"ignore_default_tags": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "When true, the provider-level `default_tags` are not merged into this network group. Changing this replaces the resource.",
				PlanModifiers:       []planmodifier.Bool{boolplanmodifier.RequiresReplace()},
			},
			"network": schema.StringAttribute{Computed: true, MarkdownDescription: "Network CIDR of the network group"},
		},
	}
}

// val result = name.value.trim.replaceAll(",", "-").replaceAll(" ", "-").replaceAll("\\.", "-").replaceAll("_", "-")
var validateLabel = pkg.NewStringValidator(
	"Validate label property",
	func(ctx context.Context, request validator.StringRequest, response *validator.StringResponse) {
		value := request.ConfigValue

		if value.IsNull() || value.IsUnknown() {
			return
		}
		if strings.Contains(value.ValueString(), ",") ||
			strings.Contains(value.ValueString(), " ") ||
			strings.Contains(value.ValueString(), ".") ||
			strings.Contains(value.ValueString(), "_") {

			response.Diagnostics.AddError(
				"Label must not have characters ',' or ' ' or '.' or '_'",
				"label is used as prefix for DNS resolution inside the Networkgroup")
		}

	},
)
