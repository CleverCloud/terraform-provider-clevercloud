package mysql_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg/resources/database/mysql"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

func apiMySQL(features ...tmp.MySQLFeature) *tmp.MySQL {
	return &tmp.MySQL{
		Host:     "bwf32ifhr5cofspgzrbb-mysql.services.clever-cloud.com",
		Port:     3306,
		Database: "bwf32ifhr5cofspgzrbb",
		User:     "uxw1ikwnp6gflbgp5iun",
		Password: "omEbGQw628gIxHK9Bp8d",
		Version:  "8.0",
		Features: features,
		ReadOnlyUsers: []tmp.MySQLReadOnlyUser{
			{User: "ro-user", Password: "ro-password"},
		},
	}
}

func apiAddon() *tmp.AddonResponse {
	return &tmp.AddonResponse{
		ID:           "addon_a32e4d4c-9987-4fab-9ea8-eceae7527134",
		Name:         "tf-test-mysql",
		RealID:       "mysql_c48094c0-51a0-4aa7-8b9a-54fbbc58b1b2",
		Region:       "par",
		CreationDate: 1759000000000,
		Plan:         tmp.AddonPlan{Slug: "xs_sml"},
	}
}

// Nothing the API returns may be left null: Read is the only import path, so a
// forgotten attribute stays empty and the next plan is never empty.
func TestMySQLFromAPI_PopulatesEveryAttribute(t *testing.T) {
	var diags diag.Diagnostics
	state := &mysql.MySQL{}

	state.FromAddon(t.Context(), apiAddon(), &diags)
	state.FromMySQL(t.Context(), apiMySQL(
		tmp.MySQLFeature{Name: "do-backup", Enabled: true},
		tmp.MySQLFeature{Name: "encryption", Enabled: true},
		tmp.MySQLFeature{Name: "direct-host-only", Enabled: true},
		tmp.MySQLFeature{Name: "skip-log-bin", Enabled: true},
	), &diags)

	if diags.HasError() {
		t.Fatalf("mapping reported %v", diags.Errors())
	}

	attributes := map[string]types.String{
		"name":     state.Name,
		"plan":     state.Plan,
		"region":   state.Region,
		"host":     state.Host,
		"database": state.Database,
		"user":     state.User,
		"password": state.Password,
		"version":  state.Version,
		"uri":      state.Uri,
	}
	for name, value := range attributes {
		if value.IsNull() || value.IsUnknown() {
			t.Errorf("%s is %v, want a value", name, value)
		}
	}
	if state.CreationDate.IsNull() || state.Port.IsNull() {
		t.Error("creation_date and port must be set")
	}
	if state.ReadOnlyUsers.IsNull() {
		t.Error("read_only_users must be set")
	}

	if got := state.Name.ValueString(); got != "tf-test-mysql" {
		t.Errorf("name = %q", got)
	}
	if got := state.Plan.ValueString(); got != "xs_sml" {
		t.Errorf("plan = %q, want the slug and not the id", got)
	}
	if got := state.Port.ValueInt64(); got != 3306 {
		t.Errorf("port = %d", got)
	}
	if got := state.Uri.ValueString(); got == "" {
		t.Error("uri is computed from the other fields and must not be empty")
	}
	for name, toggle := range map[string]types.Bool{
		"backup": state.Backup, "encryption": state.Encryption,
		"direct_host_only": state.DirectHostOnly, "skip_log_bin": state.SkipLogBin,
	} {
		if !toggle.ValueBool() {
			t.Errorf("%s = %v, want true", name, toggle)
		}
	}
}

// The API is not the authority on a value it does not send. The add-on view
// carries no connection detail, so mapping it must leave those alone.
func TestMySQLFromAPI_PreservesWhatTheAPIDoesNotReturn(t *testing.T) {
	var diags diag.Diagnostics
	state := &mysql.MySQL{
		Host:     types.StringValue("kept-host"),
		Password: types.StringValue("kept-password"),
		Uri:      types.StringValue("kept-uri"),
	}

	state.FromAddon(t.Context(), apiAddon(), &diags)

	if diags.HasError() {
		t.Fatalf("mapping reported %v", diags.Errors())
	}
	for name, got := range map[string]string{
		"host": state.Host.ValueString(), "password": state.Password.ValueString(),
		"uri": state.Uri.ValueString(),
	} {
		if got != "kept-"+name {
			t.Errorf("%s = %q, want it preserved", name, got)
		}
	}
}

// An absent toggle means "the API holds no opinion", so it has to resolve to
// the schema default — never to null, which is what #404 was.
func TestMySQLFromAPI_AbsentFeatureUsesFallback(t *testing.T) {
	var diags diag.Diagnostics
	state := &mysql.MySQL{}

	state.FromMySQL(t.Context(), apiMySQL(), &diags)

	if diags.HasError() {
		t.Fatalf("mapping reported %v", diags.Errors())
	}

	// backup defaults to true in the schema, the other three to false.
	expected := map[string]struct {
		got  types.Bool
		want bool
	}{
		"backup":           {state.Backup, true},
		"encryption":       {state.Encryption, false},
		"direct_host_only": {state.DirectHostOnly, false},
		"skip_log_bin":     {state.SkipLogBin, false},
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

// A failed fetch hands the mapper no payload, and prior state must survive it.
func TestMySQLFromAPI_NilPayloadIsANoOp(t *testing.T) {
	var diags diag.Diagnostics
	state := &mysql.MySQL{Host: types.StringValue("kept-host")}

	state.FromAddon(t.Context(), nil, &diags)
	state.FromMySQL(t.Context(), nil, &diags)

	if diags.HasError() {
		t.Fatalf("a nil payload must not report an error, got %v", diags.Errors())
	}
	if got := state.Host.ValueString(); got != "kept-host" {
		t.Errorf("host = %q, want it preserved", got)
	}
}
