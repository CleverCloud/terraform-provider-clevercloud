package helper

import (
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// only object is well supported
func SerializeValue(v tftypes.Value) map[string]any {
	if v.IsNull() {
		return nil
	}
	return serialize(v)
}

func serialize(v tftypes.Value) map[string]any {
	if v.IsNull() {
		return nil
	}
	if !v.IsKnown() {
		return map[string]any{}
	}

	x := v.Type()
	switch {
	case x.Is(tftypes.Object{}):
		o := map[string]tftypes.Value{}
		err := v.As(&o)
		if err != nil {
			panic(err)
		}

		m := map[string]any{}
		for k, v := range o {
			m[k] = v.String()
		}
		return m
	default:
		return map[string]any{"-": v.String()}
	}
}
