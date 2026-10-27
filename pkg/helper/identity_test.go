package helper

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var (
	testStateSchema    = schema.Schema{Attributes: map[string]schema.Attribute{"id": schema.StringAttribute{Computed: true}}}
	testIdentitySchema = identityschema.Schema{Attributes: map[string]identityschema.Attribute{"id": identityschema.StringAttribute{RequiredForImport: true}}}
)

type testID struct {
	ID types.String `tfsdk:"id"`
}

func testState(t *testing.T, id string) tfsdk.State {
	t.Helper()

	state := tfsdk.State{Schema: testStateSchema, Raw: tftypes.NewValue(testStateSchema.Type().TerraformType(t.Context()), nil)}
	if diags := state.Set(t.Context(), testID{ID: types.StringValue(id)}); diags.HasError() {
		t.Fatalf("set state: %v", diags)
	}
	return state
}

func TestIdentityIDFrom_UsesIdentity(t *testing.T) {
	identity := &tfsdk.ResourceIdentity{Schema: testIdentitySchema, Raw: tftypes.NewValue(testIdentitySchema.Type().TerraformType(t.Context()), nil)}
	if diags := identity.Set(t.Context(), testID{ID: types.StringValue("from-identity")}); diags.HasError() {
		t.Fatalf("set identity: %v", diags)
	}

	var diags diag.Diagnostics
	id := IdentityIDFrom(t.Context(), identity, testState(t, "from-state"), &diags)

	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if id.ValueString() != "from-identity" {
		t.Fatalf("got %q, want the identity id", id.ValueString())
	}
}

// clients without identity support hand over a null identity
func TestIdentityIDFrom_NullIdentityFallsBackOnState(t *testing.T) {
	identity := &tfsdk.ResourceIdentity{Schema: testIdentitySchema, Raw: tftypes.NewValue(testIdentitySchema.Type().TerraformType(t.Context()), nil)}

	var diags diag.Diagnostics
	id := IdentityIDFrom(t.Context(), identity, testState(t, "from-state"), &diags)

	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if id.ValueString() != "from-state" {
		t.Fatalf("got %q, want the state id", id.ValueString())
	}
}

func TestIdentityIDFrom_NilIdentityFallsBackOnState(t *testing.T) {
	var diags diag.Diagnostics
	id := IdentityIDFrom(t.Context(), nil, testState(t, "from-state"), &diags)

	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if id.ValueString() != "from-state" {
		t.Fatalf("got %q, want the state id", id.ValueString())
	}
}
