package nodegroup

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg/provider"
	"go.clever-cloud.dev/client"
	"go.clever-cloud.dev/sdk/models"
)

const nodegroupTestPath = "/v4/kubernetes/organisations/test-owner/clusters/test-cluster/node-groups/test-nodegroup"

type nodegroupTestProvider struct {
	provider.Provider
	client *client.Client
}

func (p nodegroupTestProvider) Organization() string   { return "test-owner" }
func (p nodegroupTestProvider) Client() *client.Client { return p.client }

type nodegroupHTTPResponse struct {
	method string
	status int
	body   any
	check  func(*http.Request)
}

// The ordered responses ensure a mutation cannot happen without its own GET,
// including when Terraform skips refresh before applying an existing plan.
func nodegroupTestResource(t *testing.T, responses ...nodegroupHTTPResponse) *ResourceKubernetesNodegroup {
	t.Helper()
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		i := int(calls.Add(1)) - 1
		if i >= len(responses) {
			t.Errorf("unexpected request: %s %s", req.Method, req.URL.Path)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		response := responses[i]
		wantPath := nodegroupTestPath
		if response.method == http.MethodPost {
			wantPath = "/v4/kubernetes/organisations/test-owner/clusters/test-cluster/node-groups"
		}
		if req.Method != response.method || req.URL.Path != wantPath {
			t.Errorf("request %d = %s %s, want %s %s", i+1, req.Method, req.URL.Path, response.method, wantPath)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if response.check != nil {
			response.check(req)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(response.status)
		if response.body != nil {
			if err := json.NewEncoder(w).Encode(response.body); err != nil {
				t.Errorf("encode API response: %s", err)
			}
		}
	}))
	t.Cleanup(func() {
		server.Close()
		if got := int(calls.Load()); got != len(responses) {
			t.Errorf("API received %d requests, want %d", got, len(responses))
		}
	})
	r := &ResourceKubernetesNodegroup{}
	var response resource.ConfigureResponse
	r.Configure(t.Context(), resource.ConfigureRequest{
		ProviderData: nodegroupTestProvider{client: client.New(client.WithEndpoint(server.URL))},
	}, &response)
	if response.Diagnostics.HasError() {
		t.Fatalf("configure resource: %v", response.Diagnostics)
	}
	return r
}

func nodegroupTestState(t *testing.T, r *ResourceKubernetesNodegroup) (tfsdk.State, *tfsdk.ResourceIdentity) {
	t.Helper()
	var schema resource.SchemaResponse
	r.Schema(t.Context(), resource.SchemaRequest{}, &schema)
	state := tfsdk.State{Schema: schema.Schema}
	if diags := state.Set(t.Context(), KubernetesNodegroup{
		ID:           types.StringValue("test-nodegroup"),
		KubernetesID: types.StringValue("test-cluster"),
		Name:         types.StringValue("static-workers"),
		Flavor:       types.StringValue("S"),
		Size:         types.Int64Value(1),
	}); diags.HasError() {
		t.Fatalf("create Terraform state: %v", diags)
	}
	var identitySchema resource.IdentitySchemaResponse
	r.IdentitySchema(t.Context(), resource.IdentitySchemaRequest{}, &identitySchema)
	identity := &tfsdk.ResourceIdentity{Schema: identitySchema.IdentitySchema}
	if diags := identity.Set(t.Context(), KubernetesNodegroupIdentity{ID: types.StringValue("test-nodegroup")}); diags.HasError() {
		t.Fatalf("create Terraform identity: %v", diags)
	}
	return state, identity
}

func nodegroupTestView(labels models.MapLabelkeyLabelvalue) models.NodeGroup {
	return models.NodeGroup{
		ID:              "test-nodegroup",
		ClusterID:       "test-cluster",
		Name:            "static-workers",
		Flavor:          models.NodeFlavor("S"),
		TargetNodeCount: 1,
		Status:          models.NodeGroupStatusTypeREADY,
		Labels:          labels,
	}
}

func nodegroupTestOperation(t *testing.T, r *ResourceKubernetesNodegroup, operation string, state tfsdk.State, identity *tfsdk.ResourceIdentity) (tfsdk.State, diag.Diagnostics) {
	t.Helper()
	switch operation {
	case "Read":
		response := resource.ReadResponse{State: state, Identity: identity}
		r.Read(t.Context(), resource.ReadRequest{State: state, Identity: identity}, &response)
		return response.State, response.Diagnostics
	case "Update":
		var planned KubernetesNodegroup
		if diags := state.Get(t.Context(), &planned); diags.HasError() {
			t.Fatalf("read state for update plan: %v", diags)
		}
		planned.Name = types.StringValue("updated-workers")
		planned.Size = types.Int64Value(2)
		plan := tfsdk.Plan{Schema: state.Schema}
		if diags := plan.Set(t.Context(), planned); diags.HasError() {
			t.Fatalf("create update plan: %v", diags)
		}
		// Match the framework's initialization from prior state. A rejected update
		// must not replace it with planned values that were never applied.
		response := resource.UpdateResponse{State: state, Identity: identity}
		r.Update(t.Context(), resource.UpdateRequest{
			State: state, Identity: identity, Plan: plan,
		}, &response)
		return response.State, response.Diagnostics
	case "Delete":
		response := resource.DeleteResponse{State: state, Identity: identity}
		r.Delete(t.Context(), resource.DeleteRequest{State: state, Identity: identity}, &response)
		return response.State, response.Diagnostics
	default:
		t.Fatalf("unknown operation %s", operation)
		return tfsdk.State{}, nil
	}
}

func TestNodegroupRejectsKarpenterGroups(t *testing.T) {
	for _, operation := range []string{"Read", "Update", "Delete"} {
		t.Run(operation, func(t *testing.T) {
			view := nodegroupTestView(models.MapLabelkeyLabelvalue{"karpenter.sh/nodepool": "default"})
			r := nodegroupTestResource(t, nodegroupHTTPResponse{method: http.MethodGet, status: http.StatusOK, body: view})
			state, identity := nodegroupTestState(t, r)
			gotState, diags := nodegroupTestOperation(t, r, operation, state, identity)
			if errors := diags.Errors(); len(errors) != 1 || errors[0].Summary() != "nodegroup carries Karpenter labels" {
				t.Fatalf("expected Karpenter ownership diagnostic, got %v", diags)
			}
			if !gotState.Raw.Equal(state.Raw) {
				t.Error("refusing a Karpenter group must leave Terraform state unchanged")
			}
		})
	}
}

func TestNodegroupAllowsStaticGroups(t *testing.T) {
	for _, tc := range []struct {
		name   string
		labels models.MapLabelkeyLabelvalue
	}{
		{name: "nil labels"},
		{name: "empty labels", labels: models.MapLabelkeyLabelvalue{}},
		{name: "unrelated labels", labels: models.MapLabelkeyLabelvalue{"team": "platform"}},
		{name: "empty nodepool", labels: models.MapLabelkeyLabelvalue{"karpenter.sh/nodepool": ""}},
	} {
		for _, operation := range []string{"Read", "Update", "Delete"} {
			t.Run(tc.name+"/"+operation, func(t *testing.T) {
				view := nodegroupTestView(tc.labels)
				responses := []nodegroupHTTPResponse{{method: http.MethodGet, status: http.StatusOK, body: view}}
				switch operation {
				case "Update":
					view.Name, view.TargetNodeCount = "updated-workers", 2
					responses = append(responses, nodegroupHTTPResponse{
						method: http.MethodPatch, status: http.StatusOK, body: view,
						check: func(req *http.Request) {
							var payload models.NodeGroupPatchPayload
							if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
								t.Errorf("decode update payload: %s", err)
								return
							}
							if payload.Name != "updated-workers" || payload.TargetNodeCount != 2 ||
								payload.MinNodeCount == nil || *payload.MinNodeCount != 2 ||
								payload.MaxNodeCount == nil || *payload.MaxNodeCount != 2 {
								t.Errorf("unexpected update payload: %+v", payload)
							}
						},
					})
				case "Delete":
					responses = append(responses, nodegroupHTTPResponse{method: http.MethodDelete, status: http.StatusOK, body: view})
				}
				r := nodegroupTestResource(t, responses...)
				state, identity := nodegroupTestState(t, r)
				gotState, diags := nodegroupTestOperation(t, r, operation, state, identity)
				if diags.HasError() {
					t.Fatalf("static nodegroup should remain manageable: %v", diags)
				}
				if operation == "Delete" {
					if !gotState.Raw.IsNull() {
						t.Error("successful deletion must remove the resource from state")
					}
				} else {
					var got KubernetesNodegroup
					if diags := gotState.Get(t.Context(), &got); diags.HasError() {
						t.Fatalf("decode resulting state: %v", diags)
					}
					if got.Name.ValueString() != view.Name || got.Size.ValueInt64() != int64(view.TargetNodeCount) ||
						got.ID.ValueString() != view.ID || got.KubernetesID.ValueString() != view.ClusterID ||
						got.Flavor.ValueString() != string(view.Flavor) {
						t.Errorf("unexpected state after %s: %+v", operation, got)
					}
				}
			})
		}
	}
}

func TestNodegroupGETFailurePreservesState(t *testing.T) {
	for _, operation := range []string{"Read", "Update", "Delete"} {
		for _, status := range []int{http.StatusUnauthorized, http.StatusForbidden, http.StatusInternalServerError} {
			t.Run(operation+"/"+http.StatusText(status), func(t *testing.T) {
				r := nodegroupTestResource(t, nodegroupHTTPResponse{method: http.MethodGet, status: status})
				state, identity := nodegroupTestState(t, r)
				gotState, diags := nodegroupTestOperation(t, r, operation, state, identity)
				if !diags.HasError() {
					t.Fatal("failed ownership lookup must return an error")
				}
				if !gotState.Raw.Equal(state.Raw) {
					t.Error("failed ownership lookup must leave Terraform state unchanged")
				}
			})
		}
	}
}

func TestNodegroupAlreadyDeleted(t *testing.T) {
	// The API can retain a DELETED group and its old labels instead of returning 404.
	deleted := nodegroupTestView(nil)
	deleted.Status = models.NodeGroupStatusTypeDELETED
	deletedKarpenter := deleted
	deletedKarpenter.Labels = models.MapLabelkeyLabelvalue{"karpenter.sh/nodepool": "default"}
	for _, operation := range []string{"Read", "Update", "Delete"} {
		for _, tc := range []struct {
			name     string
			response nodegroupHTTPResponse
		}{
			{name: "not found", response: nodegroupHTTPResponse{method: http.MethodGet, status: http.StatusNotFound}},
			{name: "deleted static group", response: nodegroupHTTPResponse{method: http.MethodGet, status: http.StatusOK, body: deleted}},
			{name: "deleted Karpenter group", response: nodegroupHTTPResponse{method: http.MethodGet, status: http.StatusOK, body: deletedKarpenter}},
		} {
			t.Run(operation+"/"+tc.name, func(t *testing.T) {
				r := nodegroupTestResource(t, tc.response)
				state, identity := nodegroupTestState(t, r)
				gotState, diags := nodegroupTestOperation(t, r, operation, state, identity)
				if operation == "Update" {
					if errors := diags.Errors(); len(errors) != 1 || errors[0].Summary() != "nodegroup no longer exists" {
						t.Fatalf("expected missing nodegroup diagnostic, got %v", diags)
					}
					if detail := diags.Errors()[0].Detail(); detail != "Node group \"test-nodegroup\" no longer exists. Run a new Terraform plan with refresh enabled to reconcile its deletion before applying changes." {
						t.Errorf("missing nodegroup diagnostic must explain how to reconcile state, got %q", detail)
					}
					if !gotState.Raw.Equal(state.Raw) {
						t.Error("refusing an already deleted group must leave Terraform state unchanged")
					}
					return
				}
				if diags.HasError() {
					t.Fatalf("an already deleted group should be forgotten: %v", diags)
				}
				if !gotState.Raw.IsNull() {
					t.Error("an already deleted group must be removed from state")
				}
			})
		}
	}
}

func TestNodegroupDeleteDisappearsAfterLookup(t *testing.T) {
	r := nodegroupTestResource(t,
		nodegroupHTTPResponse{method: http.MethodGet, status: http.StatusOK, body: nodegroupTestView(nil)},
		nodegroupHTTPResponse{method: http.MethodDelete, status: http.StatusNotFound},
	)
	state, identity := nodegroupTestState(t, r)
	gotState, diags := nodegroupTestOperation(t, r, "Delete", state, identity)
	if diags.HasError() || !gotState.Raw.IsNull() {
		t.Fatalf("DELETE 404 after a successful GET should remove state: %v, %s", diags, gotState.Raw)
	}
}

func TestNodegroupMutationFailurePreservesState(t *testing.T) {
	for _, tc := range []struct{ operation, method string }{{"Update", http.MethodPatch}, {"Delete", http.MethodDelete}} {
		t.Run(tc.operation, func(t *testing.T) {
			r := nodegroupTestResource(t,
				nodegroupHTTPResponse{method: http.MethodGet, status: http.StatusOK, body: nodegroupTestView(nil)},
				nodegroupHTTPResponse{method: tc.method, status: http.StatusConflict},
			)
			state, identity := nodegroupTestState(t, r)
			gotState, diags := nodegroupTestOperation(t, r, tc.operation, state, identity)
			if !diags.HasError() || !gotState.Raw.Equal(state.Raw) {
				t.Fatalf("failed mutation must report an error and preserve state: %v, %s", diags, gotState.Raw)
			}
		})
	}
}

func TestNodegroupCreateDoesNotRequireClusterFeatureLookup(t *testing.T) {
	view := nodegroupTestView(nil)
	r := nodegroupTestResource(t, nodegroupHTTPResponse{
		method: http.MethodPost, status: http.StatusCreated, body: view,
		check: func(req *http.Request) {
			var payload models.NodeGroupCreationPayload
			if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
				t.Errorf("decode creation payload: %s", err)
				return
			}
			if payload.Name != view.Name || payload.TargetNodeCount != 1 || payload.Flavor != view.Flavor {
				t.Errorf("unexpected creation payload: %+v", payload)
			}
			if payload.Labels != nil && len(*payload.Labels) != 0 {
				t.Errorf("fixed group must not receive Karpenter labels: %v", *payload.Labels)
			}
		},
	})
	state, identity := nodegroupTestState(t, r)
	response := resource.CreateResponse{State: state, Identity: &tfsdk.ResourceIdentity{Schema: identity.Schema}}
	r.Create(t.Context(), resource.CreateRequest{Plan: tfsdk.Plan(state)}, &response)
	if response.Diagnostics.HasError() {
		t.Fatalf("create fixed group: %v", response.Diagnostics)
	}
	if !response.State.Raw.Equal(state.Raw) || !response.Identity.Raw.Equal(identity.Raw) {
		t.Fatalf("creation did not save the nodegroup state and identity: %s, %s", response.State.Raw, response.Identity.Raw)
	}
}
