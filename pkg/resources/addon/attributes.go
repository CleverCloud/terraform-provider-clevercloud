package addon

import (
	"context"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg"
)

type CommonAttributes struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Plan         types.String `tfsdk:"plan"`
	Region       types.String `tfsdk:"region"`
	CreationDate types.Int64  `tfsdk:"creation_date"`
}

var addonCommon = map[string]schema.Attribute{
	"id":   schema.StringAttribute{Computed: true, MarkdownDescription: "Generated unique identifier", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
	"name": schema.StringAttribute{Required: true, MarkdownDescription: "Name of the service"},
	"plan": schema.StringAttribute{
		Required:            true,
		MarkdownDescription: "Database size and spec",
		Validators:          []validator.String{slugValidator},
	},
	"region": schema.StringAttribute{
		Optional:            true,
		Computed:            true,
		Default:             stringdefault.StaticString("par"),
		MarkdownDescription: "Geographical region where the data will be stored",
	},
	"creation_date": schema.Int64Attribute{Computed: true, MarkdownDescription: "Date of database creation", PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()}},
}

func WithAddonCommons(runtimeSpecifics map[string]schema.Attribute) map[string]schema.Attribute {
	return pkg.Merge(addonCommon, runtimeSpecifics)
}

// https://regex101.com/r/bMOotf/1
var slugRegex = regexp.MustCompile(`^[a-z1-9_]*$`)
var slugValidator = pkg.NewStringValidator(
	"expect slug value",
	func(ctx context.Context, req validator.StringRequest, res *validator.StringResponse) {
		if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
			return
		}

		if !slugRegex.MatchString(req.ConfigValue.ValueString()) {
			res.Diagnostics.AddAttributeError(req.Path, "expect lowercase and underscore characters only", "Invalid slug, expect something like `xs_tny`")
		}
	})
