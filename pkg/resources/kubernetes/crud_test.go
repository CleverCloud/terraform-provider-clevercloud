package kubernetes

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

func TestAutoprovisioningFeatureEnabled(t *testing.T) {
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
			if got := autoprovisioningFeatureEnabled(tt.features); got != tt.expected {
				t.Errorf("autoprovisioningFeatureEnabled() = %v, want %v", got, tt.expected)
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
			state.NodeAutoprovisioning = pkg.FromBool(autoprovisioningFeatureEnabled(tt.features))

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
			name:       "404 says the cluster is gone",
			statusCode: http.StatusNotFound,
			enabled:    true,
			wantDetail: "does not exist any more",
		},
		{
			name:       "401 says the credentials were refused",
			statusCode: http.StatusUnauthorized,
			enabled:    true,
			wantDetail: "credentials were refused",
		},
		{
			name:          "a 2xx we could not read is retryable, the body failed and not the patch",
			statusCode:    http.StatusOK,
			enabled:       true,
			wantRetryable: true,
		},
		{
			name:       "500 is not retryable and has no extra detail",
			statusCode: http.StatusInternalServerError,
			enabled:    true,
		},
		{
			name:          "no status is retryable, the idempotent patch may or may not have landed",
			statusCode:    0,
			enabled:       true,
			wantRetryable: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			retryable, detail := classifyPatchError(tt.statusCode, tt.enabled)
			if retryable != tt.wantRetryable {
				t.Errorf("classifyPatchError(%d, %v) retryable = %v, want %v", tt.statusCode, tt.enabled, retryable, tt.wantRetryable)
			}
			// asserted with Contains only when a detail is expected: an empty
			// wantDetail has to mean "no detail at all", and Contains(detail, "")
			// is true for every string
			if tt.wantDetail == "" {
				if detail != "" {
					t.Errorf("classifyPatchError(%d, %v) detail = %q, want no detail", tt.statusCode, tt.enabled, detail)
				}
			} else if !strings.Contains(detail, tt.wantDetail) {
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

func TestNodeAutoprovisioningValue(t *testing.T) {
	tests := []struct {
		name         string
		plan         types.Bool
		config       types.Bool
		want         types.Bool
		wantResolved bool
	}{
		{
			name:         "a known plan is used as is",
			plan:         types.BoolValue(true),
			config:       types.BoolValue(true),
			want:         types.BoolValue(true),
			wantResolved: true,
		},
		{
			name:         "a plan defaulted to false is used as is",
			plan:         types.BoolValue(false),
			config:       types.BoolNull(),
			want:         types.BoolValue(false),
			wantResolved: true,
		},
		{
			name:         "a null plan is left alone",
			plan:         types.BoolNull(),
			config:       types.BoolNull(),
			want:         types.BoolNull(),
			wantResolved: true,
		},
		{
			name:         "an unknown plan takes the configured true",
			plan:         types.BoolUnknown(),
			config:       types.BoolValue(true),
			want:         types.BoolValue(true),
			wantResolved: true,
		},
		{
			name:         "an unknown plan takes the configured false",
			plan:         types.BoolUnknown(),
			config:       types.BoolValue(false),
			want:         types.BoolValue(false),
			wantResolved: true,
		},
		{
			name:         "an unknown plan with an absent attribute takes the schema default",
			plan:         types.BoolUnknown(),
			config:       types.BoolNull(),
			want:         types.BoolValue(false),
			wantResolved: true,
		},
		{
			name:         "an unknown plan and an unknown configuration cannot be resolved",
			plan:         types.BoolUnknown(),
			config:       types.BoolUnknown(),
			want:         types.BoolUnknown(),
			wantResolved: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, resolved := nodeAutoprovisioningValue(tt.plan, tt.config)
			if resolved != tt.wantResolved {
				t.Errorf("nodeAutoprovisioningValue(%v, %v) resolved = %v, want %v", tt.plan, tt.config, resolved, tt.wantResolved)
			}
			if !got.Equal(tt.want) {
				t.Errorf("nodeAutoprovisioningValue(%v, %v) = %v, want %v", tt.plan, tt.config, got, tt.want)
			}
		})
	}
}

func TestPollErrorFatal(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		wantFatal  bool
		wantDetail string
	}{
		{name: "404 means the cluster is gone", statusCode: http.StatusNotFound, wantFatal: true, wantDetail: "does not exist any more"},
		{name: "401 means the credentials stopped working", statusCode: http.StatusUnauthorized, wantFatal: true, wantDetail: "credentials were refused"},
		{name: "403 means the credentials stopped working", statusCode: http.StatusForbidden, wantFatal: true, wantDetail: "credentials were refused"},
		{name: "500 is transient", statusCode: http.StatusInternalServerError},
		{name: "502 is transient", statusCode: http.StatusBadGateway},
		{name: "429 is transient", statusCode: http.StatusTooManyRequests},
		{name: "a transport failure is transient", statusCode: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if fatal := pollErrorFatal(tt.statusCode); fatal != tt.wantFatal {
				t.Errorf("pollErrorFatal(%d) = %v, want %v", tt.statusCode, fatal, tt.wantFatal)
			}
			detail := pollErrorDetail(tt.statusCode)
			if tt.wantDetail == "" {
				if detail != "" {
					t.Errorf("pollErrorDetail(%d) = %q, want no detail", tt.statusCode, detail)
				}
			} else if !strings.Contains(detail, tt.wantDetail) {
				t.Errorf("pollErrorDetail(%d) = %q, want it to contain %q", tt.statusCode, detail, tt.wantDetail)
			}
		})
	}
}

func TestClusterTerminalStatus(t *testing.T) {
	tests := []struct {
		status string
		want   bool
	}{
		{"FAILED", true},
		{"DELETING", true},
		{"DELETED", true},
		{"ACTIVE", false},
		{"DEPLOYING", false},
		{"RECONCILING", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			if got := clusterTerminalStatus(tt.status); got != tt.want {
				t.Errorf("clusterTerminalStatus(%q) = %v, want %v", tt.status, got, tt.want)
			}
		})
	}
}

// TestUnknownPlanWouldUninstallWithoutResolution pins the hazard the acceptance
// step in kubernetes_test.go exercises against the real API: an unknown planned
// value unwraps to false, which against a cluster that has the feature on reads
// as a request to uninstall Karpenter. Resolving it first has to turn that same
// input into a no-op
func TestUnknownPlanWouldUninstallWithoutResolution(t *testing.T) {
	unknown := types.BoolUnknown()
	enabledState := types.BoolValue(true)

	if unknown.ValueBool() {
		t.Fatal("expected an unknown types.Bool to unwrap to false, the whole hazard rests on it")
	}
	if !nodeAutoprovisioningChanged(unknown, enabledState) {
		t.Fatal("expected the raw unknown to look like a change away from enabled, the hazard is gone and this test is stale")
	}

	resolved, ok := nodeAutoprovisioningValue(unknown, types.BoolValue(true))
	if !ok {
		t.Fatal("expected a resolved configuration value to be usable")
	}
	if nodeAutoprovisioningChanged(resolved, enabledState) {
		t.Error("expected the resolved value to leave an already enabled cluster alone, not to patch Karpenter off")
	}
}

func TestPatchExpiryDetail(t *testing.T) {
	tests := []struct {
		name       string
		lastStatus int
		want       string
		notWant    string
	}{
		{
			name:       "only a 412 licenses the locked features claim",
			lastStatus: http.StatusPreconditionFailed,
			want:       "features stay locked",
		},
		{
			name:       "no status blames the API being unreachable, not the cluster",
			lastStatus: 0,
			want:       "could not be reached",
			notWant:    "features stay locked",
		},
		{
			name:       "anything else says only what is known",
			lastStatus: http.StatusOK,
			want:       "never accepted the change",
			notWant:    "features stay locked",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := patchExpiryDetail(tt.lastStatus, 10*time.Minute)
			if !strings.Contains(got, tt.want) {
				t.Errorf("patchExpiryDetail(%d) = %q, want it to contain %q", tt.lastStatus, got, tt.want)
			}
			if tt.notWant != "" && strings.Contains(got, tt.notWant) {
				t.Errorf("patchExpiryDetail(%d) = %q, want it not to claim %q", tt.lastStatus, got, tt.notWant)
			}
		})
	}
}

func TestTerminalStatusDetail(t *testing.T) {
	if got := terminalStatusDetail("FAILED"); !strings.Contains(got, "has to be resumed") {
		t.Errorf("terminalStatusDetail(FAILED) = %q, want the resume remedy", got)
	}

	for _, status := range []string{"DELETING", "DELETED"} {
		got := terminalStatusDetail(status)
		if strings.Contains(got, "has to be resumed") {
			t.Errorf("terminalStatusDetail(%s) = %q, a cluster on its way out cannot be resumed", status, got)
		}
		if !strings.Contains(got, "remove it from state") {
			t.Errorf("terminalStatusDetail(%s) = %q, want the recreate or forget remedy", status, got)
		}
	}
}
