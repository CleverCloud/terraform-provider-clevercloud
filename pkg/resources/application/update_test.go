package application

import "testing"

func TestEnvKeysDiffer(t *testing.T) {
	keys := []string{"CC_RUST_FEATURES"}

	tests := []struct {
		name     string
		planEnv  map[string]string
		stateEnv map[string]string
		want     bool
	}{
		{
			name:     "unchanged",
			planEnv:  map[string]string{"CC_RUST_FEATURES": "tls", "OTHER": "a"},
			stateEnv: map[string]string{"CC_RUST_FEATURES": "tls", "OTHER": "b"},
			want:     false,
		},
		{
			name:     "absent on both sides",
			planEnv:  map[string]string{"OTHER": "a"},
			stateEnv: map[string]string{},
			want:     false,
		},
		{
			name:     "changed",
			planEnv:  map[string]string{"CC_RUST_FEATURES": "tls,json"},
			stateEnv: map[string]string{"CC_RUST_FEATURES": "tls"},
			want:     true,
		},
		{
			name:     "added",
			planEnv:  map[string]string{"CC_RUST_FEATURES": "tls"},
			stateEnv: map[string]string{},
			want:     true,
		},
		{
			name:     "removed",
			planEnv:  map[string]string{},
			stateEnv: map[string]string{"CC_RUST_FEATURES": "tls"},
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := envKeysDiffer(tt.planEnv, tt.stateEnv, keys); got != tt.want {
				t.Errorf("envKeysDiffer() = %v, want %v", got, tt.want)
			}
		})
	}
}
