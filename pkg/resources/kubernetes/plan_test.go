package kubernetes

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"go.clever-cloud.com/terraform-provider/pkg/resources"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
	"go.clever-cloud.dev/sdk/models"
)

func simpleTopology(t *testing.T, kind, flavor string, factor int64) types.Object {
	t.Helper()

	var diags diag.Diagnostics
	value := topologyFromAPI(context.Background(), &tmp.KubernetesTopology{Topology: kind, Flavor: flavor, ReplicationFactor: factor}, &diags)
	if diags.HasError() {
		t.Fatalf("failed to build the topology: %v", diags)
	}

	return value
}

func distributedPayload(flavor string, factor int64) *tmp.KubernetesTopology {
	component := tmp.KubernetesTopologyComponent{Flavor: flavor, ReplicationFactor: factor}

	return &tmp.KubernetesTopology{
		Topology: topologyDistributed,
		Components: &tmp.KubernetesTopologyComponents{
			Apiserver:              component,
			CloudControllerManager: component,
			ControllerManager:      component,
			NodeGroupOperator:      component,
			Scheduler:              component,
		},
	}
}

func TestTopologyRoundTrip(t *testing.T) {
	ctx := context.Background()

	for name, payload := range map[string]*tmp.KubernetesTopology{
		"simple":      {Topology: topologyAllInOne, Flavor: "S", ReplicationFactor: 3},
		"distributed": distributedPayload("XS", 2),
	} {
		t.Run(name, func(t *testing.T) {
			var diags diag.Diagnostics
			got := topologyToAPI(ctx, topologyFromAPI(ctx, payload, &diags), &diags)
			if diags.HasError() {
				t.Fatalf("unexpected diagnostics: %v", diags)
			}

			want, _ := json.Marshal(payload)
			gotJSON, _ := json.Marshal(got)
			if string(gotJSON) != string(want) {
				t.Errorf("round trip = %s, want %s", gotJSON, want)
			}
		})
	}

	var diags diag.Diagnostics
	if got := topologyToAPI(ctx, types.ObjectUnknown(topologyAttrTypes), &diags); got != nil {
		t.Errorf("an unknown topology has to be left to the API, got %+v", got)
	}
	if got := topologyFromAPI(ctx, nil, &diags); !got.IsNull() {
		t.Errorf("a missing topology has to be null, got %s", got)
	}
}

// TestTopologyFromAPIKeepsExclusiveFieldsNull pins that a simple topology never
// reads back components: it would differ from the configuration and replace the
// cluster on every plan
func TestTopologyFromAPIKeepsExclusiveFieldsNull(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics
	var simple, distributed Topology

	diags.Append(topologyFromAPI(ctx, &tmp.KubernetesTopology{Topology: topologyDedicatedCompute, Flavor: "XS", ReplicationFactor: 1}, &diags).As(ctx, &simple, asOptions)...)
	diags.Append(topologyFromAPI(ctx, distributedPayload("XS", 1), &diags).As(ctx, &distributed, asOptions)...)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	if !simple.Components.IsNull() {
		t.Errorf("simple topology components = %s, want null", simple.Components)
	}
	if !distributed.Flavor.IsNull() || !distributed.ReplicationFactor.IsNull() {
		t.Errorf("distributed topology flavor = %s, factor = %s, want null", distributed.Flavor, distributed.ReplicationFactor)
	}
}

func TestValidateTopologyShape(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics
	components := topologyAttr(t, distributedPayload("XS", 1), &diags)

	tests := []struct {
		name     string
		topology Topology
		errors   int
	}{
		{name: "simple", topology: Topology{Type: types.StringValue(topologyAllInOne), Flavor: types.StringValue("S"), ReplicationFactor: types.Int64Value(1), Components: types.ObjectNull(topologyComponentsAttrTypes)}},
		{name: "simple without flavor", topology: Topology{Type: types.StringValue(topologyAllInOne), Flavor: types.StringNull(), ReplicationFactor: types.Int64Value(1), Components: types.ObjectNull(topologyComponentsAttrTypes)}, errors: 1},
		{name: "simple without factor", topology: Topology{Type: types.StringValue(topologyDedicatedCompute), Flavor: types.StringValue("S"), ReplicationFactor: types.Int64Null(), Components: types.ObjectNull(topologyComponentsAttrTypes)}, errors: 1},
		{name: "simple with components", topology: Topology{Type: types.StringValue(topologyAllInOne), Flavor: types.StringValue("S"), ReplicationFactor: types.Int64Value(1), Components: components}, errors: 1},
		{name: "distributed", topology: Topology{Type: types.StringValue(topologyDistributed), Flavor: types.StringNull(), ReplicationFactor: types.Int64Null(), Components: components}},
		{name: "distributed without components", topology: Topology{Type: types.StringValue(topologyDistributed), Flavor: types.StringNull(), ReplicationFactor: types.Int64Null(), Components: types.ObjectNull(topologyComponentsAttrTypes)}, errors: 1},
		{name: "distributed with flavor and factor", topology: Topology{Type: types.StringValue(topologyDistributed), Flavor: types.StringValue("S"), ReplicationFactor: types.Int64Value(1), Components: components}, errors: 2},
		{name: "no type", topology: Topology{Type: types.StringNull(), Flavor: types.StringValue("S"), ReplicationFactor: types.Int64Value(1), Components: types.ObjectNull(topologyComponentsAttrTypes)}, errors: 1},
		{name: "distributed without a component", topology: Topology{Type: types.StringValue(topologyDistributed), Flavor: types.StringNull(), ReplicationFactor: types.Int64Null(), Components: withoutScheduler(t, components)}, errors: 1},
		{name: "values only known at apply", topology: Topology{Type: types.StringUnknown(), Flavor: types.StringUnknown(), ReplicationFactor: types.Int64Unknown(), Components: types.ObjectUnknown(topologyComponentsAttrTypes)}},
		{name: "fields only known at apply", topology: Topology{Type: types.StringValue(topologyAllInOne), Flavor: types.StringUnknown(), ReplicationFactor: types.Int64Unknown(), Components: types.ObjectNull(topologyComponentsAttrTypes)}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, d := types.ObjectValueFrom(ctx, topologyAttrTypes, tt.topology)
			if d.HasError() {
				t.Fatalf("failed to build the topology: %v", d)
			}

			var diags diag.Diagnostics
			validateTopologyShape(ctx, value, &diags)
			if diags.ErrorsCount() != tt.errors {
				t.Errorf("got %d errors, want %d: %v", diags.ErrorsCount(), tt.errors, diags)
			}
		})
	}
}

func withoutScheduler(t *testing.T, components types.Object) types.Object {
	t.Helper()

	attributes := components.Attributes()
	attributes["scheduler"] = types.ObjectNull(topologyComponentAttrTypes)
	value, d := types.ObjectValue(topologyComponentsAttrTypes, attributes)
	if d.HasError() {
		t.Fatalf("failed to build the components: %v", d)
	}

	return value
}

func topologyAttr(t *testing.T, payload *tmp.KubernetesTopology, diags *diag.Diagnostics) types.Object {
	t.Helper()

	var topology Topology
	diags.Append(topologyFromAPI(context.Background(), payload, diags).As(context.Background(), &topology, asOptions)...)
	if diags.HasError() {
		t.Fatalf("failed to build the components: %v", *diags)
	}

	return topology.Components
}

func TestPlanChecks(t *testing.T) {
	same := simpleTopology(t, topologyAllInOne, "S", 1)
	other := simpleTopology(t, topologyAllInOne, "M", 1)
	unset := types.ObjectNull(topologyAttrTypes)

	tests := []struct {
		name          string
		plan          Kubernetes
		config        Kubernetes
		state         *Kubernetes
		checkVersion  bool
		checkTopology bool
	}{
		{
			name:   "create without any option leaves everything to the API",
			plan:   Kubernetes{Version: types.StringUnknown(), Topology: types.ObjectUnknown(topologyAttrTypes)},
			config: Kubernetes{Version: types.StringNull(), Topology: unset},
		},
		{
			name:          "create with options",
			plan:          Kubernetes{Version: types.StringValue("1.35"), Topology: same},
			config:        Kubernetes{Version: types.StringValue("1.35"), Topology: same},
			checkVersion:  true,
			checkTopology: true,
		},
		{
			// the platform may stop offering a version a running cluster is on,
			// that must not fail every later plan
			name:   "unchanged values are not checked",
			plan:   Kubernetes{Version: types.StringValue("1.30"), Topology: same},
			config: Kubernetes{Version: types.StringValue("1.30"), Topology: same},
			state:  &Kubernetes{Version: types.StringValue("1.30"), Topology: same},
		},
		{
			name:   "values kept from the state are not checked",
			plan:   Kubernetes{Version: types.StringValue("1.30"), Topology: same},
			config: Kubernetes{Version: types.StringNull(), Topology: unset},
			state:  &Kubernetes{Version: types.StringValue("1.30"), Topology: same},
		},
		{
			name:          "changed values are checked",
			plan:          Kubernetes{Version: types.StringValue("1.36"), Topology: other},
			config:        Kubernetes{Version: types.StringValue("1.36"), Topology: other},
			state:         &Kubernetes{Version: types.StringValue("1.35"), Topology: same},
			checkVersion:  true,
			checkTopology: true,
		},
		{
			name:   "values only known at apply are not checked",
			plan:   Kubernetes{Version: types.StringUnknown(), Topology: types.ObjectUnknown(topologyAttrTypes)},
			config: Kubernetes{Version: types.StringUnknown(), Topology: types.ObjectUnknown(topologyAttrTypes)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkVersion, checkTopology := planChecks(&tt.plan, &tt.config, tt.state)
			if checkVersion != tt.checkVersion || checkTopology != tt.checkTopology {
				t.Errorf("planChecks() = %v, %v, want %v, %v", checkVersion, checkTopology, tt.checkVersion, tt.checkTopology)
			}
		})
	}
}

func TestValidateVersion(t *testing.T) {
	versions := models.VersionsConfig{Available: []string{"1.37", "1.36"}, Default: "1.36"}

	var diags diag.Diagnostics
	validateVersion("1.36", versions, &diags)
	if diags.HasError() {
		t.Errorf("unexpected diagnostics: %v", diags)
	}

	validateVersion("1.99", versions, &diags)
	if diags.ErrorsCount() != 1 {
		t.Fatalf("got %d errors, want 1", diags.ErrorsCount())
	}
	if detail := diags.Errors()[0].Detail(); !strings.Contains(detail, "1.37, 1.36") {
		t.Errorf("the error has to list the available versions, got %q", detail)
	}
}

func TestValidateTopology(t *testing.T) {
	ctx := context.Background()
	constraints := []models.TopologyConstraints{
		{Topology: models.TopologyTypeAllInOne, AvailableFlavors: []models.NodeFlavor{models.NodeFlavorS}, ReplicationFactor: models.ReplicationFactorConfig{Min: 1, Max: 5}},
		{Topology: models.TopologyTypeDISTRIBUTED, AvailableFlavors: []models.NodeFlavor{models.NodeFlavor2XS}, ReplicationFactor: models.ReplicationFactorConfig{Min: 1, Max: 5}},
	}

	tests := []struct {
		name    string
		payload *tmp.KubernetesTopology
		errors  int
	}{
		{name: "valid", payload: &tmp.KubernetesTopology{Topology: topologyAllInOne, Flavor: "S", ReplicationFactor: 1}},
		{name: "flavor of another topology", payload: &tmp.KubernetesTopology{Topology: topologyAllInOne, Flavor: "2XS", ReplicationFactor: 1}, errors: 1},
		{name: "factor out of bounds", payload: &tmp.KubernetesTopology{Topology: topologyAllInOne, Flavor: "S", ReplicationFactor: 6}, errors: 1},
		{name: "topology the API does not describe", payload: &tmp.KubernetesTopology{Topology: topologyDedicatedCompute, Flavor: "XS", ReplicationFactor: 1}},
		{name: "valid components", payload: distributedPayload("2XS", 1)},
		{name: "every component is checked", payload: distributedPayload("S", 9), errors: 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var diags diag.Diagnostics
			validateTopology(ctx, topologyFromAPI(ctx, tt.payload, &diags), constraints, &diags)
			if diags.ErrorsCount() != tt.errors {
				t.Errorf("got %d errors, want %d: %v", diags.ErrorsCount(), tt.errors, diags)
			}
		})
	}
}

func TestCreatePayload(t *testing.T) {
	ctx := context.Background()
	tags, d := types.SetValueFrom(ctx, types.StringType, []string{"env:test"})
	if d.HasError() {
		t.Fatalf("failed to build the tags: %v", d)
	}

	tests := []struct {
		name                 string
		nodeAutoprovisioning bool
		persistentStorage    types.Bool
		region               types.String
		tags                 types.Set
		topology             types.Object
		version              types.String
		expected             string
	}{
		{
			// what a configuration with nothing but a name always sent
			name:              "defaults stay up to the API",
			persistentStorage: types.BoolNull(),
			region:            types.StringNull(),
			tags:              resources.NullTags,
			topology:          types.ObjectNull(topologyAttrTypes),
			version:           types.StringNull(),
			expected:          `{"name":"c"}`,
		},
		{
			name:              "disabled storage is what the API does anyway",
			persistentStorage: types.BoolValue(false),
			region:            types.StringNull(),
			tags:              resources.NullTags,
			topology:          types.ObjectNull(topologyAttrTypes),
			version:           types.StringNull(),
			expected:          `{"name":"c"}`,
		},
		{
			name:                 "every option",
			nodeAutoprovisioning: true,
			persistentStorage:    types.BoolValue(true),
			region:               types.StringValue("par"),
			tags:                 tags,
			topology:             simpleTopology(t, topologyAllInOne, "S", 3),
			version:              types.StringValue("1.35"),
			expected:             `{"features":{"csi":true,"nodeAutoprovisioning":true},"name":"c","tags":["env:test"],"topologyConfig":{"topology":"ALL_IN_ONE","flavor":"S","replicationFactor":3},"version":"1.35","zone":"par"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var diags diag.Diagnostics
			payload := createPayload(ctx, "c", tt.nodeAutoprovisioning, tt.persistentStorage, tt.region, tt.tags, tt.topology, tt.version, &diags)
			if diags.HasError() {
				t.Fatalf("unexpected diagnostics: %v", diags)
			}

			got, err := json.Marshal(payload)
			if err != nil {
				t.Fatalf("Marshal() error = %v", err)
			}
			if string(got) != tt.expected {
				t.Errorf("createPayload() = %s, want %s", got, tt.expected)
			}
		})
	}
}

// TestRegionRequiresReplace pins that only a zone the state knows can replace the
// cluster: the API returns no zone name, so a null state is no proof of a move
// TestRegionPlan runs the plan modifiers of region in order on an existing
// cluster. A configured zone other than the state's replaces it, a region absent
// from the configuration keeps the state's, and one set on a cluster whose zone
// is unknown is refused
func TestRegionPlan(t *testing.T) {
	par, deep, null, unknown := types.StringValue("par"), types.StringValue("deep"), types.StringNull(), types.StringUnknown()
	tests := []struct {
		name        string
		state       types.String
		config      types.String
		wantPlan    types.String
		wantReplace bool
		wantError   bool
	}{
		{name: "unset keeps the state", state: par, config: null, wantPlan: par},
		{name: "unchanged", state: par, config: par, wantPlan: par},
		{name: "moved", state: par, config: deep, wantPlan: deep, wantReplace: true},
		{name: "set on a cluster without a known zone", state: null, config: par, wantError: true},
		{name: "unset on a cluster without a known zone", state: null, config: null, wantPlan: null},
		{name: "unknown until apply", state: par, config: unknown, wantPlan: unknown, wantReplace: true},
	}

	ctx := t.Context()
	var s resource.SchemaResponse
	(&ResourceKubernetes{}).Schema(ctx, resource.SchemaRequest{}, &s)
	region, ok := s.Schema.Attributes["region"].(schema.StringAttribute)
	if !ok {
		t.Fatal("region is not a string attribute")
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := tfsdk.State{Schema: s.Schema}
			if d := state.Set(ctx, Kubernetes{
				ID:       types.StringValue("cluster"),
				Name:     types.StringValue("c"),
				Region:   tt.state,
				Tags:     types.SetNull(types.StringType),
				Topology: types.ObjectNull(topologyAttrTypes),
			}); d.HasError() {
				t.Fatalf("set state: %v", d)
			}
			// the framework plans a computed attribute absent from the configuration as unknown
			planned := tt.config
			if planned.IsNull() {
				planned = unknown
			}
			req := planmodifier.StringRequest{
				Path:        path.Root("region"),
				State:       state,
				Plan:        tfsdk.Plan{Schema: s.Schema, Raw: state.Raw},
				StateValue:  tt.state,
				ConfigValue: tt.config,
				PlanValue:   planned,
			}
			resp := &planmodifier.StringResponse{PlanValue: planned}
			for _, m := range region.PlanModifiers {
				m.PlanModifyString(ctx, req, resp)
				req.PlanValue = resp.PlanValue
			}
			if resp.Diagnostics.HasError() != tt.wantError {
				t.Fatalf("diagnostics %v, want an error: %v", resp.Diagnostics, tt.wantError)
			}
			if tt.wantError {
				return
			}
			if !resp.PlanValue.Equal(tt.wantPlan) || resp.RequiresReplace != tt.wantReplace {
				t.Errorf("plan %v replace %v, want %v replace %v", resp.PlanValue, resp.RequiresReplace, tt.wantPlan, tt.wantReplace)
			}
		})
	}
}

func TestMetadataPatch(t *testing.T) {
	ctx := context.Background()
	name, other := types.StringValue("a"), types.StringValue("b")
	tags, d := types.SetValueFrom(ctx, types.StringType, []string{"x"})
	if d.HasError() {
		t.Fatalf("failed to build the tags: %v", d)
	}
	noTags, d := types.SetValueFrom(ctx, types.StringType, []string{})
	if d.HasError() {
		t.Fatalf("failed to build the tags: %v", d)
	}

	tests := []struct {
		name      string
		newName   types.String
		tags      types.Set
		stateTags types.Set
		expected  string
	}{
		{name: "nothing changed", newName: name, tags: tags, stateTags: tags},
		{name: "unmanaged tags are left alone", newName: name, tags: resources.NullTags, stateTags: tags},
		{name: "rename", newName: other, tags: tags, stateTags: tags, expected: `{"name":"b"}`},
		{name: "new tags", newName: name, tags: tags, stateTags: resources.NullTags, expected: `{"tags":["x"]}`},
		{name: "removing every tag sends an empty list", newName: name, tags: noTags, stateTags: tags, expected: `{"tags":[]}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var diags diag.Diagnostics
			payload := metadataPatch(ctx, tt.newName, name, tt.tags, tt.stateTags, &diags)
			if diags.HasError() {
				t.Fatalf("unexpected diagnostics: %v", diags)
			}

			if tt.expected == "" {
				if payload != nil {
					t.Errorf("metadataPatch() = %+v, want nil", payload)
				}
				return
			}

			got, err := json.Marshal(payload)
			if err != nil {
				t.Fatalf("Marshal() error = %v", err)
			}
			if string(got) != tt.expected {
				t.Errorf("metadataPatch() = %s, want %s", got, tt.expected)
			}
		})
	}
}

func TestVersionLower(t *testing.T) {
	tests := []struct {
		a, b string
		want bool
	}{
		{"1.34", "1.35", true},
		{"1.35", "1.34", false},
		{"1.35", "1.35", false},
		{"1.9", "1.10", true}, // not a string comparison
		{"1.10", "1.9", false},
		{"1.35", "2.0", true},
		{"1.35.2", "1.36", true},
		{"latest", "1.36", false},
		{"1", "1.36", false},
	}

	for _, tt := range tests {
		if got := versionLower(tt.a, tt.b); got != tt.want {
			t.Errorf("versionLower(%s, %s) = %v, want %v", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestResolveValue(t *testing.T) {
	if got := resolveValue(types.StringUnknown(), types.StringValue("1.35")); !got.Equal(types.StringValue("1.35")) {
		t.Errorf("an unknown plan has to fall back on the configuration, got %s", got)
	}
	if got := resolveValue(types.StringValue("1.35"), types.StringNull()); !got.Equal(types.StringValue("1.35")) {
		t.Errorf("a known plan wins, got %s", got)
	}
	if got := resolveValue(types.StringUnknown(), types.StringNull()); !got.IsNull() {
		t.Errorf("an attribute absent from the configuration is null, got %s", got)
	}
}

// TestChangeClusterVersionWaitsForAStableCluster pins the order against the API: it
// answers a 409 to an upgrade unless the cluster is ACTIVE, and reports the
// target version as soon as it accepts one, so neither the reconciliation before
// nor the version field after can be trusted to mean the upgrade is over
func TestChangeClusterVersionWaitsForAStableCluster(t *testing.T) {
	status, version, upgrades := "RECONCILING", "1.35", 0
	polls := 0

	cc := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method == http.MethodPost {
			if status != "ACTIVE" {
				w.WriteHeader(http.StatusConflict)
				return
			}

			upgrades++
			status, version = "UPDATING", "1.36"
			_ = json.NewEncoder(w).Encode(tmp.ClusterView{ID: "cluster", Status: status, Version: version, CreationDate: "2026-01-01T00:00:00Z"})
			return
		}

		polls++
		if polls > 2 {
			status = map[string]string{"RECONCILING": "ACTIVE", "UPDATING": "UPDATING", "ACTIVE": "ACTIVE"}[status]
		}
		if polls > 5 {
			status = "ACTIVE"
		}
		_ = json.NewEncoder(w).Encode(tmp.ClusterView{ID: "cluster", Status: status, Version: version})
	})

	var diags diag.Diagnostics
	if !changeClusterVersion(t.Context(), cc, "orga", "cluster", "1.36", time.Millisecond, time.Second, time.Second, &diags) {
		t.Fatalf("the upgrade failed: %v", diags)
	}
	if upgrades != 1 {
		t.Errorf("sent %d upgrades, want 1", upgrades)
	}
	if polls <= 5 {
		t.Errorf("returned after %d polls, while the cluster was still upgrading", polls)
	}
}

func TestChangeClusterVersionReportsARejection(t *testing.T) {
	cc := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPost {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"code":"clever.kubernetes.invalid-version","message":"Version '1.99' is not available"}`))
			return
		}
		_ = json.NewEncoder(w).Encode(tmp.ClusterView{ID: "cluster", Status: "ACTIVE"})
	})

	var diags diag.Diagnostics
	if changeClusterVersion(t.Context(), cc, "orga", "cluster", "1.99", time.Millisecond, time.Second, time.Second, &diags) {
		t.Fatal("expected the upgrade to fail")
	}
	if diags.ErrorsCount() != 1 {
		t.Errorf("got %d errors, want 1: %v", diags.ErrorsCount(), diags)
	}
}

func TestChangeClusterVersionStopsOnAFailedCluster(t *testing.T) {
	status, version := "ACTIVE", "1.35"
	cc := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPost {
			status, version = "FAILED", "1.36"
		}
		_ = json.NewEncoder(w).Encode(tmp.ClusterView{ID: "cluster", Status: status, Version: version, CreationDate: "2026-01-01T00:00:00Z"})
	})

	var diags diag.Diagnostics
	if changeClusterVersion(t.Context(), cc, "orga", "cluster", "1.36", time.Millisecond, time.Second, time.Minute, &diags) {
		t.Fatal("expected the upgrade to fail")
	}
	if detail := diags.Errors()[0].Detail(); !strings.Contains(detail, "resumed") {
		t.Errorf("the error has to give the resume remedy, got %q", detail)
	}
}

// TestWaitForPersistentStorageWaitsForActive pins that the flag alone is not
// enough: disabling reads false from the first poll while the cluster reconciles,
// and the next operation would be refused
func TestWaitForPersistentStorageWaitsForActive(t *testing.T) {
	polls := 0
	disabled := false
	cc := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		polls++
		status := "RECONCILING"
		if polls > 2 {
			status = "ACTIVE"
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(tmp.ClusterView{ID: "cluster", Status: status, Features: &tmp.KubernetesFeatures{Csi: &disabled}})
	})

	var diags diag.Diagnostics
	if !waitForPersistentStorage(t.Context(), cc, "orga", "cluster", false, time.Millisecond, time.Second, &diags) {
		t.Fatalf("the wait failed: %v", diags)
	}
	if polls != 3 {
		t.Errorf("returned after %d polls, want 3", polls)
	}
}

func TestPatchPersistentStorageSendsOnlyItsFlag(t *testing.T) {
	var body string
	cc := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		var payload json.RawMessage
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Errorf("failed to read the request: %s", err)
		}
		body = string(payload)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(tmp.ClusterView{ID: "cluster", Status: "ACTIVE"})
	})

	var diags diag.Diagnostics
	if !patchPersistentStorage(t.Context(), cc, "orga", "cluster", false, time.Millisecond, time.Second, &diags) {
		t.Fatalf("the patch failed: %v", diags)
	}
	if body != `{"features":{"csi":false}}` {
		t.Errorf("sent %s", body)
	}
}

// modifyPlan runs ModifyPlan on an unconfigured resource, which covers everything
// but the checks against the platform
func modifyPlan(t *testing.T, plan, config Kubernetes, state *Kubernetes) (resource.ModifyPlanResponse, Kubernetes) {
	t.Helper()
	ctx := t.Context()

	r := &ResourceKubernetes{}
	var s resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &s)
	null := tftypes.NewValue(s.Schema.Type().TerraformType(ctx), nil)

	req := resource.ModifyPlanRequest{
		Plan:   tfsdk.Plan{Schema: s.Schema, Raw: null.Copy()},
		Config: tfsdk.Config{Schema: s.Schema, Raw: null.Copy()},
		State:  tfsdk.State{Schema: s.Schema, Raw: null.Copy()},
	}
	if d := req.Plan.Set(ctx, withNulls(plan)); d.HasError() {
		t.Fatalf("set plan: %v", d)
	}
	rawConfig := tfsdk.State{Schema: s.Schema, Raw: null.Copy()}
	if d := rawConfig.Set(ctx, withNulls(config)); d.HasError() {
		t.Fatalf("set config: %v", d)
	}
	req.Config.Raw = rawConfig.Raw
	if state != nil {
		if d := req.State.Set(ctx, withNulls(*state)); d.HasError() {
			t.Fatalf("set state: %v", d)
		}
	}

	resp := resource.ModifyPlanResponse{Plan: req.Plan}
	r.ModifyPlan(ctx, req, &resp)

	var planned Kubernetes
	if d := resp.Plan.Get(ctx, &planned); d.HasError() {
		t.Fatalf("get plan: %v", d)
	}

	return resp, planned
}

// withNulls types the attributes a test case leaves out
func withNulls(k Kubernetes) Kubernetes {
	if k.Tags.ElementType(context.Background()) == nil {
		k.Tags = resources.NullTags
	}
	if k.Topology.AttributeTypes(context.Background()) == nil {
		k.Topology = types.ObjectNull(topologyAttrTypes)
	}

	return k
}

func TestModifyPlanTopology(t *testing.T) {
	upper := simpleTopology(t, topologyAllInOne, "S", 1)
	other := simpleTopology(t, topologyAllInOne, "M", 1)
	unset := types.ObjectNull(topologyAttrTypes)

	tests := []struct {
		name     string
		config   types.Object
		state    types.Object
		replaced bool
		planned  types.Object
	}{
		{name: "same topology", config: upper, state: upper, planned: upper},
		{name: "another topology", config: other, state: upper, replaced: true, planned: other},
		{name: "a state without topology never replaces", config: upper, state: unset, planned: upper},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, planned := modifyPlan(t,
				Kubernetes{Topology: tt.config},
				Kubernetes{Topology: tt.config},
				&Kubernetes{Topology: tt.state},
			)
			if resp.Diagnostics.HasError() {
				t.Fatalf("unexpected diagnostics: %v", resp.Diagnostics)
			}

			if got := len(resp.RequiresReplace) == 1; got != tt.replaced {
				t.Errorf("replaced = %v, want %v", got, tt.replaced)
			}
			if got := resp.Diagnostics.WarningsCount() == 1; got != tt.replaced {
				t.Errorf("warned = %v, a replacement and only a replacement warns: %v", got, resp.Diagnostics)
			}
			if !planned.Topology.Equal(tt.planned) {
				t.Errorf("planned %s, want %s", planned.Topology, tt.planned)
			}
		})
	}
}

func TestModifyPlanWarnsOnDowngrade(t *testing.T) {
	other := simpleTopology(t, topologyAllInOne, "M", 1)
	current := simpleTopology(t, topologyAllInOne, "S", 1)

	tests := []struct {
		name     string
		version  string
		topology types.Object
		warnings int
		detail   string
	}{
		{name: "downgrade", version: "1.35", topology: current, warnings: 1, detail: "from 1.36 to 1.35"},
		{name: "upgrade", version: "1.37", topology: current},
		{name: "same version", version: "1.36", topology: current},
		// only the replacement warns, the new cluster does not move anywhere
		{name: "older version on a replacement", version: "1.35", topology: other, warnings: 1, detail: "creates a new one"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			version := types.StringValue(tt.version)
			resp, _ := modifyPlan(t,
				Kubernetes{Version: version, Topology: tt.topology},
				Kubernetes{Version: version, Topology: tt.topology},
				&Kubernetes{Version: types.StringValue("1.36"), Topology: current},
			)
			if resp.Diagnostics.HasError() {
				t.Fatalf("a downgrade is allowed: %v", resp.Diagnostics)
			}
			if got := resp.Diagnostics.WarningsCount(); got != tt.warnings {
				t.Fatalf("got %d warnings, want %d: %v", got, tt.warnings, resp.Diagnostics)
			}
			if tt.warnings > 0 && !strings.Contains(resp.Diagnostics.Warnings()[0].Detail(), tt.detail) {
				t.Errorf("warning %q, want it to contain %q", resp.Diagnostics.Warnings()[0].Detail(), tt.detail)
			}
		})
	}
}

func TestFlavorValidators(t *testing.T) {
	ctx := context.Background()

	for value, valid := range map[string]bool{"XS": true, "2XS": true, "S": true, "xs": false, "2xs": false, "Xs": false, "": false} {
		var diags diag.Diagnostics
		for _, v := range flavorValidators {
			res := &validator.StringResponse{}
			v.ValidateString(ctx, validator.StringRequest{ConfigValue: types.StringValue(value)}, res)
			diags.Append(res.Diagnostics...)
		}
		if got := !diags.HasError(); got != valid {
			t.Errorf("flavor %q valid = %v, want %v", value, got, valid)
		}
	}
}

func TestVersionFromAPI(t *testing.T) {
	old := types.StringValue("1.35")

	tests := []struct {
		name     string
		status   string
		previous types.String
		want     types.String
		warnings int
	}{
		{name: "rollout over", status: "ACTIVE", previous: old, want: types.StringValue("1.36")},
		{name: "rollout failed", status: "FAILED", previous: old, want: old, warnings: 1},
		{name: "rollout running", status: "UPDATING", previous: old, want: old, warnings: 1},
		{name: "failed with nothing to say about the version", status: "FAILED", previous: types.StringValue("1.36"), want: types.StringValue("1.36")},
		{name: "import of a failed cluster", status: "FAILED", previous: types.StringNull(), want: types.StringValue("1.36")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var diags diag.Diagnostics
			if got := versionFromAPI(tt.status, "1.36", tt.previous, &diags); !got.Equal(tt.want) {
				t.Errorf("versionFromAPI() = %s, want %s", got, tt.want)
			}
			if got := diags.WarningsCount(); got != tt.warnings {
				t.Errorf("got %d warnings, want %d", got, tt.warnings)
			}
		})
	}
}

func clusterPayload(status, version string) tmp.ClusterView {
	return tmp.ClusterView{ID: "cluster", Status: status, Version: version, CreationDate: "2026-01-01T00:00:00Z"}
}

// TestChangeClusterVersionReportsAnAsynchronousRejection pins that a change the
// API refuses after accepting it fails on the spot: the cluster goes back to
// ACTIVE on its previous version instead of going FAILED
func TestChangeClusterVersionReportsAnAsynchronousRejection(t *testing.T) {
	status, version, polls := "ACTIVE", "1.35", 0
	cc := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPost {
			status, version = "UPDATING", "1.36"
			_ = json.NewEncoder(w).Encode(clusterPayload(status, version))
			return
		}

		if status == "UPDATING" {
			polls++
			if polls > 2 {
				status, version = "ACTIVE", "1.35"
			}
		}
		_ = json.NewEncoder(w).Encode(clusterPayload(status, version))
	})

	var diags diag.Diagnostics
	if changeClusterVersion(t.Context(), cc, "orga", "cluster", "1.36", time.Millisecond, time.Second, time.Minute, &diags) {
		t.Fatal("expected the version change to fail")
	}
	if detail := diags.Errors()[0].Detail(); !strings.Contains(detail, "stays on 1.35") {
		t.Errorf("the error has to name the version the cluster stays on, got %q", detail)
	}
}

// TestChangeClusterVersionDoesNotReplayARolloutThatEnded pins the other side of
// keeping the previous version in state during a rollout: the plan can show a
// change that is already done by the time it is applied
func TestChangeClusterVersionDoesNotReplayARolloutThatEnded(t *testing.T) {
	cc := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			t.Error("the cluster already runs that version, nothing has to be sent")
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(clusterPayload("ACTIVE", "1.36"))
	})

	var diags diag.Diagnostics
	if !changeClusterVersion(t.Context(), cc, "orga", "cluster", "1.36", time.Millisecond, time.Second, time.Second, &diags) {
		t.Fatalf("unexpected failure: %v", diags)
	}
}

func TestUpgradeStateFromV0(t *testing.T) {
	ctx := t.Context()
	current := Kubernetes{}
	var s resource.SchemaResponse
	(&ResourceKubernetes{}).Schema(ctx, resource.SchemaRequest{}, &s)

	run := func(t *testing.T, r *ResourceKubernetes) (Kubernetes, diag.Diagnostics) {
		t.Helper()

		prior := tfsdk.State{Schema: schemaKubernetesV0, Raw: tftypes.NewValue(schemaKubernetesV0.Type().TerraformType(ctx), nil)}
		if d := prior.Set(ctx, KubernetesV0{
			ID:                   types.StringValue("cluster"),
			Name:                 types.StringValue("c"),
			KubeConfig:           types.StringValue("kubeconfig"),
			NodeAutoprovisioning: types.BoolValue(true),
		}); d.HasError() {
			t.Fatalf("set prior state: %v", d)
		}

		res := resource.UpgradeStateResponse{State: tfsdk.State{Schema: s.Schema, Raw: tftypes.NewValue(s.Schema.Type().TerraformType(ctx), nil)}}
		r.UpgradeState(ctx)[0].StateUpgrader(ctx, resource.UpgradeStateRequest{State: &prior}, &res)

		got := current
		if d := res.State.Get(ctx, &got); d.HasError() {
			t.Fatalf("get state: %v", d)
		}

		return got, res.Diagnostics
	}

	keepsWhatItHad := func(t *testing.T, got Kubernetes) {
		t.Helper()
		if got.ID.ValueString() != "cluster" || got.Name.ValueString() != "c" || got.KubeConfig.ValueString() != "kubeconfig" || !got.NodeAutoprovisioning.ValueBool() {
			t.Errorf("lost part of the prior state: %+v", got)
		}
	}

	t.Run("reads the new attributes from the API", func(t *testing.T) {
		enabled := true
		cc := testClient(t, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			view := clusterPayload("ACTIVE", "1.36")
			view.Features = &tmp.KubernetesFeatures{Csi: &enabled}
			view.Topology = &tmp.KubernetesTopology{Topology: topologyAllInOne, Flavor: "S", ReplicationFactor: 1}
			_ = json.NewEncoder(w).Encode(view)
		})
		r := &ResourceKubernetes{}
		r.Provider = clusterReadTestProvider{client: cc}

		got, diags := run(t, r)
		if diags.HasError() {
			t.Fatalf("unexpected diagnostics: %v", diags)
		}
		keepsWhatItHad(t, got)
		if got.Version.ValueString() != "1.36" || !got.PersistentStorage.ValueBool() || !got.Topology.Equal(simpleTopology(t, topologyAllInOne, "S", 1)) {
			t.Errorf("did not read the API: %+v", got)
		}
		if !got.Tags.IsNull() {
			t.Errorf("tags are not managed until configured, got %s", got.Tags)
		}
	})

	t.Run("does not fail without a configured provider", func(t *testing.T) {
		got, diags := run(t, &ResourceKubernetes{})
		if diags.HasError() {
			t.Fatalf("unexpected diagnostics: %v", diags)
		}
		keepsWhatItHad(t, got)
		if !got.Topology.IsNull() || !got.Version.IsNull() {
			t.Errorf("expected the new attributes to stay null: %+v", got)
		}
	})

	t.Run("does not fail when the API refuses", func(t *testing.T) {
		cc := testClient(t, func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusInternalServerError) })
		r := &ResourceKubernetes{}
		r.Provider = clusterReadTestProvider{client: cc}

		got, diags := run(t, r)
		if diags.HasError() {
			t.Fatalf("unexpected diagnostics: %v", diags)
		}
		keepsWhatItHad(t, got)
		if !got.Topology.IsNull() {
			t.Errorf("expected the topology to stay null: %s", got.Topology)
		}
	})
}
