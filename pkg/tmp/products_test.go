package tmp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"

	"go.clever-cloud.dev/client"
)

func catalogueServer(t *testing.T, handler http.HandlerFunc) *client.Client {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	return client.New(client.WithEndpoint(server.URL))
}

// Twenty resources created at once read the add-on catalogue together: it is
// downloaded once, and the callers arriving during the download wait for it.
func TestGetAddonsProvidersDownloadsTheCatalogueOnce(t *testing.T) {
	var hits atomic.Int32
	cc := catalogueServer(t, func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		if r.URL.Path != "/v2/products/addonproviders" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"id":"redis-addon","plans":[]}]`))
	})

	var wg sync.WaitGroup
	for range 20 {
		wg.Go(func() {
			res := GetAddonsProviders(context.Background(), cc)
			if res.HasError() {
				t.Errorf("unexpected error: %s", res.Error())
				return
			}
			if providers := *res.Payload(); len(providers) != 1 || providers[0].ID != "redis-addon" {
				t.Errorf("unexpected catalogue: %+v", providers)
			}
		})
	}
	wg.Wait()

	if got := hits.Load(); got != 1 {
		t.Errorf("the catalogue was downloaded %d times, want 1", got)
	}
}

// A failed download is not remembered, the next caller gets a fresh attempt.
func TestGetAddonsProvidersDoesNotKeepAFailure(t *testing.T) {
	var hits atomic.Int32
	cc := catalogueServer(t, func(w http.ResponseWriter, r *http.Request) {
		if hits.Add(1) == 1 {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[]`))
	})

	if res := GetAddonsProviders(context.Background(), cc); !res.HasError() {
		t.Fatal("the first download should fail")
	}
	if res := GetAddonsProviders(context.Background(), cc); res.HasError() {
		t.Fatalf("the second download should succeed: %s", res.Error())
	}
	if res := GetAddonsProviders(context.Background(), cc); res.HasError() {
		t.Fatalf("the third call should be answered from memory: %s", res.Error())
	}
	if got := hits.Load(); got != 2 {
		t.Errorf("the API was called %d times, want 2", got)
	}
}

// The runtimes offered depend on the owner asking, and each client has its
// own credentials: neither shares its catalogue with another.
func TestGetProductInstancesKeepsOneCatalogueByOwnerAndClient(t *testing.T) {
	var hits atomic.Int32
	var queries sync.Map
	handler := func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		queries.Store(r.URL.RawQuery, true)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[]`))
	}
	cc, other := catalogueServer(t, handler), catalogueServer(t, handler)
	ctx := context.Background()
	owner := "orga_1"

	for range 3 {
		GetProductInstances(ctx, cc, nil)
		GetProductInstances(ctx, cc, &owner)
		GetProductInstances(ctx, other, &owner)
	}

	if got := hits.Load(); got != 3 {
		t.Errorf("the API was called %d times, want 3", got)
	}
	for _, query := range []string{"", "for=orga_1"} {
		if _, ok := queries.Load(query); !ok {
			t.Errorf("no request with the query %q", query)
		}
	}
}
