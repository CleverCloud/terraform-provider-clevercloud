package tmp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go.clever-cloud.dev/client"
)

// fakeElasticsearchAPI answers the cluster creation with the given statuses in
// order, then serves the clusters it holds on list/get.
type fakeElasticsearchAPI struct {
	createStatuses []int
	clusters       []ElasticsearchCluster
	posts          int
}

func (f *fakeElasticsearchAPI) handler(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodPost:
			status := f.createStatuses[f.posts]
			f.posts++
			if status >= 300 {
				w.WriteHeader(status)
				_, _ = w.Write([]byte("There was an internal server error."))
				return
			}
			cluster := ElasticsearchCluster{ID: "elasticsearchCluster_new", Name: "tf-test", CreationDate: time.Now()}
			f.clusters = append(f.clusters, cluster)
			_ = json.NewEncoder(w).Encode(cluster)
		case r.Method == http.MethodGet && r.URL.Path == "/v4/elasticsearch/organisations/orga_test/clusters":
			_ = json.NewEncoder(w).Encode(f.clusters)
		case r.Method == http.MethodGet:
			for _, c := range f.clusters {
				if r.URL.Path == "/v4/elasticsearch/organisations/orga_test/clusters/"+c.ID {
					_ = json.NewEncoder(w).Encode(c)
					return
				}
			}
			w.WriteHeader(http.StatusNotFound)
		default:
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}
}

func TestCreateElasticsearchClusterWithRetry(t *testing.T) {
	req := WannabeElasticsearchCluster{Name: "tf-test", NumberOfNodes: 3, Plan: "M"}

	tests := []struct {
		name       string
		api        *fakeElasticsearchAPI
		wantID     string
		wantErr    bool
		wantStatus int
		wantPosts  int
	}{
		{
			name:      "first attempt succeeds",
			api:       &fakeElasticsearchAPI{createStatuses: []int{201}},
			wantID:    "elasticsearchCluster_new",
			wantPosts: 1,
		},
		{
			name:      "500 then success",
			api:       &fakeElasticsearchAPI{createStatuses: []int{500, 201}},
			wantID:    "elasticsearchCluster_new",
			wantPosts: 2,
		},
		{
			name: "500 after the cluster was provisioned adopts it",
			api: &fakeElasticsearchAPI{
				createStatuses: []int{500},
				clusters:       []ElasticsearchCluster{{ID: "elasticsearchCluster_orphan", Name: "tf-test", CreationDate: time.Now()}},
			},
			wantID:    "elasticsearchCluster_orphan",
			wantPosts: 1,
		},
		{
			name: "older cluster with the same name is not adopted",
			api: &fakeElasticsearchAPI{
				createStatuses: []int{500, 201},
				clusters:       []ElasticsearchCluster{{ID: "elasticsearchCluster_old", Name: "tf-test", CreationDate: time.Now().Add(-time.Hour)}},
			},
			wantID:    "elasticsearchCluster_new",
			wantPosts: 2,
		},
		{
			name:       "4xx is not retried",
			api:        &fakeElasticsearchAPI{createStatuses: []int{400}},
			wantErr:    true,
			wantStatus: 400,
			wantPosts:  1,
		},
		{
			name:       "gives up after three 5xx",
			api:        &fakeElasticsearchAPI{createStatuses: []int{500, 502, 500}},
			wantErr:    true,
			wantStatus: 500,
			wantPosts:  3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(tt.api.handler(t))
			defer srv.Close()
			cc := client.New(client.WithEndpoint(srv.URL), client.WithBearerAuth("test"))

			res := CreateElasticsearchClusterWithRetry(context.Background(), cc, "orga_test", req)

			if tt.api.posts != tt.wantPosts {
				t.Errorf("expected %d creation attempts, got %d", tt.wantPosts, tt.api.posts)
			}
			if tt.wantErr {
				if !res.HasError() {
					t.Fatalf("expected an error, got cluster %s", res.Payload().ID)
				}
				if res.StatusCode() != tt.wantStatus {
					t.Errorf("expected status %d, got %d", tt.wantStatus, res.StatusCode())
				}
				return
			}
			if res.HasError() {
				t.Fatalf("unexpected error: %s", res.Error())
			}
			if res.Payload().ID != tt.wantID {
				t.Errorf("expected cluster %s, got %s", tt.wantID, res.Payload().ID)
			}
		})
	}
}
