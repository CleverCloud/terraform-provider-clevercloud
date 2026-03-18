package pkg

import (
	"context"
	"fmt"
	"regexp"
	"slices"
	"strings"

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
		if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
			return
		}

		value := req.ConfigValue.ValueString()

		if !rg.MatchString(value) {
			res.Diagnostics.AddAttributeError(
				req.Path,
				"invalid value for attribute",
				fmt.Sprintf("value do not match validation pattern, got '%s')", value),
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

func NewStringEnumValidator(description string, values ...string) validator.String {
	return &stringValidator{
		description,
		func(ctx context.Context, req validator.StringRequest, res *validator.StringResponse) {
			if !slices.Contains(values, req.ConfigValue.ValueString()) {
				res.Diagnostics.AddAttributeError(
					req.Path,
					"invalid enum value",
					fmt.Sprintf("value must be one of %s", strings.Join(values, ", ")),
				)
			}
		},
	}
}

// NewLowercaseValidator creates a validator that ensures the string value is lowercase
func NewLowercaseValidator() validator.String {
	return &stringValidator{
		"value must be lowercase",
		func(ctx context.Context, req validator.StringRequest, res *validator.StringResponse) {
			if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
				return
			}

			value := req.ConfigValue.ValueString()
			valueLower := strings.ToLower(value)
			if value != valueLower {
				res.Diagnostics.AddAttributeError(
					req.Path,
					"Value must be lowercase",
					fmt.Sprintf("The value '%s' must be lowercase. Use '%s' instead.", value, valueLower),
				)
			}
		},
	}
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

type mapValidator struct {
	description string
	fn          func(context.Context, validator.MapRequest, *validator.MapResponse)
}

func NewMapValidator(description string, fn func(context.Context, validator.MapRequest, *validator.MapResponse)) validator.Map {
	return &mapValidator{description, fn}
}

func (mv *mapValidator) Description(context.Context) string {
	return mv.description
}

func (mv *mapValidator) MarkdownDescription(ctx context.Context) string {
	return mv.Description(ctx)
}

func (mv *mapValidator) ValidateMap(ctx context.Context, req validator.MapRequest, res *validator.MapResponse) {
	mv.fn(ctx, req, res)
}

// NoNullMapValuesValidator creates a validator that rejects maps with null values
func NoNullMapValuesValidator() validator.Map {
	return NewMapValidator(
		"map values cannot be null",
		func(ctx context.Context, req validator.MapRequest, res *validator.MapResponse) {
			if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
				return
			}

			elements := req.ConfigValue.Elements()
			nullKeys := []string{}

			for key, value := range elements {
				if value.IsNull() {
					nullKeys = append(nullKeys, key)
				}
			}

			if len(nullKeys) > 0 {
				res.Diagnostics.AddAttributeError(
					req.Path,
					"Null values are not allowed in environment variables",
					fmt.Sprintf(
						"The following environment variable(s) have null values: %s\n\n"+
							"To conditionally include environment variables, use a for expression:\n"+
							"environment = { for k, v in {\n"+
							"  VAR1 = \"value\"\n"+
							"  VAR2 = var.optional_var\n"+
							"} : k => v if v != null }",
						strings.Join(nullKeys, ", "),
					),
				)
			}
		},
	)
}
