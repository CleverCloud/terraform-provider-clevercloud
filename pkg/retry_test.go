package pkg

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"go.clever-cloud.dev/client"
)

// recordSleeps replaces the waits of the policy with a record of them.
func recordSleeps(t *testing.T, proceed bool) *[]time.Duration {
	t.Helper()

	var waits []time.Duration
	original := sleep
	sleep = func(_ context.Context, d time.Duration) bool {
		waits = append(waits, d)
		return proceed
	}
	t.Cleanup(func() { sleep = original })

	return &waits
}

// rateLimitedAPI answers 429 to its first refusals requests, then 200.
func rateLimitedAPI(t *testing.T, refusals int32, header http.Header, body string) (*client.Client, *atomic.Int32) {
	t.Helper()

	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sent, _ := io.ReadAll(r.Body)
		if r.Method == http.MethodPost && string(sent) != `{"name":"db"}` {
			t.Errorf("request %d carries the body %q", calls.Load()+1, sent)
		}

		if calls.Add(1) <= refusals {
			for name, values := range header {
				w.Header()[name] = values
			}
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(body))
			return
		}
		_, _ = w.Write([]byte(`{}`))
	}))
	t.Cleanup(server.Close)

	return client.New(client.WithEndpoint(server.URL), client.WithRetryPolicy(RetryPolicy)), &calls
}

func TestRateLimitWait(t *testing.T) {
	now := time.Date(2026, 11, 17, 10, 0, 0, 0, time.UTC)

	for name, tc := range map[string]struct {
		header http.Header
		want   time.Duration
		told   bool
	}{
		"Retry-After in seconds":                               {http.Header{"Retry-After": {"60"}}, time.Minute, true},
		"Retry-After as a date":                                {http.Header{"Retry-After": {"Tue, 17 Nov 2026 10:00:30 GMT"}}, 30 * time.Second, true},
		"Retry-After in the past":                              {http.Header{"Retry-After": {"Tue, 17 Nov 2026 09:00:00 GMT"}}, 0, true},
		"RateLimit-Reset alone":                                {http.Header{"Ratelimit-Reset": {"45"}}, 45 * time.Second, true},
		"Retry-After comes first":                              {http.Header{"Retry-After": {"10"}, "Ratelimit-Reset": {"45"}}, 10 * time.Second, true},
		"unreadable Retry-After falls back on RateLimit-Reset": {http.Header{"Retry-After": {"soon"}, "Ratelimit-Reset": {"45"}}, 45 * time.Second, true},
		"negative values are ignored":                          {http.Header{"Retry-After": {"-5"}, "Ratelimit-Reset": {"-1"}}, 0, false},
		"no header":                                            {http.Header{}, 0, false},
		"nil header":                                           {nil, 0, false},
	} {
		t.Run(name, func(t *testing.T) {
			got, told := rateLimitWait(tc.header, now)
			if got != tc.want || told != tc.told {
				t.Errorf("rateLimitWait = %s, %t, want %s, %t", got, told, tc.want, tc.told)
			}
		})
	}
}

// A rate-limited request is sent again after the wait the API asks for,
// whatever the body of the 429 looks like: only its status and headers count.
func TestRetryPolicyWaitsForARateLimitedRequest(t *testing.T) {
	for name, body := range map[string]string{
		"id and message body": `{"id":11000,"message":"Rate limit exceeded: too many requests, retry later","type":"error"}`,
		"code and error body": `{"apiRequestId":"req_1","code":"clever.core.too-many-requests","context":{"type":"Empty"},"error":"Rate limit exceeded"}`,
		"unknown body":        `{"some":"other shape"}`,
		"no body":             ``,
	} {
		t.Run(name, func(t *testing.T) {
			waits := recordSleeps(t, true)
			cc, calls := rateLimitedAPI(t, 2, http.Header{"Retry-After": {"7"}}, body)

			res := client.Post[map[string]any](t.Context(), cc, "/v2/organisations/orga_1/addons", map[string]string{"name": "db"})
			if res.HasError() {
				t.Fatalf("unexpected error: %s", res.Error())
			}
			if got := calls.Load(); got != 3 {
				t.Errorf("the API was called %d times, want 3", got)
			}
			if len(*waits) != 2 || (*waits)[0] != 7*time.Second || (*waits)[1] != 7*time.Second {
				t.Errorf("waits = %v, want twice 7s", *waits)
			}
		})
	}
}

func TestRetryPolicyBacksOffWithoutAWaitHeader(t *testing.T) {
	waits := recordSleeps(t, true)
	cc, calls := rateLimitedAPI(t, 3, nil, ``)

	if res := client.Get[map[string]any](t.Context(), cc, "/v2/self"); res.HasError() {
		t.Fatalf("unexpected error: %s", res.Error())
	}
	if got := calls.Load(); got != 4 {
		t.Errorf("the API was called %d times, want 4", got)
	}
	want := []time.Duration{2 * time.Second, 4 * time.Second, 8 * time.Second}
	if len(*waits) != len(want) || (*waits)[0] != want[0] || (*waits)[1] != want[1] || (*waits)[2] != want[2] {
		t.Errorf("waits = %v, want %v", *waits, want)
	}
}

func TestRetryPolicyGivesUpOnARateLimit(t *testing.T) {
	t.Run("still refused after the retries", func(t *testing.T) {
		waits := recordSleeps(t, true)
		cc, calls := rateLimitedAPI(t, 100, http.Header{"Retry-After": {"1"}}, ``)

		res := client.Get[map[string]any](t.Context(), cc, "/v2/self")
		if !res.HasError() || res.StatusCode() != http.StatusTooManyRequests {
			t.Fatalf("expected the 429 to be reported, got %v", res.Error())
		}
		if got := calls.Load(); got != 1+MaxRateLimitRetries {
			t.Errorf("the API was called %d times, want %d", got, 1+MaxRateLimitRetries)
		}
		if len(*waits) != MaxRateLimitRetries {
			t.Errorf("waited %d times, want %d", len(*waits), MaxRateLimitRetries)
		}
	})

	t.Run("asked to wait longer than the provider does", func(t *testing.T) {
		waits := recordSleeps(t, true)
		cc, calls := rateLimitedAPI(t, 100, http.Header{"Retry-After": {"3600"}}, ``)

		res := client.Get[map[string]any](t.Context(), cc, "/v2/self")
		if !res.HasError() || res.StatusCode() != http.StatusTooManyRequests {
			t.Fatalf("expected the 429 to be reported, got %v", res.Error())
		}
		if got := calls.Load(); got != 1 || len(*waits) != 0 {
			t.Errorf("the API was called %d times after %d waits, want 1 call and no wait", got, len(*waits))
		}
	})

	t.Run("the operation is cancelled during the wait", func(t *testing.T) {
		recordSleeps(t, false)
		cc, calls := rateLimitedAPI(t, 100, http.Header{"Retry-After": {"7"}}, ``)

		if res := client.Get[map[string]any](t.Context(), cc, "/v2/self"); !res.HasError() {
			t.Fatal("expected an error")
		}
		if got := calls.Load(); got != 1 {
			t.Errorf("the API was called %d times, want 1", got)
		}
	})
}

func TestSleepStopsWhenTheContextEnds(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	if sleep(ctx, time.Hour) {
		t.Error("sleep reported a full wait on a cancelled context")
	}
	if !sleep(t.Context(), time.Millisecond) {
		t.Error("sleep did not report a full wait")
	}
}
