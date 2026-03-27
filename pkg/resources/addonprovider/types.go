package addonprovider

import (
	"context"
	"fmt"
	"net/url"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var _ basetypes.StringTypable = (*URLType)(nil)

// URLType is a custom type that wraps a string and parses it as a URL
type URLType struct {
	basetypes.StringType
}

func (t URLType) Equal(o attr.Type) bool {
	other, ok := o.(URLType)
	if !ok {
		return false
	}
	return t.StringType.Equal(other.StringType)
}

func (t URLType) String() string {
	return "URLType"
}

func (t URLType) ValueFromString(ctx context.Context, in basetypes.StringValue) (basetypes.StringValuable, diag.Diagnostics) {
	var diags diag.Diagnostics

	if in.IsNull() || in.IsUnknown() {
		return URLValue{StringValue: in}, diags
	}

	urlStr := in.ValueString()

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		diags.AddError(
			"Invalid URL",
			fmt.Sprintf("Value %q cannot be parsed as a URL: %s", urlStr, err),
		)
		return URLValue{StringValue: in}, diags
	}

	return URLValue{
		StringValue: in,
		url:         parsedURL,
	}, diags
}

func (t URLType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	attrValue, err := t.StringType.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, err
	}

	stringValue, ok := attrValue.(basetypes.StringValue)
	if !ok {
		return nil, fmt.Errorf("unexpected value type of %T", attrValue)
	}

	stringValuable, diags := t.ValueFromString(ctx, stringValue)
	if diags.HasError() {
		return nil, fmt.Errorf("error converting StringValue to URLValue: %v", diags)
	}

	return stringValuable, nil
}

func (t URLType) ValueType(ctx context.Context) attr.Value {
	return URLValue{}
}

var _ basetypes.StringValuable = (*URLValue)(nil)

// URLValue is a custom value that wraps a string and provides access to the parsed URL
type URLValue struct {
	basetypes.StringValue
	url *url.URL
}

func (v URLValue) Equal(o attr.Value) bool {
	other, ok := o.(URLValue)
	if !ok {
		return false
	}
	return v.StringValue.Equal(other.StringValue)
}

func (v URLValue) Type(ctx context.Context) attr.Type {
	return URLType{}
}

// URL returns the parsed URL, or nil if the value is null or unknown
func (v URLValue) URL() *url.URL {
	return v.url
}

// NewURLValue creates a new URLValue from a string
func NewURLValue(value *url.URL) URLValue {
	return URLValue{
		StringValue: basetypes.NewStringValue(value.String()),
		url:         value,
	}
}

// NewURLNull creates a null URLValue
func NewURLNull() URLValue {
	return URLValue{
		StringValue: basetypes.NewStringNull(),
	}
}

// NewURLUnknown creates an unknown URLValue
func NewURLUnknown() URLValue {
	return URLValue{
		StringValue: basetypes.NewStringUnknown(),
	}
}
