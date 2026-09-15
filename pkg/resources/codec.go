package resources

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/miton18/helper/maps"
)

// FeatureKind picks the value type a Codec entry encodes/decodes.
type FeatureKind int

const (
	KindBool FeatureKind = iota
	KindString
	KindInt64
	// KindEnvFlag: true → api[key]=TruthyValue, false → absent. Binary switch only; for an enumeration use KindString.
	KindEnvFlag
	// KindBoolString: bool ⇄ explicit "true"/"false", both states written, decoded via strconv.ParseBool.
	KindBoolString
)

// FeatureMapping pairs a state struct field with an API key.
type FeatureMapping struct {
	StateField  string
	APIKeyName  string
	Kind        FeatureKind
	TruthyValue string
}

// Codec is the bidirectional mapping between schema attributes and API keys for one resource.
type Codec []FeatureMapping

// stateValueRead accepts a struct value or non-nil pointer for read-only access.
func stateValueRead(state any) (reflect.Value, diag.Diagnostics) {
	var diags diag.Diagnostics
	v := reflect.ValueOf(state)
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			diags.AddError(
				"Codec internal error",
				fmt.Sprintf("state must be a non-nil pointer, got %T", state),
			)
			return reflect.Value{}, diags
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		diags.AddError(
			"Codec internal error",
			fmt.Sprintf("state must be a struct or pointer to struct, got %s", v.Kind()),
		)
		return reflect.Value{}, diags
	}
	return v, diags
}

// stateValueWrite requires a non-nil pointer to a struct for mutation.
func stateValueWrite(state any) (reflect.Value, diag.Diagnostics) {
	var diags diag.Diagnostics
	v := reflect.ValueOf(state)
	if v.Kind() != reflect.Pointer || v.IsNil() {
		diags.AddError(
			"Codec internal error",
			fmt.Sprintf("state must be a non-nil pointer to a struct, got %T", state),
		)
		return reflect.Value{}, diags
	}
	elem := v.Elem()
	if elem.Kind() != reflect.Struct {
		diags.AddError(
			"Codec internal error",
			fmt.Sprintf("state must point to a struct, got pointer to %s", elem.Kind()),
		)
		return reflect.Value{}, diags
	}
	return elem, diags
}

// fieldAs type-asserts a state struct field to the expected terraform value type T,
// recording a uniform "Codec internal error" diagnostic on mismatch.
func fieldAs[T any](field reflect.Value, m FeatureMapping, want string) (T, diag.Diagnostics) {
	var diags diag.Diagnostics
	v, ok := field.Interface().(T)
	if !ok {
		diags.AddError(
			"Codec internal error",
			fmt.Sprintf("field %q has type %s, expected %s", m.StateField, field.Type(), want),
		)
	}
	return v, diags
}

// rawAs type-asserts a raw API value to the expected Go type T,
// recording a uniform "Codec API type mismatch" diagnostic on mismatch.
func rawAs[T any](raw any, m FeatureMapping, want string) (T, diag.Diagnostics) {
	var diags diag.Diagnostics
	v, ok := raw.(T)
	if !ok {
		diags.AddError(
			"Codec API type mismatch",
			fmt.Sprintf("api[%q] is %T, expected %s", m.APIKeyName, raw, want),
		)
	}
	return v, diags
}

// Validate checks every mapping resolves to an existing struct field of the
// Kind's expected type. Use in a per-resource test to catch wiring mistakes at CI time.
func (c Codec) Validate(state any) diag.Diagnostics {
	sv, diags := stateValueRead(state)
	if diags.HasError() {
		return diags
	}
	for _, m := range c {
		field := sv.FieldByName(m.StateField)
		if !field.IsValid() {
			diags.AddError("Codec contract error", fmt.Sprintf("no field %q (API key %q)", m.StateField, m.APIKeyName))
			continue
		}
		var want string
		switch m.Kind {
		case KindString:
			if _, ok := field.Interface().(types.String); !ok {
				want = "types.String"
			}
		case KindInt64:
			if _, ok := field.Interface().(types.Int64); !ok {
				want = "types.Int64"
			}
		default: // KindBool, KindEnvFlag, KindBoolString
			if _, ok := field.Interface().(types.Bool); !ok {
				want = "types.Bool"
			}
		}
		if want != "" {
			diags.AddError("Codec contract error", fmt.Sprintf("field %q is %s, expected %s", m.StateField, field.Type(), want))
		}
	}
	return diags
}

// StateToAPI encodes state struct fields into api; null/unknown fields are skipped.
func (c Codec) StateToAPI(state any, api map[string]any) diag.Diagnostics {
	sv, diags := stateValueRead(state)
	if diags.HasError() {
		return diags
	}

	for _, m := range c {
		field := sv.FieldByName(m.StateField)
		if !field.IsValid() {
			diags.AddError(
				"Codec internal error",
				fmt.Sprintf("state struct has no field %q (mapping for API key %q)", m.StateField, m.APIKeyName),
			)
			continue
		}

		switch m.Kind {
		case KindBool:
			b, d := fieldAs[types.Bool](field, m, "types.Bool")
			diags.Append(d...)
			if d.HasError() || b.IsNull() || b.IsUnknown() {
				continue
			}
			api[m.APIKeyName] = b.ValueBool()

		case KindString:
			s, d := fieldAs[types.String](field, m, "types.String")
			diags.Append(d...)
			if d.HasError() || s.IsNull() || s.IsUnknown() {
				continue
			}
			api[m.APIKeyName] = s.ValueString()

		case KindInt64:
			n, d := fieldAs[types.Int64](field, m, "types.Int64")
			diags.Append(d...)
			if d.HasError() || n.IsNull() || n.IsUnknown() {
				continue
			}
			api[m.APIKeyName] = n.ValueInt64()

		case KindEnvFlag:
			b, d := fieldAs[types.Bool](field, m, "types.Bool")
			diags.Append(d...)
			if d.HasError() || b.IsNull() || b.IsUnknown() || !b.ValueBool() {
				continue
			}
			api[m.APIKeyName] = m.TruthyValue

		case KindBoolString:
			b, d := fieldAs[types.Bool](field, m, "types.Bool")
			diags.Append(d...)
			if d.HasError() || b.IsNull() || b.IsUnknown() {
				continue
			}
			api[m.APIKeyName] = strconv.FormatBool(b.ValueBool())
		}
	}
	return diags
}

// APIToState decodes api into state struct fields. Absent-key handling mirrors
// the pkg helper each kind replaced: KindBool and KindEnvFlag preserve the
// existing value (sparse feature lists / SetBoolIf), while KindString, KindInt64
// and KindBoolString yield typed null (FromStrPtr/FromIntPtr/FromBoolPtr).
func (c Codec) APIToState(api map[string]any, state any) diag.Diagnostics {
	sv, diags := stateValueWrite(state)
	if diags.HasError() {
		return diags
	}

	for _, m := range c {
		field := sv.FieldByName(m.StateField)
		if !field.IsValid() {
			diags.AddError(
				"Codec internal error",
				fmt.Sprintf("state struct has no field %q (mapping for API key %q)", m.StateField, m.APIKeyName),
			)
			continue
		}
		if !field.CanSet() {
			diags.AddError(
				"Codec internal error",
				fmt.Sprintf("state field %q is not settable (unexported?)", m.StateField),
			)
			continue
		}

		raw, present := api[m.APIKeyName]

		switch m.Kind {
		case KindBool:
			// absent → preserve: the API reports a sparse feature list, so a
			// missing key means "unchanged", not "false" (mirrors the old
			// override-only loops in the database resources).
			if !present {
				continue
			}
			b, d := rawAs[bool](raw, m, "bool")
			diags.Append(d...)
			if d.HasError() {
				continue
			}
			field.Set(reflect.ValueOf(types.BoolValue(b)))

		case KindString:
			if !present {
				field.Set(reflect.ValueOf(types.StringNull()))
				continue
			}
			s, d := rawAs[string](raw, m, "string")
			diags.Append(d...)
			if d.HasError() {
				continue
			}
			field.Set(reflect.ValueOf(types.StringValue(s)))

		case KindInt64:
			if !present {
				field.Set(reflect.ValueOf(types.Int64Null()))
				continue
			}
			n, d := rawAs[int64](raw, m, "int64")
			diags.Append(d...)
			if d.HasError() {
				continue
			}
			field.Set(reflect.ValueOf(types.Int64Value(n)))

		case KindEnvFlag:
			// like pkg.SetBoolIf: set true only on a truthy match, otherwise preserve the plan/prior value
			if !present {
				continue
			}
			s, d := rawAs[string](raw, m, "string")
			diags.Append(d...)
			if d.HasError() {
				continue
			}
			if s == m.TruthyValue {
				field.Set(reflect.ValueOf(types.BoolValue(true)))
			}

		case KindBoolString:
			if !present {
				field.Set(reflect.ValueOf(types.BoolNull()))
				continue
			}
			s, d := rawAs[string](raw, m, "string")
			diags.Append(d...)
			if d.HasError() {
				continue
			}
			// unparseable yields null, like pkg.FromBoolPtr
			if b, err := strconv.ParseBool(s); err == nil {
				field.Set(reflect.ValueOf(types.BoolValue(b)))
			} else {
				field.Set(reflect.ValueOf(types.BoolNull()))
			}
		}
	}
	return diags
}

// WriteEnv encodes the codec's mappings into a string env map (helper for ToEnv).
func (c Codec) WriteEnv(state any, env map[string]string) diag.Diagnostics {
	api := map[string]any{}
	diags := c.StateToAPI(state, api)
	for k, v := range api {
		switch val := v.(type) {
		case string:
			env[k] = val
		case bool:
			env[k] = strconv.FormatBool(val)
		case int64:
			env[k] = strconv.FormatInt(val, 10)
		}
	}
	return diags
}

// ReadEnv pops each mapped key out of the env map and decodes it into state (helper for FromEnv).
func (c Codec) ReadEnv(env *maps.Map[string, string], state any) diag.Diagnostics {
	api := map[string]any{}
	for _, m := range c {
		if v := env.PopPtr(m.APIKeyName); v != nil {
			api[m.APIKeyName] = *v
		}
	}
	return c.APIToState(api, state)
}
