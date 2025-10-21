package bucket

import (
	"context"
	_ "embed"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type CellarBucket struct {
	// Should be name, but ID is mandatory for now
	// https://github.com/hashicorp/terraform-plugin-testing/issues/84
	// TODO: Name instead of ID when issue is resolved
	Name     types.String `tfsdk:"id"`
	CellarID types.String `tfsdk:"cellar_id"`
}

//go:embed doc.md
var resourceCellarBucketDoc string

func (r ResourceCellarBucket) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:             0,
		MarkdownDescription: resourceCellarBucketDoc,
		Attributes: map[string]schema.Attribute{
			// customer provided
			"id":        schema.StringAttribute{Required: true, MarkdownDescription: "Name of the bucket"},
			"cellar_id": schema.StringAttribute{Required: true, MarkdownDescription: "Cellar's reference"},
		},
	}
}
