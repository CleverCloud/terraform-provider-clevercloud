package addon

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/provider"
	"go.clever-cloud.dev/client"
	"go.clever-cloud.dev/sdk"
	"go.clever-cloud.dev/sdk/models"
)

type createTestProvider struct {
	provider.Provider
	client *client.Client
}

func (p createTestProvider) Organization() string   { return "orga" }
func (p createTestProvider) Client() *client.Client { return p.client }

// The add-on the API accepted stays in state when reading its environment
// fails afterwards: it exists, and without it the next apply creates a second
// one and destroy leaves the first behind.
func TestCreateSavesAnAddonWhoseEnvironmentCannotBeRead(t *testing.T) {
	ctx := t.Context()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v2/products/addonproviders":
			_ = json.NewEncoder(w).Encode([]models.AddonProviderInfoFullView{{
				ID:    "mailpace",
				Plans: []models.AddonPlanView{{ID: "plan_1", Slug: pkg.Ptr("s")}},
			}})
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/addons"):
			_ = json.NewEncoder(w).Encode(models.AddonView{
				ID:     "addon_1",
				RealID: pkg.Ptr("mailpace_1"),
				Name:   pkg.Ptr("tf-test"),
				Region: "par",
			})
		default:
			// the environment of the add-on, among others
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))
	t.Cleanup(server.Close)
	cc := client.New(client.WithEndpoint(server.URL))

	res := &ResourceAddon{}
	res.Provider = createTestProvider{client: cc}
	res.SDK = sdk.NewSDK(sdk.WithClient(cc))

	var s resource.SchemaResponse
	res.Schema(ctx, resource.SchemaRequest{}, &s)
	typ := s.Schema.Type().TerraformType(ctx).(tftypes.Object)

	// what Terraform plans for a new add-on: the configuration, and unknown
	// for what the provider computes
	attrs := map[string]tftypes.Value{}
	for name, attrType := range typ.AttributeTypes {
		attrs[name] = tftypes.NewValue(attrType, nil)
	}
	for name, value := range map[string]string{"name": "tf-test", "region": "par", "third_party_provider": "mailpace", "plan": "s"} {
		attrs[name] = tftypes.NewValue(tftypes.String, value)
	}
	for _, name := range []string{"id", "creation_date", "configurations"} {
		attrs[name] = tftypes.NewValue(typ.AttributeTypes[name], tftypes.UnknownValue)
	}

	resp := resource.CreateResponse{State: tfsdk.State{Schema: s.Schema, Raw: tftypes.NewValue(typ, nil)}}
	res.Create(ctx, resource.CreateRequest{Plan: tfsdk.Plan{Schema: s.Schema, Raw: tftypes.NewValue(typ, attrs)}}, &resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("Create reports no error although the environment could not be read")
	}
	if resp.State.Raw.IsNull() {
		t.Fatal("the created add-on is not in state")
	}
	var id types.String
	if d := resp.State.GetAttribute(ctx, path.Root("id"), &id); d.HasError() {
		t.Fatalf("get id: %v", d)
	}
	if id.ValueString() != "addon_1" {
		t.Errorf("id = %q, want addon_1", id.ValueString())
	}
}
