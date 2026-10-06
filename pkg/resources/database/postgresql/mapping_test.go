package postgresql

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

func apiAddon() *tmp.AddonResponse {
	return &tmp.AddonResponse{
		ID: "addon_75a69f1f-fd05-432d-aa24-33c5f03b1be9", Name: "tf-test-pg",
		RealID: "postgresql_f62a2191-aad5-4fab-bb6d-357faaf257aa", Region: "par",
		CreationDate: 1759000000000, Plan: tmp.AddonPlan{Slug: "xs_sml"},
	}
}

func apiPostgreSQL(features ...tmp.PostgreSQLFeature) *tmp.PostgreSQL {
	return &tmp.PostgreSQL{
		Host: "host.services.clever-cloud.com", Port: 5432, Database: "db",
		User: "user", Password: "secret", Version: "15", Locale: "fr_FR",
		Features: features,
	}
}

func TestPostgreSQLFromAPI_PopulatesEveryAttribute(t *testing.T) {
	var diags diag.Diagnostics
	state := &PostgreSQL{}

	state.
		FromAddon(t.Context(), apiAddon(), &diags).
		FromPostgreSQL(t.Context(), apiPostgreSQL(
			tmp.PostgreSQLFeature{Name: "do-backup", Enabled: true},
			tmp.PostgreSQLFeature{Name: "encryption", Enabled: true},
			tmp.PostgreSQLFeature{Name: "direct-host-only", Enabled: true},
		), &diags)

	if diags.HasError() {
		t.Fatalf("mapping reported %v", diags.Errors())
	}
	for name, value := range map[string]types.String{
		"name": state.Name, "plan": state.Plan, "region": state.Region,
		"host": state.Host, "database": state.Database, "user": state.User,
		"password": state.Password, "version": state.Version, "uri": state.Uri,
		"locale": state.Locale,
	} {
		if value.IsNull() || value.IsUnknown() {
			t.Errorf("%s is %v, want a value", name, value)
		}
	}
	if state.CreationDate.IsNull() || state.Port.ValueInt64() != 5432 {
		t.Error("creation_date and port must be set")
	}
	if !state.Backup.ValueBool() || !state.Encryption.ValueBool() || !state.DirectHostOnly.ValueBool() {
		t.Error("all three toggles should be true")
	}
}

func TestPostgreSQLFromAPI_PreservesWhatTheAPIDoesNotReturn(t *testing.T) {
	var diags diag.Diagnostics
	state := &PostgreSQL{
		Host:     types.StringValue("kept-host"),
		Password: types.StringValue("kept-password"),
		Locale:   types.StringValue("de_DE"),
	}

	state.FromAddon(t.Context(), apiAddon(), &diags)

	for name, got := range map[string]string{
		"host": state.Host.ValueString(), "password": state.Password.ValueString(),
	} {
		if got != "kept-"+name {
			t.Errorf("%s = %q, want it preserved", name, got)
		}
	}
	// The locale is fixed at creation and the add-on view never carries it.
	if state.Locale.ValueString() != "de_DE" {
		t.Errorf("locale = %q, want it preserved", state.Locale.ValueString())
	}
}

func TestPostgreSQLFromAPI_AbsentFeatureUsesFallback(t *testing.T) {
	var diags diag.Diagnostics
	state := &PostgreSQL{}

	state.FromPostgreSQL(t.Context(), apiPostgreSQL(), &diags)

	expected := map[string]struct {
		got  types.Bool
		want bool
	}{
		"backup":           {state.Backup, true}, // schema default
		"encryption":       {state.Encryption, false},
		"direct_host_only": {state.DirectHostOnly, false},
	}
	for name, test := range expected {
		if test.got.IsNull() || test.got.IsUnknown() {
			t.Errorf("%s is %v: a Computed attribute must never be null after a read", name, test.got)
			continue
		}
		if test.got.ValueBool() != test.want {
			t.Errorf("%s = %v, want %v", name, test.got, test.want)
		}
	}
}

func TestPostgreSQLFromAPI_NilPayloadIsANoOp(t *testing.T) {
	var diags diag.Diagnostics
	state := &PostgreSQL{Host: types.StringValue("kept-host")}

	state.
		FromAddon(t.Context(), nil, &diags).
		FromPostgreSQL(t.Context(), nil, &diags)

	if diags.HasError() {
		t.Fatalf("a nil payload must not report an error, got %v", diags.Errors())
	}
	if state.Host.ValueString() != "kept-host" {
		t.Error("host must be preserved")
	}
}
