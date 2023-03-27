package application

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
	"go.clever-cloud.dev/client"
)

// Lookup for the instance matching this criteria
// return the
func LookupInstance(ctx context.Context, cc *client.Client, kind, name string, diags diag.Diagnostics) *tmp.ProductInstance {

	productRes := tmp.GetProductInstance(ctx, cc)
	if productRes.HasError() {
		diags.AddError("failed to get variant", productRes.Error().Error())
		return nil
	}

	instances := *productRes.Payload()

	instanceKind := pkg.Filter(instances, func(instance tmp.ProductInstance) bool {
		return instance.Type == kind && instance.Enabled
	})

	if len(instanceKind) == 0 {
		diags.AddError("failed to get variant", fmt.Sprintf("there id no product matching type '%s'", kind))
		return nil
	}

	variants := pkg.Filter(instanceKind, func(instance tmp.ProductInstance) bool {
		return instance.Name == name
	})

	if len(variants) == 0 {
		diags.AddError("failed to get variant", fmt.Sprintf("there id no product matching this name '%s'", name))
		return nil
	} else if len(variants) > 1 {
		diags.AddWarning("failed to get the right variant", "more than one variant match this criteria, take last one")
	}

	variant := variants[len(variants)-1]
	return &variant
}
