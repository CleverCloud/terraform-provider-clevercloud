package otoroshi_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg/resources/software/otoroshi"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

func apiOtoroshi() *tmp.OtoroshiInfo {
	api := &tmp.OtoroshiInfo{
		AccessURL: "https://otoroshi.example.com",
		API: &tmp.OtoroshiAPI{
			URL: "https://otoroshi-api.example.com", User: "client-id", Secret: "client-secret",
		},
	}
	api.Initialredentials.User = "admin"
	api.Initialredentials.Passsword = "secret"
	return api
}

func TestOtoroshiFromAPI_PopulatesEveryAttribute(t *testing.T) {
	var diags diag.Diagnostics
	state := &otoroshi.Otoroshi{}

	state.
		FromAddon(t.Context(), &tmp.AddonResponse{
			Name: "tf-test-otoroshi", Region: "par", CreationDate: 1759000000000,
		}, &diags).
		FromOtoroshi(t.Context(), apiOtoroshi(), &diags)

	if diags.HasError() {
		t.Fatalf("mapping reported %v", diags.Errors())
	}
	for name, value := range map[string]types.String{
		"name": state.Name, "region": state.Region, "url": state.URL,
		"api_url": state.APIURL, "api_client_id": state.APIClientID,
		"api_client_secret":      state.APIClientSecret,
		"initial_admin_login":    state.InitialAdminLogin,
		"initial_admin_password": state.InitialAdminPassword,
	} {
		if value.IsNull() || value.IsUnknown() {
			t.Errorf("%s is %v, want a value", name, value)
		}
	}
	if state.CreationDate.IsNull() {
		t.Error("creation_date must be set")
	}
}

// The API client only exists once the API is reachable, and a payload without
// one must leave whatever state holds rather than clearing it.
func TestOtoroshiFromAPI_PreservesWhatTheAPIDoesNotReturn(t *testing.T) {
	var diags diag.Diagnostics
	state := &otoroshi.Otoroshi{
		APIClientSecret: types.StringValue("kept-secret"),
		APIURL:          types.StringValue("kept-url"),
	}
	api := apiOtoroshi()
	api.API = nil

	state.FromOtoroshi(t.Context(), api, &diags)

	if state.APIClientSecret.ValueString() != "kept-secret" || state.APIURL.ValueString() != "kept-url" {
		t.Error("the api client must be preserved while the API is not reachable yet")
	}
}

func TestOtoroshiFromAPI_NilPayloadIsANoOp(t *testing.T) {
	var diags diag.Diagnostics
	state := &otoroshi.Otoroshi{URL: types.StringValue("kept-url")}

	state.
		FromAddon(t.Context(), nil, &diags).
		FromOtoroshi(t.Context(), nil, &diags)

	if diags.HasError() {
		t.Fatalf("a nil payload must not report an error, got %v", diags.Errors())
	}
	if state.URL.ValueString() != "kept-url" {
		t.Error("url must be preserved")
	}
}
