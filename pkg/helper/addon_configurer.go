package helper

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"go.clever-cloud.com/terraform-provider/pkg/addonid"
)

const addonIDPrefix = "addon_"

// AddonConfigurer is the Configurer of the add-ons identified by their real ID
// (`redis_xxx`), which is what their Create writes to the state. The generic
// clevercloud_addon is identified by its add-on ID and keeps the plain Configurer
type AddonConfigurer struct {
	Configurer
}

// ImportState refuses an add-on ID. Most of the API takes both IDs, so it would
// seem to work, but a network group member registered with an add-on ID never gets
// a peer: the add-on would silently stay out of the network group
func (c *AddonConfigurer) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if !strings.HasPrefix(req.ID, addonIDPrefix) {
		c.Configurer.ImportState(ctx, req, resp)
		return
	}

	resp.Diagnostics.AddAttributeError(
		path.Root("id"),
		"cannot import an add-on with its add-on ID",
		c.importHint(ctx, req.ID),
	)
}

// importHint tells which ID to use instead. Looking it up is only a courtesy, the
// import fails either way
func (c *AddonConfigurer) importHint(ctx context.Context, addonID string) string {
	generic := fmt.Sprintf("%s is an add-on ID, import this resource with its real ID instead (clever-tools terraform generate prints it)", addonID)
	if c.Provider == nil {
		return generic
	}

	realID, err := addonid.ToRealID(ctx, c.Client(), c.Organization(), addonID)
	if err != nil {
		return generic
	}

	return fmt.Sprintf("%s is an add-on ID, import this resource with its real ID instead: %s", addonID, realID)
}
