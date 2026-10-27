package kubernetes

import (
	"encoding/json"
	"testing"

	"go.clever-cloud.dev/sdk/models"
)

// The payloads the cluster resource sends and the views it reads, as the SDK
// encodes them: the features patch is a merge patch, so an omitted field and an
// explicit false must stay distinguishable.

func TestClusterPatchPayload_Marshal(t *testing.T) {
	enabled, disabled := true, false

	tests := []struct {
		name     string
		req      models.ClusterPatchPayload
		expected string
	}{
		{
			name:     "no feature set",
			req:      models.ClusterPatchPayload{},
			expected: `{}`,
		},
		{
			name:     "empty features object",
			req:      models.ClusterPatchPayload{Features: &models.ClusterFeaturesPatch{}},
			expected: `{"features":{}}`,
		},
		{
			name:     "enable node autoprovisioning",
			req:      models.ClusterPatchPayload{Features: &models.ClusterFeaturesPatch{NodeAutoprovisioning: &enabled}},
			expected: `{"features":{"nodeAutoprovisioning":true}}`,
		},
		{
			// a non pointer bool would be dropped by omitempty here and would
			// silently turn every disable request into a no-op
			name:     "disable node autoprovisioning",
			req:      models.ClusterPatchPayload{Features: &models.ClusterFeaturesPatch{NodeAutoprovisioning: &disabled}},
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

func TestClusterCreationPayload_Marshal(t *testing.T) {
	enabled := true

	tests := []struct {
		name     string
		req      models.ClusterCreationPayload
		expected string
	}{
		{
			// the payload production receives today, adding the features field
			// must not change it
			name:     "without features",
			req:      models.ClusterCreationPayload{Name: "tf-test-kubernetes"},
			expected: `{"name":"tf-test-kubernetes"}`,
		},
		{
			name: "with node autoprovisioning",
			req: models.ClusterCreationPayload{
				Name:     "tf-test-kubernetes",
				Features: &models.ClusterFeatures{NodeAutoprovisioning: &enabled},
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

func TestKubernetesCluster_UnmarshalFeatures(t *testing.T) {
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
			cluster := models.KubernetesCluster{}
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
