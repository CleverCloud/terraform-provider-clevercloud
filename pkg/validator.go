package pkg

import (
	"context"
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type attributeValidator struct {
	desc         string
	validateFunc func(context.Context, tfsdk.ValidateAttributeRequest, *tfsdk.ValidateAttributeResponse)
}

func (av *attributeValidator) Description(context.Context) string {
	return av.desc
}

func (av *attributeValidator) MarkdownDescription(context.Context) string {
	return ""
}

func (av *attributeValidator) Validate(ctx context.Context, req tfsdk.ValidateAttributeRequest, res *tfsdk.ValidateAttributeResponse) {
	av.validateFunc(ctx, req, res)
}

func NewValidator(description string, fn func(context.Context, tfsdk.ValidateAttributeRequest, *tfsdk.ValidateAttributeResponse)) tfsdk.AttributeValidator {
	return &attributeValidator{desc: description, validateFunc: fn}
}

func NewValidatorRegex(description string, rg *regexp.Regexp) tfsdk.AttributeValidator {
	return NewValidator(description, func(ctx context.Context, req tfsdk.ValidateAttributeRequest, res *tfsdk.ValidateAttributeResponse) {

		var str types.String
		diags := tfsdk.ValueAs(ctx, req.AttributeConfig, &str)
		res.Diagnostics.Append(diags...)
		if res.Diagnostics.HasError() {
			return
		}

		if str.Null || str.Unknown {
			return
		}

		if !rg.MatchString(str.Value) {
			res.Diagnostics.AddAttributeError(
				req.AttributePath,
				"invalid organisation ID",
				fmt.Sprintf("organisation do not starts with 'user_' or 'orga_' ('%s')", str.Value),
			)
		}
	})
}
