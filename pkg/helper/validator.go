package helper

import (
	"context"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"go.clever-cloud.com/terraform-provider/pkg"
)

// https://regex101.com/r/bMOotf/1
var planRegex = regexp.MustCompile(`^[a-zA-Z_]*$`)
var CCPlanFlavorValidator = pkg.NewStringValidator(
	"Expect CleverCloud plan flavor only",
	func(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
		if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
			return
		}

		if !planRegex.MatchString(req.ConfigValue.ValueString()) {
			resp.Diagnostics.AddAttributeError(
				req.Path,
				"Invalid value",
				"Expect letters and underscores only",
			)
		}

		// TODO: check if plan flavor exists in CC plan flavor list
	})
