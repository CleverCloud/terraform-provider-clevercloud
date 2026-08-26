package kubernetes

import (
	"net/http"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

func TestFeatureEnabled(t *testing.T) {
	enabled, disabled := true, false

	tests := []struct {
		name     string
		features *tmp.KubernetesFeatures
		expected bool
	}{
		{
			name:     "no features object",
			features: nil,
			expected: false,
		},
		{
			name:     "features object without the field",
			features: &tmp.KubernetesFeatures{},
			expected: false,
		},
		{
			name:     "node autoprovisioning disabled",
			features: &tmp.KubernetesFeatures{NodeAutoprovisioning: &disabled},
			expected: false,
		},
		{
			name:     "node autoprovisioning enabled",
			features: &tmp.KubernetesFeatures{NodeAutoprovisioning: &enabled},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := featureEnabled(tt.features); got != tt.expected {
				t.Errorf("featureEnabled() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestReadNodeAutoprovisioning pins what Read writes in state: a computed
// attribute left null produces a permanent diff on import and refresh
func TestReadNodeAutoprovisioning(t *testing.T) {
	enabled, disabled := true, false

	tests := []struct {
		name     string
		features *tmp.KubernetesFeatures
		expected bool
	}{
		{name: "no features object", features: nil, expected: false},
		{name: "features object without the field", features: &tmp.KubernetesFeatures{}, expected: false},
		{name: "node autoprovisioning disabled", features: &tmp.KubernetesFeatures{NodeAutoprovisioning: &disabled}, expected: false},
		{name: "node autoprovisioning enabled", features: &tmp.KubernetesFeatures{NodeAutoprovisioning: &enabled}, expected: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// same mapping as Read, on a zero value state
			state := Kubernetes{}
			state.NodeAutoprovisioning = pkg.FromBool(featureEnabled(tt.features))

			if state.NodeAutoprovisioning.IsNull() || state.NodeAutoprovisioning.IsUnknown() {
				t.Fatalf("NodeAutoprovisioning = %v, want a known value", state.NodeAutoprovisioning)
			}
			if state.NodeAutoprovisioning.ValueBool() != tt.expected {
				t.Errorf("NodeAutoprovisioning = %v, want %v", state.NodeAutoprovisioning.ValueBool(), tt.expected)
			}
		})
	}
}

func TestClassifyPatchError(t *testing.T) {
	tests := []struct {
		name          string
		statusCode    int
		enabled       bool
		wantRetryable bool
		wantDetail    string
	}{
		{
			name:          "412 on enable is retryable, features are locked",
			statusCode:    http.StatusPreconditionFailed,
			enabled:       true,
			wantRetryable: true,
		},
		{
			name:          "412 on disable is retryable, features are locked",
			statusCode:    http.StatusPreconditionFailed,
			enabled:       false,
			wantRetryable: true,
		},
		{
			name:       "400 points at the mutually exclusive autoscaling feature",
			statusCode: http.StatusBadRequest,
			enabled:    true,
			wantDetail: "node group autoscaling",
		},
		{
			name:       "409 on enable points at the foreign Karpenter",
			statusCode: http.StatusConflict,
			enabled:    true,
			wantDetail: "kube-system",
		},
		{
			name:       "409 on disable points at the leftover custom resources",
			statusCode: http.StatusConflict,
			enabled:    false,
			wantDetail: "CleverNodeClass",
		},
		{
			name:       "404 is not retryable and has no extra detail",
			statusCode: http.StatusNotFound,
			enabled:    true,
		},
		{
			name:       "500 is not retryable and has no extra detail",
			statusCode: http.StatusInternalServerError,
			enabled:    true,
		},
		{
			name:       "no status means the request never left",
			statusCode: 0,
			enabled:    true,
			wantDetail: "never reached",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			retryable, detail := classifyPatchError(tt.statusCode, tt.enabled)
			if retryable != tt.wantRetryable {
				t.Errorf("classifyPatchError(%d, %v) retryable = %v, want %v", tt.statusCode, tt.enabled, retryable, tt.wantRetryable)
			}
			if !strings.Contains(detail, tt.wantDetail) {
				t.Errorf("classifyPatchError(%d, %v) detail = %q, want it to contain %q", tt.statusCode, tt.enabled, detail, tt.wantDetail)
			}
		})
	}
}

func TestNodeAutoprovisioningChanged(t *testing.T) {
	tests := []struct {
		name     string
		plan     types.Bool
		state    types.Bool
		expected bool
	}{
		{
			// a state written before the attribute existed, the default plans
			// false against it and patching would disable a feature nobody asked
			// about
			name:     "null state against the default",
			plan:     types.BoolValue(false),
			state:    types.BoolNull(),
			expected: false,
		},
		{
			name:     "null state against an enable",
			plan:     types.BoolValue(true),
			state:    types.BoolNull(),
			expected: true,
		},
		{
			name:     "unchanged and disabled",
			plan:     types.BoolValue(false),
			state:    types.BoolValue(false),
			expected: false,
		},
		{
			name:     "unchanged and enabled",
			plan:     types.BoolValue(true),
			state:    types.BoolValue(true),
			expected: false,
		},
		{
			name:     "enable",
			plan:     types.BoolValue(true),
			state:    types.BoolValue(false),
			expected: true,
		},
		{
			name:     "disable",
			plan:     types.BoolValue(false),
			state:    types.BoolValue(true),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := nodeAutoprovisioningChanged(tt.plan, tt.state); got != tt.expected {
				t.Errorf("nodeAutoprovisioningChanged(%v, %v) = %v, want %v", tt.plan, tt.state, got, tt.expected)
			}
		})
	}
}

func TestPatchErrorDetail(t *testing.T) {
	tests := []struct {
		name        string
		explanation string
		apiError    string
		requestID   string
		expected    string
	}{
		{
			name:        "every part is known",
			explanation: "a Karpenter installation already runs",
			apiError:    "invalid response from CleverCloud API (status=409)",
			requestID:   "01H-abc",
			expected:    "a Karpenter installation already runs\ninvalid response from CleverCloud API (status=409)\nrequest id: 01H-abc",
		},
		{
			// unmapped statuses carry no explanation of ours, the API error is
			// the whole detail and must not be preceded by a blank line
			name:      "no explanation",
			apiError:  "invalid response from CleverCloud API (status=500)",
			requestID: "01H-abc",
			expected:  "invalid response from CleverCloud API (status=500)\nrequest id: 01H-abc",
		},
		{
			name:        "the request never left, there is no request id",
			explanation: "the request never reached the Clever Cloud API",
			apiError:    "dial tcp: lookup api.clever-cloud.com: no such host",
			expected:    "the request never reached the Clever Cloud API\ndial tcp: lookup api.clever-cloud.com: no such host",
		},
		{
			name:     "nothing to report",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := patchErrorDetail(tt.explanation, tt.apiError, tt.requestID); got != tt.expected {
				t.Errorf("patchErrorDetail() = %q, want %q", got, tt.expected)
			}
		})
	}
}
