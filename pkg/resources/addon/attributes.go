package addon

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg"
)

type CommonAttributes struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	Plan          types.String `tfsdk:"plan"`
	Region        types.String `tfsdk:"region"`
	CreationDate  types.Int64  `tfsdk:"creation_date"`
	Networkgroups types.Set    `tfsdk:"networkgroups"`
}

var addonCommon = map[string]schema.Attribute{
	"id":   schema.StringAttribute{Computed: true, MarkdownDescription: "Generated unique identifier", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
	"name": schema.StringAttribute{Required: true, MarkdownDescription: "Name of the service"},
	"plan": schema.StringAttribute{Required: true, MarkdownDescription: "Database size and spec"},
	"region": schema.StringAttribute{
		Optional:            true,
		Computed:            true,
		Default:             stringdefault.StaticString("par"),
		MarkdownDescription: "Geographical region where the data will be stored",
	},
	"creation_date": schema.Int64Attribute{Computed: true, MarkdownDescription: "Date of database creation", PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()}},
	"networkgroups": schema.SetNestedAttribute{
		Optional:            true,
		MarkdownDescription: "List of networkgroups the addon must be part of",
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"networkgroup_id": schema.StringAttribute{
					Required:            true,
					MarkdownDescription: "ID of the networkgroup",
				},
				"fqdn": schema.StringAttribute{
					Required:            true,
					MarkdownDescription: "domain name which will resolve to addon instances inside the networkgroup",
				},
			},
		},
	},
}

func WithAddonCommons(runtimeSpecifics map[string]schema.Attribute) map[string]schema.Attribute {
	return pkg.Merge(addonCommon, runtimeSpecifics)
}
