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

		// a UUID is accepted with and without its hyphens
		{Str: "pulsar_590b35ea-63d0-47da-a31b-deaf8f8b8bd7", Match: true},
		{Str: "pulsar_a45dc233c202481c8243fd37257b48bd", Match: true},
		{Str: "metabase_550e8400e29b41d4a716446655440000", Match: true},
		{Str: "pulsar_A45DC233C202481C8243FD37257B48BD", Match: true},
		// but not with some of them, nor with another length
		{Str: "pulsar_a45dc233-c202481c8243fd37257b48bd", Match: false},
		{Str: "pulsar_a45dc233c202481c8243fd37257b48b", Match: false},
		{Str: "pulsar_a45dc233c202481c8243fd37257b48bd0", Match: false},
		{Str: "pulsar_g45dc233c202481c8243fd37257b48bd", Match: false},

		// every add-on exposing its real ID
		{Str: "bucket_00000000-0000-0000-0000-000000000000", Match: true},
		{Str: "cellar_00000000-0000-0000-0000-000000000000", Match: true},
		{Str: "config_00000000-0000-0000-0000-000000000000", Match: true},
		{Str: "elasticsearch_00000000-0000-0000-0000-000000000000", Match: true},
		{Str: "keycloak_00000000-0000-0000-0000-000000000000", Match: true},
		{Str: "matomo_00000000-0000-0000-0000-000000000000", Match: true},
		{Str: "metabase_00000000-0000-0000-0000-000000000000", Match: true},
		{Str: "mongodb_00000000-0000-0000-0000-000000000000", Match: true},
		{Str: "mysql_00000000-0000-0000-0000-000000000000", Match: true},
		{Str: "ng_00000000-0000-0000-0000-000000000000", Match: true},
		{Str: "otoroshi_00000000-0000-0000-0000-000000000000", Match: true},
		{Str: "redis_00000000-0000-0000-0000-000000000000", Match: true},

		// the whole string has to match
		{Str: "postgresql_00000000-0000-0000-0000-000000000000-suffix", Match: false},
		{Str: "prefix-postgresql_00000000-0000-0000-0000-000000000000", Match: false},
		{Str: "kv_01K5C2S48VXPP4VQE7RMD4P609X", Match: false},
		{Str: "prefix-kv_01K5C2S48VXPP4VQE7RMD4P609", Match: false},
		// a ULID only goes with Materia KV, a UUID with the others
		{Str: "pulsar_01K5C2S48VXPP4VQE7RMD4P609", Match: false},
		{Str: "kv_00000000-0000-0000-0000-000000000000", Match: false},
		{Str: "pulsar_", Match: false},
		{Str: "", Match: false},
	}

	for _, test := range tests {
		t.Run(test.Str, func(t *testing.T) {
			if match := ServiceRegExp.MatchString(test.Str); match != test.Match {
				t.Errorf("ServiceRegExp.MatchString(%s) = %v, want %v", test.Str, match, test.Match)
			}
		})
	}
}
