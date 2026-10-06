package metabase_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg/resources/software/metabase"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

func TestMetabaseFromAPI_PopulatesEveryAttribute(t *testing.T) {
	var diags diag.Diagnostics
	state := &metabase.Metabase{}

	state.FromAddon(t.Context(), &tmp.AddonResponse{Name: "tf-test-metabase", Region: "par"}, &diags)
	state.FromMetabase(t.Context(), &tmp.Metabase{AccessURL: "https://metabase.example.com"}, &diags)

	if diags.HasError() {
		t.Fatalf("mapping reported %v", diags.Errors())
	}
	for name, value := range map[string]types.String{
		"name": state.Name, "region": state.Region, "host": state.Host,
	} {
		if value.IsNull() || value.IsUnknown() {
			t.Errorf("%s is %v, want a value", name, value)
		}
	}
}

func TestMetabaseFromAPI_PreservesWhatTheAPIDoesNotReturn(t *testing.T) {
	var diags diag.Diagnostics
	state := &metabase.Metabase{Region: types.StringValue("mtl")}

	state.FromMetabase(t.Context(), &tmp.Metabase{AccessURL: "https://metabase.example.com"}, &diags)

	if state.Region.ValueString() != "mtl" {
		t.Error("region must be preserved by the product mapper")
	}
}

func TestMetabaseFromAPI_NilPayloadIsANoOp(t *testing.T) {
	var diags diag.Diagnostics
	state := &metabase.Metabase{Host: types.StringValue("kept-host")}

	state.FromAddon(t.Context(), nil, &diags)
	state.FromMetabase(t.Context(), nil, &diags)

	if diags.HasError() {
		t.Fatalf("a nil payload must not report an error, got %v", diags.Errors())
	}
	if state.Host.ValueString() != "kept-host" {
		t.Error("host must be preserved")
	}
}
