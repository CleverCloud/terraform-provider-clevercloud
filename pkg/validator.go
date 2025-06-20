package pkg

import (
	"context"
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

type attributeValidator struct {
	desc         string
	validateFunc func(context.Context, validator.StringRequest, *validator.StringResponse)
}

func (av *attributeValidator) Description(context.Context) string {
	return av.desc
}

func (av *attributeValidator) MarkdownDescription(context.Context) string {
	return ""
}

func (av *attributeValidator) ValidateString(ctx context.Context, req validator.StringRequest, res *validator.StringResponse) {
	av.validateFunc(ctx, req, res)
}

func NewValidator(description string, fn func(ctx context.Context, req validator.StringRequest, res *validator.StringResponse)) validator.String {
	return &attributeValidator{desc: description, validateFunc: fn}
}

func NewValidatorRegex(description string, rg *regexp.Regexp) validator.String {
	return NewValidator(description, func(ctx context.Context, req validator.StringRequest, res *validator.StringResponse) {

		/*var str types.String

		diags := tfsdk.ValueAs(ctx, req.AttributeConfig, &str)
		res.Diagnostics.Append(diags...)
		if res.Diagnostics.HasError() {
			return
		}*/
		value := req.ConfigValue.ValueString()

		if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
			return
		}

		if !rg.MatchString(value) {
			res.Diagnostics.AddAttributeError(
				req.Path,
				"invalid organisation ID",
				fmt.Sprintf("organisation do not starts with 'user_' or 'orga_' ('%s')", value),
			)
		}
	})
}

type SetValidator struct {
	description string
	fn          func(context.Context, validator.SetRequest, *validator.SetResponse)
}

func NewSetValidator(description string, fn func(context.Context, validator.SetRequest, *validator.SetResponse)) validator.Set {
	return &SetValidator{description, fn}
}

func (sv *SetValidator) Description(context.Context) string {
	return sv.description
}

func (sv *SetValidator) MarkdownDescription(ctx context.Context) string {
	return sv.Description(ctx)
}

func (sv *SetValidator) ValidateSet(ctx context.Context, req validator.SetRequest, res *validator.SetResponse) {
	sv.fn(ctx, req, res)
}

type stringValidator struct {
	description string
	fn          func(context.Context, validator.StringRequest, *validator.StringResponse)
}

func NewStringValidator(description string, fn func(context.Context, validator.StringRequest, *validator.StringResponse)) validator.String {
	return &stringValidator{description, fn}
}

func (sv *stringValidator) Description(context.Context) string {
	return sv.description
}

func (sv *stringValidator) MarkdownDescription(ctx context.Context) string {
	return sv.Description(ctx)
}

func (sv *stringValidator) ValidateString(ctx context.Context, req validator.StringRequest, res *validator.StringResponse) {
	sv.fn(ctx, req, res)
}
