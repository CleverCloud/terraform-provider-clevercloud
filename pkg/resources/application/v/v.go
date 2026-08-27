package v

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/resources/application"
)

type ResourceV struct {
	application.Configurer[*V]
}

func NewResourceV() resource.Resource {
	return &ResourceV{}
}

func (r *ResourceV) Metadata(ctx context.Context, req resource.MetadataRequest, res *resource.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_v"
}

// UpgradeState implements state migration from version 0 to 1.
//
// v was created after vhosts became {fqdn, path_begin} objects, so a version 0
// state already has the current shape and only lacks the attributes added since
// (exposed_environment in 1.8.0, integrations in 1.9.0, redirection in 2.1.0). The
// framework decodes the raw state against the current schema, ignores attributes
// it does not know and leaves missing ones null, so nothing else needs to change.
//
// This holds as long as no existing attribute changes type. If one does in a later
// schema version, snapshot this schema as it was before the change.
func (r *ResourceV) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		0: {
			PriorSchema: &schemaV,
			StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, res *resource.UpgradeStateResponse) {
				tflog.Info(ctx, "Upgrading V resource state from version 0 to 1")

				state := helper.StateFrom[V](ctx, *req.State, &res.Diagnostics)
				if res.Diagnostics.HasError() {
					return
				}

				res.Diagnostics.Append(res.State.Set(ctx, state)...)
			},
		},
	}
}

func (r *ResourceV) GetVariantSlug() string {
	return "v"
}
