package helper

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"go.clever-cloud.com/terraform-provider/pkg/provider"
	"go.clever-cloud.dev/client"
)

type addonTestProvider struct {
	provider.Provider
	client *client.Client
}

func (p addonTestProvider) Organization() string   { return "orga" }
func (p addonTestProvider) Client() *client.Client { return p.client }

var addonTestSchema = schema.Schema{Attributes: map[string]schema.Attribute{"id": schema.StringAttribute{Computed: true}}}

func addonTestState(ctx context.Context) tfsdk.State {
	return tfsdk.State{Schema: addonTestSchema, Raw: tftypes.NewValue(addonTestSchema.Type().TerraformType(ctx), nil)}
}

// addonListClient answers the add-on listing, or the given status when not zero
func addonListClient(t *testing.T, status int, calls *int) *client.Client {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*calls++
		if r.URL.Path != "/v2/organisations/orga/addons" {
			t.Errorf("unexpected request: %s", r.URL.Path)
		}
		if status != 0 {
			w.WriteHeader(status)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]map[string]string{{"id": "addon_1", "realId": "redis_1"}})
	}))
	t.Cleanup(server.Close)

	return client.New(client.WithEndpoint(server.URL))
}

func importedID(t *testing.T, ctx context.Context, state tfsdk.State) types.String {
	t.Helper()

	var id types.String
	if d := state.GetAttribute(ctx, path.Root("id"), &id); d.HasError() {
		t.Fatalf("get id: %v", d)
	}

	return id
}

func TestAddonConfigurerImportState(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		status  int
		noProv  bool
		wantErr string
		calls   int
	}{
		{name: "a real ID goes through without any API call", id: "redis_1"},
		{name: "an add-on ID is refused and the error gives the real ID", id: "addon_1", wantErr: "real ID instead: redis_1", calls: 1},
		{name: "the add-on does not exist", id: "addon_2", wantErr: "clever-tools terraform generate", calls: 1},
		{name: "the API does not answer", id: "addon_1", status: http.StatusInternalServerError, wantErr: "clever-tools terraform generate", calls: 1},
		{name: "the provider is not configured", id: "addon_1", noProv: true, wantErr: "clever-tools terraform generate"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			calls := 0
			c := &AddonConfigurer{}
			if !tt.noProv {
				c.Provider = addonTestProvider{client: addonListClient(t, tt.status, &calls)}
			}

			resp := resource.ImportStateResponse{State: addonTestState(ctx)}
			c.ImportState(ctx, resource.ImportStateRequest{ID: tt.id}, &resp)

			if calls != tt.calls {
				t.Errorf("got %d API calls, want %d", calls, tt.calls)
			}
			if tt.wantErr == "" {
				if resp.Diagnostics.HasError() {
					t.Fatalf("unexpected diagnostics: %v", resp.Diagnostics)
				}
				if got := importedID(t, ctx, resp.State); got.ValueString() != tt.id {
					t.Errorf("imported %s, want %s", got, tt.id)
				}
				return
			}

			if resp.Diagnostics.ErrorsCount() != 1 {
				t.Fatalf("got %d errors, want 1: %v", resp.Diagnostics.ErrorsCount(), resp.Diagnostics)
			}
			if detail := resp.Diagnostics.Errors()[0].Detail(); !strings.Contains(detail, tt.wantErr) {
				t.Errorf("error %q, want it to contain %q", detail, tt.wantErr)
			}
			if !resp.State.Raw.IsNull() {
				t.Error("nothing should be imported")
			}
		})
	}
}

// TestConfigurerImportStateKeepsAddonIDs pins the generic clevercloud_addon: its
// state holds the add-on ID, and clever-tools generates its import with that ID
func TestConfigurerImportStateKeepsAddonIDs(t *testing.T) {
	ctx := t.Context()
	resp := resource.ImportStateResponse{State: addonTestState(ctx)}
	(&Configurer{}).ImportState(ctx, resource.ImportStateRequest{ID: "addon_1"}, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %v", resp.Diagnostics)
	}
	if got := importedID(t, ctx, resp.State); got.ValueString() != "addon_1" {
		t.Errorf("imported %s", got)
	}
}
