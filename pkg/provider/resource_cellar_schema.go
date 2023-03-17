package provider

import (
	"context"
	_ "embed"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type Cellar struct {
	ID types.String `tfsdk:"id"`

	Name   types.String `tfsdk:"name"`
	Region types.String `tfsdk:"region"`

	Host      types.String `tfsdk:"host"`
	KeyID     types.String `tfsdk:"key_id"`
	KeySecret types.String `tfsdk:"key_secret"`
}

//go:embed resource_cellar.md
var resourceCellarDoc string

func (r ResourceCellar) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: resourceCellarDoc,
		Attributes: map[string]schema.Attribute{
			// customer provided
			"name":   schema.StringAttribute{Required: true, MarkdownDescription: "Name of the Cellar"},
			"region": schema.StringAttribute{Required: true, MarkdownDescription: "Geographical region where the data will be stored"},

			// provider
			"id":         schema.StringAttribute{Computed: true, MarkdownDescription: "Generated unique identifier"},
			"host":       schema.StringAttribute{Computed: true, MarkdownDescription: "S3 compatible Cellar endpoint"},
			"key_id":     schema.StringAttribute{Computed: true, MarkdownDescription: "Key ID used to authenticate"},
			"key_secret": schema.StringAttribute{Computed: true, Sensitive: true, MarkdownDescription: "Key secret used to authenticate"},
		},
	}
}
