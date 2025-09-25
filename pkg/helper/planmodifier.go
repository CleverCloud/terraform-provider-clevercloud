package helper

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"go.clever-cloud.com/terraform-provider/pkg"
)

var UpperCaseValidator = pkg.NewStringValidator(
	"Uppercase letters only",
	func(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
		if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
			return
		}

		v := req.ConfigValue.ValueString()
		if strings.ToUpper(v) != v {
			resp.Diagnostics.AddAttributeError(
				req.Path,
				"Invalid value",
				"Expect uppercase letters only",
			)
		}
	})
