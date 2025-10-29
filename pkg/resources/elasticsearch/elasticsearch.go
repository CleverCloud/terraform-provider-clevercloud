package elasticsearch

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/miton18/helper/set"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
)

type ResourceElasticsearch struct {
	helper.Configurer
	versions *set.Set[string]
}

func NewResourceElasticsearch() resource.Resource {
	return &ResourceElasticsearch{
		versions: set.New[string](),
	}
}

func (r *ResourceElasticsearch) Metadata(ctx context.Context, req resource.MetadataRequest, res *resource.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_elasticsearch"
}

type ElasticsearchIdentity struct {
	ID types.String `tfsdk:"id"`
}

func (r *ResourceElasticsearch) IdentitySchema(ctx context.Context, req resource.IdentitySchemaRequest, res *resource.IdentitySchemaResponse) {
	res.IdentitySchema = identityschema.Schema{
		Attributes: map[string]identityschema.Attribute{
			"id": identityschema.StringAttribute{
				RequiredForImport: true,
				Description:       "Unique identifier of the Elasticsearch addon",
			},
		},
	}
}
