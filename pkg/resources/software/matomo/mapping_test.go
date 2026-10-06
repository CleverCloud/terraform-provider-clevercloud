package matomo_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg/resources/software/matomo"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

func TestMatomoFromAPI_PopulatesEveryAttribute(t *testing.T) {
	var diags diag.Diagnostics
	state := &matomo.Matomo{}

	state.FromAddon(t.Context(), &tmp.AddonResponse{Name: "tf-test-matomo", Region: "par"}, &diags)
	state.FromMatomo(t.Context(), &tmp.Matomo{AccessURL: "https://matomo.example.com", Version: "5.1"}, &diags)

	if diags.HasError() {
		t.Fatalf("mapping reported %v", diags.Errors())
	}
	for name, value := range map[string]types.String{
		"name": state.Name, "region": state.Region,
		"host": state.Host, "version": state.Version,
	} {
		if value.IsNull() || value.IsUnknown() {
			t.Errorf("%s is %v, want a value", name, value)
		}
	}
}

// The product view carries no region, which is why Read also reads the add-on.
func TestMatomoFromAPI_PreservesWhatTheAPIDoesNotReturn(t *testing.T) {
	var diags diag.Diagnostics
	state := &matomo.Matomo{Region: types.StringValue("mtl")}

	state.FromMatomo(t.Context(), &tmp.Matomo{AccessURL: "https://matomo.example.com"}, &diags)

	if state.Region.ValueString() != "mtl" {
		t.Error("region must be preserved by the product mapper")
	}
}

func TestMatomoFromAPI_NilPayloadIsANoOp(t *testing.T) {
	var diags diag.Diagnostics
	state := &matomo.Matomo{Host: types.StringValue("kept-host")}

	state.FromAddon(t.Context(), nil, &diags)
	state.FromMatomo(t.Context(), nil, &diags)

	if diags.HasError() {
		t.Fatalf("a nil payload must not report an error, got %v", diags.Errors())
	}
	if state.Host.ValueString() != "kept-host" {
		t.Error("host must be preserved")
	}
}
