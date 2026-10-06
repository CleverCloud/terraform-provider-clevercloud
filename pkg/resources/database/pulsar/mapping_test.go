package pulsar_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg/resources/database/pulsar"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

func TestPulsarFromAPI_PopulatesEveryAttribute(t *testing.T) {
	var diags diag.Diagnostics
	state := &pulsar.Pulsar{}

	state.FromAddon(t.Context(), &tmp.AddonResponse{Name: "tf-test-pulsar", Region: "par"}, &diags)
	state.FromPulsar(t.Context(), &tmp.Pulsar{Tenant: "tenant", Namespace: "ns", Token: "secret"}, &diags)
	state.FromCluster(t.Context(), &tmp.PulsarCluster{
		URL: "pulsar.services.clever-cloud.com", PulsarTLSPort: 6651, WebTLSPort: 8443,
	}, &diags)

	if diags.HasError() {
		t.Fatalf("mapping reported %v", diags.Errors())
	}
	for name, value := range map[string]types.String{
		"name": state.Name, "region": state.Region, "tenant": state.Tenant,
		"namespace": state.Namespace, "token": state.Token,
		"binary_url": state.BinaryURL, "http_url": state.HTTPUrl,
	} {
		if value.IsNull() || value.IsUnknown() {
			t.Errorf("%s is %v, want a value", name, value)
		}
	}
}

// The scheme follows the presence of a TLS port: the cluster answers both port
// pairs and only one of each is live.
func TestPulsarFromAPI_URLSchemeFollowsTLS(t *testing.T) {
	tests := map[string]struct {
		cluster              *tmp.PulsarCluster
		wantBinary, wantHTTP string
	}{
		"tls": {
			cluster:    &tmp.PulsarCluster{URL: "host", PulsarTLSPort: 6651, WebTLSPort: 8443},
			wantBinary: "pulsar+ssl://host:6651", wantHTTP: "https://host:8443",
		},
		"plaintext": {
			cluster:    &tmp.PulsarCluster{URL: "host", PulsarPort: 6650, WebPort: 8080},
			wantBinary: "pulsar://host:6650", wantHTTP: "http://host:8080",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			var diags diag.Diagnostics
			state := &pulsar.Pulsar{}

			state.FromCluster(t.Context(), test.cluster, &diags)

			if got := state.BinaryURL.ValueString(); got != test.wantBinary {
				t.Errorf("binary_url = %q, want %q", got, test.wantBinary)
			}
			if got := state.HTTPUrl.ValueString(); got != test.wantHTTP {
				t.Errorf("http_url = %q, want %q", got, test.wantHTTP)
			}
		})
	}
}

func TestPulsarFromAPI_PreservesWhatTheAPIDoesNotReturn(t *testing.T) {
	var diags diag.Diagnostics
	state := &pulsar.Pulsar{
		Token:           types.StringValue("kept-token"),
		RetentionSize:   types.Int64Value(1024),
		RetentionPeriod: types.Int64Value(60),
	}

	state.FromAddon(t.Context(), &tmp.AddonResponse{Name: "n", Region: "par"}, &diags)

	if state.Token.ValueString() != "kept-token" {
		t.Error("the add-on view has no token and must not clear it")
	}
	// Retention is read from the Pulsar admin API, never from a mapper.
	if state.RetentionSize.ValueInt64() != 1024 || state.RetentionPeriod.ValueInt64() != 60 {
		t.Error("retention must be left to readRetention")
	}
}

func TestPulsarFromAPI_NilPayloadIsANoOp(t *testing.T) {
	var diags diag.Diagnostics
	state := &pulsar.Pulsar{Tenant: types.StringValue("kept-tenant")}

	state.FromAddon(t.Context(), nil, &diags)
	state.FromPulsar(t.Context(), nil, &diags)
	state.FromCluster(t.Context(), nil, &diags)

	if diags.HasError() {
		t.Fatalf("a nil payload must not report an error, got %v", diags.Errors())
	}
	if state.Tenant.ValueString() != "kept-tenant" {
		t.Error("tenant must be preserved")
	}
}
