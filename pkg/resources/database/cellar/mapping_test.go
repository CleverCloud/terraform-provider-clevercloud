package cellar_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg/resources/database/cellar"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

func apiAddon() *tmp.AddonResponse {
	return &tmp.AddonResponse{
		ID: "addon_1", Name: "tf-test-cellar", RealID: "cellar_1", Region: "par",
		Plan: tmp.AddonPlan{Slug: "S"},
	}
}

func apiEnv() tmp.EnvVars {
	return tmp.EnvVars{
		{Name: "CELLAR_ADDON_HOST", Value: "cellar-c2.services.clever-cloud.com"},
		{Name: "CELLAR_ADDON_KEY_ID", Value: "key-id"},
		{Name: "CELLAR_ADDON_KEY_SECRET", Value: "key-secret"},
	}
}

func TestCellarFromAPI_PopulatesEveryAttribute(t *testing.T) {
	var diags diag.Diagnostics
	state := &cellar.Cellar{}

	state.
		FromAddon(t.Context(), apiAddon(), &diags).
		FromEnv(t.Context(), apiEnv(), &diags)

	if diags.HasError() {
		t.Fatalf("mapping reported %v", diags.Errors())
	}
	for name, value := range map[string]types.String{
		"name": state.Name, "region": state.Region, "host": state.Host,
		"key_id": state.KeyID, "key_secret": state.KeySecret,
	} {
		if value.IsNull() || value.IsUnknown() {
			t.Errorf("%s is %v, want a value", name, value)
		}
	}
}

func TestCellarFromAPI_PreservesWhatTheAPIDoesNotReturn(t *testing.T) {
	var diags diag.Diagnostics
	state := &cellar.Cellar{KeySecret: types.StringValue("kept-secret")}

	state.FromAddon(t.Context(), apiAddon(), &diags)

	if state.KeySecret.ValueString() != "kept-secret" {
		t.Error("the add-on view carries no credential and must not clear one")
	}
}

func TestCellarFromAPI_NilPayloadIsANoOp(t *testing.T) {
	var diags diag.Diagnostics
	state := &cellar.Cellar{Host: types.StringValue("kept-host")}

	state.
		FromAddon(t.Context(), nil, &diags).
		FromEnv(t.Context(), nil, &diags)

	if diags.HasError() {
		t.Fatalf("a nil payload must not report an error, got %v", diags.Errors())
	}
	if state.Host.ValueString() != "kept-host" {
		t.Error("host must be preserved")
	}
}
