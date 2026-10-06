package keycloak_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg/resources/software/keycloak"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
	"go.clever-cloud.dev/sdk/models"
)

func strPtr(s string) *string { return &s }

func apiKeycloak() *models.Keycloak {
	api := &models.Keycloak{
		Name:      "tf-test-keycloak",
		AccessURL: "https://keycloak.example.com",
		Version:   "26.0.0",
		EnvVars:   models.MapString{"CC_KEYCLOAK_HOSTNAME": "keycloak.example.com"},
		Resources: models.KeycloakResources{FsbucketID: strPtr("fsbucket_1")},
	}
	api.InitialCredentials.User = "admin"
	api.InitialCredentials.Password = "secret"
	return api
}

func TestKeycloakFromAPI_PopulatesEveryAttribute(t *testing.T) {
	var diags diag.Diagnostics
	state := &keycloak.Keycloak{}

	state.FromAddon(t.Context(), &tmp.AddonResponse{Name: "tf-test-keycloak", Region: "par"}, &diags)
	state.FromKeycloak(t.Context(), apiKeycloak(), &diags)

	if diags.HasError() {
		t.Fatalf("mapping reported %v", diags.Errors())
	}
	for name, value := range map[string]types.String{
		"name": state.Name, "region": state.Region, "host": state.Host,
		"admin_username": state.AdminUsername, "admin_password": state.AdminPassword,
		"version": state.Version, "access_domain": state.AccessDomain,
		"fsbucket_id": state.FSBucketID,
	} {
		if value.IsNull() || value.IsUnknown() {
			t.Errorf("%s is %v, want a value", name, value)
		}
	}
	// access_domain lives in the payload's own env map, not in a field.
	if got := state.AccessDomain.ValueString(); got != "keycloak.example.com" {
		t.Errorf("access_domain = %q", got)
	}
}

func TestKeycloakFromAPI_PreservesWhatTheAPIDoesNotReturn(t *testing.T) {
	var diags diag.Diagnostics
	state := &keycloak.Keycloak{AdminPassword: types.StringValue("kept-secret")}

	state.FromAddon(t.Context(), &tmp.AddonResponse{Name: "n", Region: "par"}, &diags)

	if state.AdminPassword.ValueString() != "kept-secret" {
		t.Error("the add-on view carries no credential and must not clear one")
	}
}

func TestKeycloakFromAPI_NilPayloadIsANoOp(t *testing.T) {
	var diags diag.Diagnostics
	state := &keycloak.Keycloak{Host: types.StringValue("kept-host")}

	state.FromAddon(t.Context(), nil, &diags)
	state.FromKeycloak(t.Context(), nil, &diags)

	if diags.HasError() {
		t.Fatalf("a nil payload must not report an error, got %v", diags.Errors())
	}
	if state.Host.ValueString() != "kept-host" {
		t.Error("host must be preserved")
	}
}
