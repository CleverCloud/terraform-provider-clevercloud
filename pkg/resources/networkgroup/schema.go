package networkgroup

import (
	"context"
	_ "embed"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"go.clever-cloud.dev/sdk/models"

	"go.clever-cloud.com/terraform-provider/pkg"
)

type Networkgroup struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Tags        types.Set    `tfsdk:"tags"`
	Network     types.String `tfsdk:"network"`
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
			"tags":        schema.SetAttribute{ElementType: types.StringType, Optional: true, MarkdownDescription: "Tags of the network group"},
			"network":     schema.StringAttribute{Computed: true, MarkdownDescription: "Network CIDR of the network group"},
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

// The API-to-state mappers for clevercloud_networkgroup follow.
// See CONTRIBUTING.md § "API → state mapping".

// FromNetworkGroup maps the network-group view.
//
// For the optional fields (description, tags), the rule is:
//   - state null + API empty → keep null (user never set it, nothing to sync)
//   - otherwise              → sync from API (state was set, or API returned a value)
func (ng *Networkgroup) FromNetworkGroup(ctx context.Context, api *models.NetworkGroup1, diags *diag.Diagnostics) *Networkgroup {
	if ng == nil || api == nil {
		return ng
	}

	ng.Name = pkg.FromStrMaxLen(api.Label)

	apiDescEmpty := api.Description == nil || *api.Description == ""
	if !ng.Description.IsNull() || !apiDescEmpty {
		ng.Description = basetypes.NewStringPointerValue(api.Description)
	}

	apiTagsEmpty := len(api.Tags) == 0
	if !ng.Tags.IsNull() || !apiTagsEmpty {
		ng.Tags = pkg.FromSetString(api.Tags, diags)
	}

	ng.Network = pkg.FromStr(api.NetworkIP)

	return ng
}
