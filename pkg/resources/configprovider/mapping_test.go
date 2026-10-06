package configprovider_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg/resources/configprovider"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

func TestConfigProviderFromAPI_PopulatesEveryAttribute(t *testing.T) {
	var diags diag.Diagnostics
	state := &configprovider.ConfigProvider{}

	state.
		FromAddon(t.Context(), &tmp.AddonResponse{Name: "tf-test-cp"}, &diags).
		FromEnv(t.Context(), tmp.EnvVars{{Name: "SOME_VAR", Value: "value"}}, &diags)

	if diags.HasError() {
		t.Fatalf("mapping reported %v", diags.Errors())
	}
	if state.Name.IsNull() {
		t.Error("name must be set — the env endpoint does not carry it, which is what left an import with a null required attribute")
	}
	if state.Environment.IsNull() {
		t.Error("environment must be set")
	}
}

func TestConfigProviderFromAPI_PreservesWhatTheAPIDoesNotReturn(t *testing.T) {
	var diags diag.Diagnostics
	state := &configprovider.ConfigProvider{Name: types.StringValue("kept-name")}

	state.FromEnv(t.Context(), tmp.EnvVars{{Name: "SOME_VAR", Value: "value"}}, &diags)

	if state.Name.ValueString() != "kept-name" {
		t.Error("the env payload has no name and must not clear it")
	}
}

func TestConfigProviderFromAPI_NilPayloadIsANoOp(t *testing.T) {
	var diags diag.Diagnostics
	state := &configprovider.ConfigProvider{Name: types.StringValue("kept-name")}

	state.
		FromAddon(t.Context(), nil, &diags).
		FromEnv(t.Context(), nil, &diags)

	if diags.HasError() {
		t.Fatalf("a nil payload must not report an error, got %v", diags.Errors())
	}
	if state.Name.ValueString() != "kept-name" {
		t.Error("name must be preserved")
	}
}
