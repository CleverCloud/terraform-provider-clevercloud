package mongodb_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg/resources/database/mongodb"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

func apiAddon() *tmp.AddonResponse {
	return &tmp.AddonResponse{
		ID: "addon_6fad3214-cfb5-4178-a633-608bd831d18b", Name: "tf-test-mongodb",
		RealID: "mongodb_0ec96802-80b7-4c38-a95e-1c8da9b5cfcf", Region: "par",
		CreationDate: 1759000000000, Plan: tmp.AddonPlan{Slug: "s_sml"},
	}
}

func apiMongoDB(features ...tmp.MongoDBFeature) *tmp.MongoDB {
	return &tmp.MongoDB{
		Host: "host.services.clever-cloud.com", Port: 27017,
		User: "user", Password: "secret", Database: "db", Features: features,
	}
}

func TestMongoDBFromAPI_PopulatesEveryAttribute(t *testing.T) {
	var diags diag.Diagnostics
	state := &mongodb.MongoDB{}

	state.FromAddon(t.Context(), apiAddon(), &diags)
	state.FromMongoDB(t.Context(), apiMongoDB(
		tmp.MongoDBFeature{Name: "encryption", Enabled: true},
		tmp.MongoDBFeature{Name: "direct-host-only", Enabled: true},
	), &diags)

	if diags.HasError() {
		t.Fatalf("mapping reported %v", diags.Errors())
	}
	for name, value := range map[string]types.String{
		"name": state.Name, "plan": state.Plan, "region": state.Region,
		"host": state.Host, "user": state.User, "password": state.Password,
		"database": state.Database, "uri": state.Uri,
	} {
		if value.IsNull() || value.IsUnknown() {
			t.Errorf("%s is %v, want a value", name, value)
		}
	}
	if state.Port.ValueInt64() != 27017 {
		t.Errorf("port = %d", state.Port.ValueInt64())
	}
	if !state.Encryption.ValueBool() || !state.DirectHostOnly.ValueBool() {
		t.Error("both toggles should be true")
	}
}

func TestMongoDBFromAPI_PreservesWhatTheAPIDoesNotReturn(t *testing.T) {
	var diags diag.Diagnostics
	state := &mongodb.MongoDB{
		Host:     types.StringValue("kept-host"),
		Password: types.StringValue("kept-password"),
	}

	state.FromAddon(t.Context(), apiAddon(), &diags)

	if state.Host.ValueString() != "kept-host" || state.Password.ValueString() != "kept-password" {
		t.Error("the add-on view carries no connection detail and must not clear one")
	}
}

func TestMongoDBFromAPI_AbsentFeatureUsesFallback(t *testing.T) {
	var diags diag.Diagnostics
	state := &mongodb.MongoDB{}

	state.FromMongoDB(t.Context(), apiMongoDB(), &diags)

	for name, toggle := range map[string]types.Bool{
		"encryption": state.Encryption, "direct_host_only": state.DirectHostOnly,
	} {
		if toggle.IsNull() || toggle.IsUnknown() {
			t.Errorf("%s is %v: a Computed attribute must never be null after a read", name, toggle)
			continue
		}
		if toggle.ValueBool() {
			t.Errorf("%s = true, want the false fallback", name)
		}
	}
}

func TestMongoDBFromAPI_NilPayloadIsANoOp(t *testing.T) {
	var diags diag.Diagnostics
	state := &mongodb.MongoDB{Host: types.StringValue("kept-host")}

	state.FromAddon(t.Context(), nil, &diags)
	state.FromMongoDB(t.Context(), nil, &diags)

	if diags.HasError() {
		t.Fatalf("a nil payload must not report an error, got %v", diags.Errors())
	}
	if state.Host.ValueString() != "kept-host" {
		t.Error("host must be preserved")
	}
}
