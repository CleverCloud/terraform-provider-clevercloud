package pkg

import "testing"

func TestRegex(t *testing.T) {
	t.Parallel()

	type RTest struct {
		Str   string
		Match bool
	}

	tests := []RTest{
		{Str: "postgresql_00000000-0000-0000-0000-000000000000", Match: true},
		{Str: "kv_01K5C2S48VXPP4VQE7RMD4P609", Match: true},
		{Str: "moi_f3ff1c24-90bb-402e-b633-0b98c9df522d", Match: false},
	}

	for _, test := range tests {
		t.Run(test.Str, func(t *testing.T) {
			if match := ServiceRegExp.MatchString(test.Str); match != test.Match {
				t.Errorf("ServiceRegExp.MatchString(%s) = %v, want %v", test.Str, match, test.Match)
			}
		})
	}
}
