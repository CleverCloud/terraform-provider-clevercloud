package kubernetes

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
	"go.clever-cloud.dev/client"
)

// testClient points the Clever Cloud client at a local server. The client only
// signs a request when it carries an authenticator, so leaving it out is enough
func testClient(t *testing.T, handler http.HandlerFunc) *client.Client {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	return client.New(client.WithEndpoint(server.URL))
}

// blockUntilCancelled keeps a request hanging without wedging the test server:
// httptest.Server.Close waits for its handlers to return, and a handler parked
// on the request context alone can outlive the client that cancelled it
func blockUntilCancelled(r *http.Request) {
	select {
	case <-r.Context().Done():
	case <-time.After(2 * time.Second):
	}
}

func writeCluster(t *testing.T, w http.ResponseWriter, status string, nodeAutoprovisioning *bool) {
	t.Helper()

	view := tmp.ClusterView{ID: "cluster", Name: "tf-test-cluster", Status: status}
	if nodeAutoprovisioning != nil {
		view.Features = &tmp.KubernetesFeatures{NodeAutoprovisioning: nodeAutoprovisioning}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(view); err != nil {
		t.Errorf("failed to encode the cluster: %s", err)
	}
}

func errorSummaries(diags diag.Diagnostics) []string {
	summaries := []string{}
	for _, d := range diags.Errors() {
		summaries = append(summaries, d.Summary()+": "+d.Detail())
	}

	return summaries
}

// The API reports the feature by proof of installation, so the first answers
// still carry the previous value: the wait has to keep polling through them
func TestWaitForNodeAutoprovisioningConverges(t *testing.T) {
	var calls int32
	cc := testClient(t, func(w http.ResponseWriter, _ *http.Request) {
		installed := atomic.AddInt32(&calls, 1) >= 3
		writeCluster(t, w, "RECONCILING", &installed)
	})

	diags := diag.Diagnostics{}
	converged := waitForNodeAutoprovisioning(t.Context(), cc, "orga", "cluster", true, time.Millisecond, 10*time.Second, &diags)

	if !converged {
		t.Fatalf("expected the wait to converge, got diagnostics %v", errorSummaries(diags))
	}
	if diags.HasError() {
		t.Errorf("expected no error diagnostic, got %v", errorSummaries(diags))
	}
	if got := atomic.LoadInt32(&calls); got < 3 {
		t.Errorf("expected the wait to poll until the value changed, got %d calls", got)
	}
}

// A cluster deleted out of band answers 404 forever. Retrying it until the
// deadline would hide the cause behind a generic timeout
func TestWaitForNodeAutoprovisioningStopsOnPermanentError(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		wantDetail string
	}{
		{name: "the cluster is gone", statusCode: http.StatusNotFound, wantDetail: "does not exist any more"},
		{name: "the credentials expired", statusCode: http.StatusUnauthorized, wantDetail: "credentials were refused"},
		{name: "access was lost", statusCode: http.StatusForbidden, wantDetail: "credentials were refused"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var calls int32
			cc := testClient(t, func(w http.ResponseWriter, _ *http.Request) {
				atomic.AddInt32(&calls, 1)
				w.WriteHeader(tt.statusCode)
			})

			diags := diag.Diagnostics{}
			start := time.Now()
			converged := waitForNodeAutoprovisioning(t.Context(), cc, "orga", "cluster", true, time.Millisecond, 10*time.Second, &diags)

			if converged {
				t.Fatal("expected the wait to fail on a permanent error")
			}
			if !diags.HasError() {
				t.Fatal("expected an error diagnostic naming the cause, got none")
			}
			if detail := strings.Join(errorSummaries(diags), "\n"); !strings.Contains(detail, tt.wantDetail) {
				t.Errorf("expected the diagnostic to explain the cause, got %q", detail)
			}
			if got := atomic.LoadInt32(&calls); got != 1 {
				t.Errorf("expected the wait to give up on the first answer, got %d calls", got)
			}
			if elapsed := time.Since(start); elapsed > 3*time.Second {
				t.Errorf("expected the wait to give up immediately, took %s", elapsed)
			}
		})
	}
}

// A transient failure is not a reason to give up: the cluster is still there
func TestWaitForNodeAutoprovisioningRetriesTransientErrors(t *testing.T) {
	var calls int32
	cc := testClient(t, func(w http.ResponseWriter, _ *http.Request) {
		if atomic.AddInt32(&calls, 1) < 3 {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		installed := true
		writeCluster(t, w, "ACTIVE", &installed)
	})

	diags := diag.Diagnostics{}
	if !waitForNodeAutoprovisioning(t.Context(), cc, "orga", "cluster", true, time.Millisecond, 10*time.Second, &diags) {
		t.Fatalf("expected the wait to ride out the transient errors, got %v", errorSummaries(diags))
	}
}

// A deleted cluster reports no features at all, which a disable request must
// not read as proof that Karpenter was successfully removed
func TestWaitForNodeAutoprovisioningRejectsTerminalStatus(t *testing.T) {
	for _, status := range []string{"FAILED", "DELETING", "DELETED"} {
		t.Run(status, func(t *testing.T) {
			cc := testClient(t, func(w http.ResponseWriter, _ *http.Request) {
				writeCluster(t, w, status, nil)
			})

			diags := diag.Diagnostics{}
			// expected == false is exactly what an absent features object reports
			converged := waitForNodeAutoprovisioning(t.Context(), cc, "orga", "cluster", false, time.Millisecond, 10*time.Second, &diags)

			if converged {
				t.Fatalf("expected a %s cluster to fail the wait, not to be read as converged", status)
			}
			if !diags.HasError() {
				t.Errorf("expected an error diagnostic for a %s cluster", status)
			}
		})
	}
}

// Giving up has to be an error: the whole point of the wait is that the value
// saved in state was actually observed
func TestWaitForNodeAutoprovisioningTimesOutWithAnError(t *testing.T) {
	cc := testClient(t, func(w http.ResponseWriter, _ *http.Request) {
		notInstalled := false
		writeCluster(t, w, "RECONCILING", &notInstalled)
	})

	diags := diag.Diagnostics{}
	converged := waitForNodeAutoprovisioning(t.Context(), cc, "orga", "cluster", true, time.Millisecond, 200*time.Millisecond, &diags)

	if converged {
		t.Fatal("expected the wait to report a failure when the value never converged")
	}
	if !diags.HasError() {
		t.Fatalf("expected an error diagnostic, got warnings only: %v", diags.Warnings())
	}
	if detail := strings.Join(errorSummaries(diags), "\n"); !strings.Contains(detail, "never reported the requested value") {
		t.Errorf("expected the diagnostic to say the value was never reported, got %q", detail)
	}
}

// The client is built on http.DefaultClient, which has no timeout: without a
// deadline-bearing context a stalled response would outlive the whole budget
func TestWaitForNodeAutoprovisioningBoundsAStalledRequest(t *testing.T) {
	cc := testClient(t, func(_ http.ResponseWriter, r *http.Request) {
		blockUntilCancelled(r) // never answers until the caller gives up
	})

	diags := diag.Diagnostics{}
	done := make(chan bool, 1)
	start := time.Now()

	go func() {
		done <- waitForNodeAutoprovisioning(t.Context(), cc, "orga", "cluster", true, time.Millisecond, 300*time.Millisecond, &diags)
	}()

	select {
	case converged := <-done:
		if converged {
			t.Fatal("expected the wait to fail, the server never answered")
		}
		if elapsed := time.Since(start); elapsed > 3*time.Second {
			t.Errorf("expected the wait to respect its 300ms budget, took %s", elapsed)
		}
		if !diags.HasError() {
			t.Error("expected an error diagnostic after a stalled request")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("the wait outlived its 300ms budget on a stalled request")
	}
}

func TestWaitForNodeAutoprovisioningReportsAnInterruptedApply(t *testing.T) {
	cc := testClient(t, func(w http.ResponseWriter, _ *http.Request) {
		notInstalled := false
		writeCluster(t, w, "RECONCILING", &notInstalled)
	})

	ctx, cancel := context.WithCancel(t.Context())
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	diags := diag.Diagnostics{}
	if waitForNodeAutoprovisioning(ctx, cc, "orga", "cluster", true, time.Millisecond, 30*time.Second, &diags) {
		t.Fatal("expected an interrupted wait to report a failure")
	}
	if detail := strings.Join(errorSummaries(diags), "\n"); !strings.Contains(detail, "interrupted") {
		t.Errorf("expected the diagnostic to name the interruption, got %q", detail)
	}
}

// Features stay locked while the cluster is not ACTIVE, which the API signals
// with a 412 and resolves on its own
func TestPatchNodeAutoprovisioningRetriesLockedFeatures(t *testing.T) {
	var calls int32
	cc := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected a PATCH, got %s", r.Method)
		}
		if atomic.AddInt32(&calls, 1) < 3 {
			w.WriteHeader(http.StatusPreconditionFailed)
			return
		}
		installed := true
		writeCluster(t, w, "RECONCILING", &installed)
	})

	diags := diag.Diagnostics{}
	if !patchNodeAutoprovisioning(t.Context(), cc, "orga", "cluster", true, time.Millisecond, 10*time.Second, &diags) {
		t.Fatalf("expected the patch to be accepted once the features unlocked, got %v", errorSummaries(diags))
	}
	if got := atomic.LoadInt32(&calls); got < 3 {
		t.Errorf("expected the patch to be retried, got %d calls", got)
	}
}

func TestPatchNodeAutoprovisioningStopsOnAPermanentError(t *testing.T) {
	var calls int32
	cc := testClient(t, func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusConflict)
	})

	diags := diag.Diagnostics{}
	if patchNodeAutoprovisioning(t.Context(), cc, "orga", "cluster", false, time.Millisecond, 10*time.Second, &diags) {
		t.Fatal("expected the patch to fail on a 409")
	}
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Errorf("expected a single attempt on a permanent error, got %d", got)
	}
	if detail := strings.Join(errorSummaries(diags), "\n"); !strings.Contains(detail, "CleverNodeClass") {
		t.Errorf("expected the diagnostic to list what to delete, got %q", detail)
	}
}

func TestPatchNodeAutoprovisioningTimesOut(t *testing.T) {
	cc := testClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusPreconditionFailed)
	})

	diags := diag.Diagnostics{}
	start := time.Now()

	if patchNodeAutoprovisioning(t.Context(), cc, "orga", "cluster", true, time.Millisecond, 200*time.Millisecond, &diags) {
		t.Fatal("expected the patch to give up once its budget expired")
	}
	if elapsed := time.Since(start); elapsed > 5*time.Second {
		t.Errorf("expected the patch to respect its budget, took %s", elapsed)
	}
	if detail := strings.Join(errorSummaries(diags), "\n"); !strings.Contains(detail, "never accepted the change") {
		t.Errorf("expected the diagnostic to say the cluster never accepted the change, got %q", detail)
	}
}

// A patch whose response was never read may or may not have landed. It carries
// an explicit boolean and is idempotent, so sending it again is how we find out
func TestPatchNodeAutoprovisioningRetriesATransportFailure(t *testing.T) {
	var calls int32
	cc := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&calls, 1) < 3 {
			// close the connection without answering, the server did receive it
			hijacker, ok := w.(http.Hijacker)
			if !ok {
				t.Error("expected the test server to support hijacking")
				return
			}
			conn, _, err := hijacker.Hijack()
			if err != nil {
				t.Errorf("failed to hijack the connection: %s", err)
				return
			}
			_ = conn.Close()

			return
		}
		installed := true
		writeCluster(t, w, "RECONCILING", &installed)
	})

	diags := diag.Diagnostics{}
	if !patchNodeAutoprovisioning(t.Context(), cc, "orga", "cluster", true, time.Millisecond, 10*time.Second, &diags) {
		t.Fatalf("expected the idempotent patch to be retried after a transport failure, got %v", errorSummaries(diags))
	}
	if got := atomic.LoadInt32(&calls); got < 3 {
		t.Errorf("expected the patch to be retried, got %d calls", got)
	}
}

// A transient failure that recovered must not be served back as the reason the
// wait later ran out of budget
func TestWaitForNodeAutoprovisioningForgetsRecoveredErrors(t *testing.T) {
	var calls int32
	cc := testClient(t, func(w http.ResponseWriter, _ *http.Request) {
		if atomic.AddInt32(&calls, 1) == 1 {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		notInstalled := false
		writeCluster(t, w, "RECONCILING", &notInstalled)
	})

	diags := diag.Diagnostics{}
	if waitForNodeAutoprovisioning(t.Context(), cc, "orga", "cluster", true, time.Millisecond, 300*time.Millisecond, &diags) {
		t.Fatal("expected the wait to run out of budget")
	}

	detail := strings.Join(errorSummaries(diags), "\n")
	if strings.Contains(detail, "502") {
		t.Errorf("expected the expiry not to blame an error the cluster recovered from, got %q", detail)
	}
	if !strings.Contains(detail, "never reported the requested value") {
		t.Errorf("expected the expiry to name the real cause, got %q", detail)
	}
}

// A DELETING or DELETED cluster cannot be resumed, so it must not be given the
// remedy that belongs to a FAILED one
func TestWaitForNodeAutoprovisioningAdvisesPerTerminalStatus(t *testing.T) {
	cc := testClient(t, func(w http.ResponseWriter, _ *http.Request) {
		writeCluster(t, w, "DELETED", nil)
	})

	diags := diag.Diagnostics{}
	waitForNodeAutoprovisioning(t.Context(), cc, "orga", "cluster", false, time.Millisecond, 10*time.Second, &diags)

	detail := strings.Join(errorSummaries(diags), "\n")
	if strings.Contains(detail, "has to be resumed") {
		t.Errorf("expected a DELETED cluster not to be told to resume, got %q", detail)
	}
	if !strings.Contains(detail, "remove it from state") {
		t.Errorf("expected the recreate or forget remedy, got %q", detail)
	}
}

// The per-request cap has to bound one attempt inside the budget, so a single
// stall cannot consume the whole thing
func TestWaitForNodeAutoprovisioningCapsASingleStalledRequest(t *testing.T) {
	original := nodeAutoprovisioningRequestTimeout
	nodeAutoprovisioningRequestTimeout = 30 * time.Millisecond
	t.Cleanup(func() { nodeAutoprovisioningRequestTimeout = original })

	var calls int32
	cc := testClient(t, func(_ http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		blockUntilCancelled(r)
	})

	diags := diag.Diagnostics{}
	if waitForNodeAutoprovisioning(t.Context(), cc, "orga", "cluster", true, time.Millisecond, 2*time.Second, &diags) {
		t.Fatal("expected the wait to fail, the server never answered")
	}
	if got := atomic.LoadInt32(&calls); got < 2 {
		t.Errorf("expected the per-request cap to free the loop for another attempt, got %d call(s)", got)
	}
}

// A dead endpoint must not hold an apply for the whole budget
func TestPatchNodeAutoprovisioningGivesUpOnADeadEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	endpoint := server.URL
	server.Close() // nothing is listening any more

	cc := client.New(client.WithEndpoint(endpoint))
	diags := diag.Diagnostics{}
	start := time.Now()

	if patchNodeAutoprovisioning(t.Context(), cc, "orga", "cluster", true, time.Millisecond, 30*time.Second, &diags) {
		t.Fatal("expected the patch to fail against a dead endpoint")
	}
	if elapsed := time.Since(start); elapsed > 5*time.Second {
		t.Errorf("expected the patch to give up quickly rather than burn its budget, took %s", elapsed)
	}

	detail := strings.Join(errorSummaries(diags), "\n")
	if strings.Contains(detail, "features stay locked") {
		t.Errorf("expected no claim about locked features, the API was never reached: %q", detail)
	}
	if !strings.Contains(detail, "could not be reached") {
		t.Errorf("expected the diagnostic to say the API could not be reached, got %q", detail)
	}
}

// A 2xx whose body the deadline truncated is not a permanent rejection
func TestPatchNodeAutoprovisioningRetriesATruncatedSuccess(t *testing.T) {
	var calls int32
	cc := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&calls, 1) < 3 {
			w.Header().Set("Content-Length", "4096")
			w.WriteHeader(http.StatusOK)
			if _, err := w.Write([]byte(`{"id":"clu`)); err != nil {
				return
			}
			if flusher, ok := w.(http.Flusher); ok {
				flusher.Flush()
			}
			blockUntilCancelled(r) // never finishes the body

			return
		}
		installed := true
		writeCluster(t, w, "RECONCILING", &installed)
	})

	original := nodeAutoprovisioningRequestTimeout
	nodeAutoprovisioningRequestTimeout = 50 * time.Millisecond
	t.Cleanup(func() { nodeAutoprovisioningRequestTimeout = original })

	diags := diag.Diagnostics{}
	if !patchNodeAutoprovisioning(t.Context(), cc, "orga", "cluster", true, time.Millisecond, 20*time.Second, &diags) {
		t.Fatalf("expected an unreadable 2xx to be retried, not treated as a rejection: %v", errorSummaries(diags))
	}
}

// The request the loop's own deadline cuts short must not overwrite the reason
// the loop was still retrying: at production constants the poll interval
// divides the budget exactly, so that last request is routinely the one killed
func TestPatchNodeAutoprovisioningKeepsTheRealCauseAtExpiry(t *testing.T) {
	var calls int32
	cc := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&calls, 1) <= 2 {
			w.WriteHeader(http.StatusPreconditionFailed)
			return
		}
		blockUntilCancelled(r) // still locked, but this attempt is cut short
	})

	diags := diag.Diagnostics{}
	if patchNodeAutoprovisioning(t.Context(), cc, "orga", "cluster", true, 10*time.Millisecond, 300*time.Millisecond, &diags) {
		t.Fatal("expected the patch to run out of budget")
	}

	detail := strings.Join(errorSummaries(diags), "\n")
	if !strings.Contains(detail, "features stay locked") {
		t.Errorf("expected the 412 to survive as the cause, got %q", detail)
	}
	if strings.Contains(detail, "could not be reached") {
		t.Errorf("expected our own cancellation not to be reported as an unreachable API, got %q", detail)
	}
}
