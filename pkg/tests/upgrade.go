package tests

import (
	"slices"
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// UpgradeResourceState replays the UpgradeResourceState RPC Terraform issues when a
// stored instance carries a schema_version older than the resource schema.
// rawJSON is the instance "attributes" object as found in a state file.
// The upgraded state is decoded against the current resource schema.
func UpgradeResourceState(t *testing.T, typeName string, fromVersion int64, rawJSON string) tftypes.Value {
	t.Helper()
	ctx := t.Context()

	server, err := ProtoV6Provider["clevercloud"]()
	if err != nil {
		t.Fatalf("provider server: %s", err)
	}

	res, err := server.UpgradeResourceState(ctx, &tfprotov6.UpgradeResourceStateRequest{
		TypeName: typeName,
		Version:  fromVersion,
		RawState: &tfprotov6.RawState{JSON: []byte(rawJSON)},
	})
	if err != nil {
		t.Fatalf("UpgradeResourceState(%s): %s", typeName, err)
	}
	for _, d := range res.Diagnostics {
		if d.Severity == tfprotov6.DiagnosticSeverityError {
			t.Fatalf("UpgradeResourceState(%s): %s: %s", typeName, d.Summary, d.Detail)
		}
	}

	schemas, err := server.GetProviderSchema(ctx, &tfprotov6.GetProviderSchemaRequest{})
	if err != nil {
		t.Fatalf("GetProviderSchema: %s", err)
	}
	schema, ok := schemas.ResourceSchemas[typeName]
	if !ok {
		t.Fatalf("no schema registered for %s", typeName)
	}

	state, err := res.UpgradedState.Unmarshal(schema.ValueType())
	if err != nil {
		t.Fatalf("decode upgraded state of %s: %s", typeName, err)
	}
	return state
}

// StateAttr returns one top-level attribute of a decoded state.
func StateAttr(t *testing.T, state tftypes.Value, name string) tftypes.Value {
	t.Helper()

	attrs := map[string]tftypes.Value{}
	if err := state.As(&attrs); err != nil {
		t.Fatalf("state is not an object: %s", err)
	}
	v, ok := attrs[name]
	if !ok {
		t.Fatalf("no attribute %q in state", name)
	}
	return v
}

// StateString returns a known, non-null string attribute.
func StateString(t *testing.T, state tftypes.Value, name string) string {
	t.Helper()

	var s string
	if err := StateAttr(t, state, name).As(&s); err != nil {
		t.Fatalf("attribute %q: %s", name, err)
	}
	return s
}

// StateNestedString returns a known, non-null string attribute of a nested object
// or block, such as deployment.repository.
func StateNestedString(t *testing.T, state tftypes.Value, object, name string) string {
	t.Helper()

	attrs := map[string]tftypes.Value{}
	if err := StateAttr(t, state, object).As(&attrs); err != nil {
		t.Fatalf("%s: %s", object, err)
	}
	var s string
	if err := attrs[name].As(&s); err != nil {
		t.Fatalf("%s.%s: %s", object, name, err)
	}
	return s
}

// StateStrings returns a set or list of strings attribute, sorted, nil when null.
func StateStrings(t *testing.T, state tftypes.Value, name string) []string {
	t.Helper()

	v := StateAttr(t, state, name)
	if v.IsNull() {
		return nil
	}
	elems := []tftypes.Value{}
	if err := v.As(&elems); err != nil {
		t.Fatalf("attribute %q: %s", name, err)
	}
	out := []string{}
	for _, e := range elems {
		var s string
		if err := e.As(&s); err != nil {
			t.Fatalf("attribute %q element: %s", name, err)
		}
		out = append(out, s)
	}
	slices.Sort(out)
	return out
}

// StateObjects returns a set or list of objects attribute as one map per element,
// holding its non-null string attributes. nil when null.
func StateObjects(t *testing.T, state tftypes.Value, name string) []map[string]string {
	t.Helper()

	v := StateAttr(t, state, name)
	if v.IsNull() {
		return nil
	}
	elems := []tftypes.Value{}
	if err := v.As(&elems); err != nil {
		t.Fatalf("attribute %q: %s", name, err)
	}
	out := []map[string]string{}
	for _, e := range elems {
		obj := map[string]tftypes.Value{}
		if err := e.As(&obj); err != nil {
			t.Fatalf("attribute %q element: %s", name, err)
		}
		m := map[string]string{}
		for k, av := range obj {
			if av.IsNull() {
				continue
			}
			var s string
			if err := av.As(&s); err != nil {
				t.Fatalf("attribute %q element %q: %s", name, k, err)
			}
			m[k] = s
		}
		out = append(out, m)
	}
	return out
}

// StateVHosts returns the vhosts attribute as fqdn -> path_begin, nil when null.
func StateVHosts(t *testing.T, state tftypes.Value) map[string]string {
	t.Helper()

	objects := StateObjects(t, state, "vhosts")
	if objects == nil {
		return nil
	}
	out := map[string]string{}
	for _, vhost := range objects {
		out[vhost["fqdn"]] = vhost["path_begin"]
	}
	return out
}
