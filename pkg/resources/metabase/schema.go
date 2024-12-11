package metabase

import (
	"context"
	_ "embed"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg/attributes"
)

type Metabase struct {
	attributes.Addon
	Host types.String `tfsdk:"host"`
	// HerokuId     types.String `tfsdk:"heroku_id"`
	// CallbackURL  types.String `tfsdk:"callback_url"`
	// LogplexToken types.String `tfsdk:"logplex_token"`
	// OwnerId      types.String `tfsdk:"owner_id"`
}

//go:embed doc.md
var resourceMetabaseDoc string

func (r ResourceMetabase) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:             0,
		MarkdownDescription: resourceMetabaseDoc,
		Attributes: attributes.WithAddonCommons(map[string]schema.Attribute{
			// customer provided
			// TODO: Markdown description
			"host": schema.StringAttribute{Computed: true, MarkdownDescription: "Metabase host, used to connect to"},
			// "heroku_id":     schema.StringAttribute{Computed: true, MarkdownDescription: "heroku_id"},
			// "callback_url":  schema.StringAttribute{Computed: true, MarkdownDescription: "callback_url"},
			// "logplex_token": schema.StringAttribute{Computed: true, MarkdownDescription: "logplex_token"},
			// "owner_id":      schema.StringAttribute{Computed: true, MarkdownDescription: "owner_id"},
		}),
	}
}

// https://developer.hashicorp.com/terraform/plugin/framework/resources/state-upgrade#implementing-state-upgrade-support
func (r ResourceMetabase) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{}
}
