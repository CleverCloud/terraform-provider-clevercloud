package tmp

import (
	"encoding/json"
	"testing"
)

func TestKubernetesPatchRequest_Marshal(t *testing.T) {
	enabled, disabled := true, false

	tests := []struct {
		name     string
		req      KubernetesPatchRequest
		expected string
	}{
		{
			name:     "no feature set",
			req:      KubernetesPatchRequest{},
			expected: `{}`,
		},
		{
			name:     "empty features object",
			req:      KubernetesPatchRequest{Features: &KubernetesFeatures{}},
			expected: `{"features":{}}`,
		},
		{
			name:     "enable node autoprovisioning",
			req:      KubernetesPatchRequest{Features: &KubernetesFeatures{NodeAutoprovisioning: &enabled}},
			expected: `{"features":{"nodeAutoprovisioning":true}}`,
		},
		{
			// a non pointer bool would be dropped by omitempty here and would
			// silently turn every disable request into a no-op
			name:     "disable node autoprovisioning",
			req:      KubernetesPatchRequest{Features: &KubernetesFeatures{NodeAutoprovisioning: &disabled}},
			expected: `{"features":{"nodeAutoprovisioning":false}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload, err := json.Marshal(tt.req)
			if err != nil {
				t.Fatalf("Marshal() error = %v", err)
			}
			if string(payload) != tt.expected {
				t.Errorf("Marshal() = %s, want %s", payload, tt.expected)
			}
		})
	}
}

func TestKubernetesCreateRequest_Marshal(t *testing.T) {
	enabled := true

	tests := []struct {
		name     string
		req      KubernetesCreateRequest
		expected string
	}{
		{
			// the payload production receives today, adding the features field
			// must not change it
			name:     "without features",
			req:      KubernetesCreateRequest{Name: "tf-test-kubernetes"},
			expected: `{"name":"tf-test-kubernetes"}`,
		},
		{
			name: "with node autoprovisioning",
			req: KubernetesCreateRequest{
				Name:     "tf-test-kubernetes",
				Features: &KubernetesFeatures{NodeAutoprovisioning: &enabled},
			},
			expected: `{"features":{"nodeAutoprovisioning":true},"name":"tf-test-kubernetes"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload, err := json.Marshal(tt.req)
			if err != nil {
				t.Fatalf("Marshal() error = %v", err)
			}
			if string(payload) != tt.expected {
				t.Errorf("Marshal() = %s, want %s", payload, tt.expected)
			}
		})
	}
}

func TestClusterView_UnmarshalFeatures(t *testing.T) {
	tests := []struct {
		name        string
		payload     string
		hasFeatures bool
		hasValue    bool
		value       bool
	}{
		{
			name:    "no features object",
			payload: `{"id":"kubernetes_xxx"}`,
		},
		{
			name:        "empty features object",
			payload:     `{"id":"kubernetes_xxx","features":{}}`,
			hasFeatures: true,
		},
		{
			name:        "node autoprovisioning disabled",
			payload:     `{"id":"kubernetes_xxx","features":{"nodeAutoprovisioning":false}}`,
			hasFeatures: true,
			hasValue:    true,
			value:       false,
		},
		{
			name:        "node autoprovisioning enabled",
			payload:     `{"id":"kubernetes_xxx","features":{"nodeAutoprovisioning":true}}`,
			hasFeatures: true,
			hasValue:    true,
			value:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cluster := ClusterView{}
			if err := json.Unmarshal([]byte(tt.payload), &cluster); err != nil {
				t.Fatalf("Unmarshal() error = %v", err)
			}

			if (cluster.Features != nil) != tt.hasFeatures {
				t.Fatalf("Features != nil = %v, want %v", cluster.Features != nil, tt.hasFeatures)
			}
			if !tt.hasFeatures {
				return
			}

			if (cluster.Features.NodeAutoprovisioning != nil) != tt.hasValue {
				t.Fatalf("NodeAutoprovisioning != nil = %v, want %v", cluster.Features.NodeAutoprovisioning != nil, tt.hasValue)
			}
			if tt.hasValue && *cluster.Features.NodeAutoprovisioning != tt.value {
				t.Errorf("NodeAutoprovisioning = %v, want %v", *cluster.Features.NodeAutoprovisioning, tt.value)
			}
		})
	}
}
