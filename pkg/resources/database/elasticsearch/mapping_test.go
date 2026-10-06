package elasticsearch_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg/resources/database/elasticsearch"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

func strPtr(s string) *string { return &s }

func apiAddon() *tmp.AddonResponse {
	return &tmp.AddonResponse{
		ID: "addon_af73db11-c879-406c-92da-7a830e54909c", Name: "tf-test-es",
		RealID: "elasticsearch_f3cbc917-032e-4536-a22f-db3d41deaf85", Region: "par",
		Plan: tmp.AddonPlan{Slug: "xs"},
	}
}

func apiElasticsearch(features ...tmp.ElasticsearchFeature) *tmp.Elasticsearch {
	return &tmp.Elasticsearch{
		Host: "host.services.clever-cloud.com", User: "user", Password: "secret",
		Version:        "8.11.3",
		KibanaHost:     strPtr("kibana.services.clever-cloud.com"),
		KibanaUser:     "kibana-user",
		KibanaPassword: "kibana-secret",
		ApmHost:        strPtr("apm.services.clever-cloud.com"),
		ApmUser:        "apm-user",
		ApmPassword:    "apm-secret",
		ApmAuthToken:   "apm-token",
		Plugins:        []string{"analysis-icu"},
		Features:       features,
	}
}

func TestElasticsearchFromAPI_PopulatesEveryAttribute(t *testing.T) {
	var diags diag.Diagnostics
	state := &elasticsearch.Elasticsearch{}

	state.
		FromAddon(t.Context(), apiAddon(), &diags).
		FromElasticsearch(t.Context(), apiElasticsearch(
			tmp.ElasticsearchFeature{Name: "encryption", Enabled: true},
			tmp.ElasticsearchFeature{Name: "kibana", Enabled: true},
			tmp.ElasticsearchFeature{Name: "apm", Enabled: true},
		), &diags)

	if diags.HasError() {
		t.Fatalf("mapping reported %v", diags.Errors())
	}
	for name, value := range map[string]types.String{
		"name": state.Name, "plan": state.Plan, "region": state.Region,
		"host": state.Host, "user": state.User, "password": state.Password,
		"kibana_host": state.KibanaHost, "kibana_user": state.KibanaUser,
		"kibana_password": state.KibanaPassword, "apm_host": state.ApmHost,
		"apm_user": state.ApmUser, "apm_password": state.ApmPassword,
		"apm_token": state.ApmToken,
	} {
		if value.IsNull() || value.IsUnknown() {
			t.Errorf("%s is %v, want a value", name, value)
		}
	}
	// The schema holds the major version alone.
	if got := state.Version.ValueString(); got != "8" {
		t.Errorf("version = %q, want the major alone", got)
	}
	if state.Plugins.IsNull() {
		t.Error("plugins must be set when the API reports some")
	}
	if !state.Encryption.ValueBool() || !state.Kibana.ValueBool() || !state.Apm.ValueBool() {
		t.Error("all three toggles should be true")
	}
}

// A disabled service has no credentials, and they must be cleared rather than
// kept from a previous read.
func TestElasticsearchFromAPI_ClearsCredentialsOfDisabledServices(t *testing.T) {
	var diags diag.Diagnostics
	state := &elasticsearch.Elasticsearch{
		KibanaUser:  types.StringValue("stale"),
		ApmUser:     types.StringValue("stale"),
		ApmPassword: types.StringValue("stale"),
		ApmToken:    types.StringValue("stale"),
	}

	state.FromElasticsearch(t.Context(), apiElasticsearch(), &diags)

	for name, value := range map[string]types.String{
		"kibana_user": state.KibanaUser, "apm_user": state.ApmUser,
		"apm_password": state.ApmPassword, "apm_token": state.ApmToken,
	} {
		if !value.IsNull() {
			t.Errorf("%s = %v, want null while the service is off", name, value)
		}
	}
}

// apm enabled with no host yet is a state the API does report while the service
// boots, and it used to panic the provider on a nil dereference.
func TestElasticsearchFromAPI_ApmEnabledWithoutHost(t *testing.T) {
	var diags diag.Diagnostics
	state := &elasticsearch.Elasticsearch{}
	api := apiElasticsearch(tmp.ElasticsearchFeature{Name: "apm", Enabled: true})
	api.ApmHost = nil

	state.FromElasticsearch(t.Context(), api, &diags)

	if diags.HasError() {
		t.Fatalf("mapping reported %v", diags.Errors())
	}
	if !state.ApmHost.IsNull() {
		t.Errorf("apm_host = %v, want null", state.ApmHost)
	}
	if state.ApmUser.IsNull() {
		t.Error("the other apm credentials are still returned and must be mapped")
	}
}

func TestElasticsearchFromAPI_PreservesWhatTheAPIDoesNotReturn(t *testing.T) {
	var diags diag.Diagnostics
	state := &elasticsearch.Elasticsearch{
		Host:     types.StringValue("kept-host"),
		Password: types.StringValue("kept-password"),
	}

	state.FromAddon(t.Context(), apiAddon(), &diags)

	for name, got := range map[string]string{
		"host": state.Host.ValueString(), "password": state.Password.ValueString(),
	} {
		if got != "kept-"+name {
			t.Errorf("%s = %q, want it preserved", name, got)
		}
	}
}

// kibana and apm are Computed with a false default AND RequiresReplace: a null
// here is a destroy-and-recreate on the next plan (#452).
func TestElasticsearchFromAPI_AbsentFeatureUsesFallback(t *testing.T) {
	var diags diag.Diagnostics
	state := &elasticsearch.Elasticsearch{}

	state.FromElasticsearch(t.Context(), apiElasticsearch(), &diags)

	for name, toggle := range map[string]types.Bool{
		"encryption": state.Encryption, "kibana": state.Kibana, "apm": state.Apm,
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

func TestElasticsearchFromAPI_NilPayloadIsANoOp(t *testing.T) {
	var diags diag.Diagnostics
	state := &elasticsearch.Elasticsearch{Host: types.StringValue("kept-host")}

	state.
		FromAddon(t.Context(), nil, &diags).
		FromElasticsearch(t.Context(), nil, &diags)

	if diags.HasError() {
		t.Fatalf("a nil payload must not report an error, got %v", diags.Errors())
	}
	if state.Host.ValueString() != "kept-host" {
		t.Error("host must be preserved")
	}
}
