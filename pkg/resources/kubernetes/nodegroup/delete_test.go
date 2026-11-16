package nodegroup

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.dev/sdk/models"
)

func init() {
	// no test waits for real between two polls of a deletion
	deletionPollInterval = time.Millisecond
}

func nodegroupWithStatus(status models.NodeGroupStatusType) models.NodeGroup {
	view := nodegroupTestView(nil)
	view.Status = status
	return view
}

// The API only marks the group for deletion. Delete returns once it is gone:
// its name stays taken until then, and a replace creates the new group next.
func TestNodegroupDeleteWaitsUntilTheGroupIsGone(t *testing.T) {
	terminating := nodegroupWithStatus(models.NodeGroupStatusTypeTERMINATING)

	for name, last := range map[string]nodegroupHTTPResponse{
		"not found": {method: http.MethodGet, status: http.StatusNotFound},
		"deleted":   {method: http.MethodGet, status: http.StatusOK, body: nodegroupWithStatus(models.NodeGroupStatusTypeDELETED)},
	} {
		t.Run(name, func(t *testing.T) {
			r := nodegroupTestResource(t,
				nodegroupHTTPResponse{method: http.MethodGet, status: http.StatusOK, body: nodegroupTestView(nil)},
				nodegroupHTTPResponse{method: http.MethodDelete, status: http.StatusOK, body: terminating},
				nodegroupHTTPResponse{method: http.MethodGet, status: http.StatusOK, body: terminating},
				// a failed poll is retried
				nodegroupHTTPResponse{method: http.MethodGet, status: http.StatusBadGateway},
				nodegroupHTTPResponse{method: http.MethodGet, status: http.StatusOK, body: terminating},
				last,
			)
			state, identity := nodegroupTestState(t, r)
			gotState, diags := nodegroupTestOperation(t, r, "Delete", state, identity)
			if diags.HasError() {
				t.Fatalf("unexpected diagnostics: %v", diags)
			}
			if !gotState.Raw.IsNull() {
				t.Error("the deleted group must be removed from state")
			}
		})
	}
}

// A destroy run again after a timeout finds the group already terminating: it
// waits for it without asking for the deletion a second time.
func TestNodegroupDeleteOnlyWaitsForAGroupAlreadyTerminating(t *testing.T) {
	for _, status := range []models.NodeGroupStatusType{models.NodeGroupStatusTypeToTerminate, models.NodeGroupStatusTypeTERMINATING} {
		t.Run(string(status), func(t *testing.T) {
			r := nodegroupTestResource(t,
				nodegroupHTTPResponse{method: http.MethodGet, status: http.StatusOK, body: nodegroupWithStatus(status)},
				nodegroupHTTPResponse{method: http.MethodGet, status: http.StatusNotFound},
			)
			state, identity := nodegroupTestState(t, r)
			gotState, diags := nodegroupTestOperation(t, r, "Delete", state, identity)
			if diags.HasError() || !gotState.Raw.IsNull() {
				t.Fatalf("a terminating group should be waited for then forgotten: %v, %s", diags, gotState.Raw)
			}
		})
	}
}

// A group still there at the deadline stays in state, so that the next destroy
// keeps waiting for it instead of forgetting a group that still holds its name.
func TestNodegroupDeleteKeepsTheStateWhenTheGroupStays(t *testing.T) {
	interval, timeout := deletionPollInterval, deletionTimeout
	deletionPollInterval, deletionTimeout = time.Hour, 50*time.Millisecond
	t.Cleanup(func() { deletionPollInterval, deletionTimeout = interval, timeout })

	terminating := nodegroupWithStatus(models.NodeGroupStatusTypeTERMINATING)
	r := nodegroupTestResource(t,
		nodegroupHTTPResponse{method: http.MethodGet, status: http.StatusOK, body: nodegroupTestView(nil)},
		nodegroupHTTPResponse{method: http.MethodDelete, status: http.StatusOK, body: terminating},
		nodegroupHTTPResponse{method: http.MethodGet, status: http.StatusOK, body: terminating},
	)
	state, identity := nodegroupTestState(t, r)
	gotState, diags := nodegroupTestOperation(t, r, "Delete", state, identity)

	errs := diags.Errors()
	if len(errs) != 1 || errs[0].Summary() != "nodegroup deletion is not finished" || !strings.Contains(errs[0].Detail(), "status TERMINATING") {
		t.Fatalf("expected the unfinished deletion diagnostic with the last status, got %v", diags)
	}
	if !gotState.Raw.Equal(state.Raw) {
		t.Error("a group that is still there must stay in state")
	}
}

// The API cannot change the flavor of a node group, nor move it to another
// cluster: both are planned as a replace. Its name and size are patched in place.
func TestNodegroupReplacedOnFlavorOrClusterChange(t *testing.T) {
	ctx := t.Context()
	r := &ResourceKubernetesNodegroup{}
	var s resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &s)
	// an existing node group: the modifier ignores creations and destructions
	state, _ := nodegroupTestState(t, r)
	plan := tfsdk.Plan{Schema: s.Schema, Raw: state.Raw}

	for name, wantReplace := range map[string]bool{"flavor": true, "kubernetes_id": true, "name": false} {
		attr, ok := s.Schema.Attributes[name].(schema.StringAttribute)
		if !ok {
			t.Fatalf("%s is not a string attribute", name)
		}

		replace := false
		for _, modifier := range attr.PlanModifiers {
			resp := &planmodifier.StringResponse{}
			modifier.PlanModifyString(ctx, planmodifier.StringRequest{
				State:       state,
				Plan:        plan,
				StateValue:  types.StringValue("before"),
				PlanValue:   types.StringValue("after"),
				ConfigValue: types.StringValue("after"),
			}, resp)
			replace = replace || resp.RequiresReplace
		}
		if replace != wantReplace {
			t.Errorf("changing %s: replace = %t, want %t", name, replace, wantReplace)
		}
	}
}
