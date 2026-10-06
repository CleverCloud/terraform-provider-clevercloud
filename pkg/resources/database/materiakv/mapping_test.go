package materiakv_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg/resources/database/materiakv"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
	"go.clever-cloud.dev/sdk/models"
)

func TestMateriaKVFromAPI_PopulatesEveryAttribute(t *testing.T) {
	var diags diag.Diagnostics
	state := &materiakv.MateriaKV{}

	state.FromAddon(t.Context(), &tmp.AddonResponse{
		ID: "addon_1", Name: "tf-test-kv", Region: "par", CreationDate: 1759000000000,
	}, &diags)
	state.FromMateriaDB(t.Context(), &models.MateriaDB{
		Host: "host.services.clever-cloud.com", Port: 6379, Token: "secret",
	}, &diags)

	if diags.HasError() {
		t.Fatalf("mapping reported %v", diags.Errors())
	}
	for name, value := range map[string]types.String{
		"name": state.Name, "region": state.Region,
		"host": state.Host, "token": state.Token,
	} {
		if value.IsNull() || value.IsUnknown() {
			t.Errorf("%s is %v, want a value", name, value)
		}
	}
	if state.CreationDate.IsNull() || state.Port.ValueInt64() != 6379 {
		t.Error("creation_date and port must be set")
	}
}

// The Materia payload carries neither name nor region, which is why an import
// used to leave the required name attribute null (#447).
func TestMateriaKVFromAPI_PreservesWhatTheAPIDoesNotReturn(t *testing.T) {
	var diags diag.Diagnostics
	state := &materiakv.MateriaKV{
		Name:  types.StringValue("kept-name"),
		Token: types.StringValue("kept-token"),
	}

	state.FromMateriaDB(t.Context(), &models.MateriaDB{Host: "host", Port: 6379}, &diags)

	if state.Name.ValueString() != "kept-name" {
		t.Error("the Materia payload has no name and must not clear it")
	}
}

func TestMateriaKVFromAPI_NilPayloadIsANoOp(t *testing.T) {
	var diags diag.Diagnostics
	state := &materiakv.MateriaKV{Host: types.StringValue("kept-host")}

	state.FromAddon(t.Context(), nil, &diags)
	state.FromMateriaDB(t.Context(), nil, &diags)

	if diags.HasError() {
		t.Fatalf("a nil payload must not report an error, got %v", diags.Errors())
	}
	if state.Host.ValueString() != "kept-host" {
		t.Error("host must be preserved")
	}
}
