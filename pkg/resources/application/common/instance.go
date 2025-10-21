package common

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
	"go.clever-cloud.dev/client"
)

// Lookup for the instance matching this criteria
// return the
func LookupInstanceByVariantSlug(ctx context.Context, cc *client.Client, ownerId *string, variantSlug string, diags *diag.Diagnostics) *tmp.ProductInstance {

	productRes := tmp.GetProductInstance(ctx, cc, ownerId)
	if productRes.HasError() {
		diags.AddError("failed to get variant", productRes.Error().Error())
		return nil
	}

	instances := *productRes.Payload()

	instance := pkg.First(instances, func(instance tmp.ProductInstance) bool {
		return strings.EqualFold(instance.Variant.Slug, variantSlug)
	})
	if instance == nil {
		diags.AddError("failed to get instance", fmt.Sprintf("there id no product matching variant slug '%s'", variantSlug))
		return nil
	}

	return instance
}

func lastVariant(variants []tmp.ProductInstance) tmp.ProductInstance {

	sort.SliceStable(variants, func(i, j int) bool {
		return variants[i].Version > variants[j].Version
	})

	return variants[0]
}
