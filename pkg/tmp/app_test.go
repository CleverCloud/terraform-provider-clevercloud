package tmp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.clever-cloud.dev/client"
)

func TestRestartApp(t *testing.T) {
	tests := []struct {
		name         string
		withoutCache bool
		wantUseCache string
	}{
		{name: "restart from the build cache", withoutCache: false, wantUseCache: ""},
		{name: "rebuild without the build cache", withoutCache: true, wantUseCache: "no"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotUseCache string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost || r.URL.Path != "/v2/organisations/orga_test/applications/app_test/instances" {
					t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
				}
				gotUseCache = r.URL.Query().Get("useCache")
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(RestartAppRes{DeploymentID: "deployment_test"})
			}))
			defer srv.Close()
			cc := client.New(client.WithEndpoint(srv.URL), client.WithBearerAuth("test"))

			res := RestartApp(context.Background(), cc, "orga_test", "app_test", tt.withoutCache)
			if res.HasError() {
				t.Fatalf("unexpected error: %s", res.Error())
			}
			if gotUseCache != tt.wantUseCache {
				t.Errorf("useCache = %q, want %q", gotUseCache, tt.wantUseCache)
			}
		})
	}
}
