package application

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/miton18/helper/maps"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.dev/sdk/app"
)

func validateEnvironment(t *testing.T, env tftypes.Value) []string {
	t.Helper()
	ctx := context.Background()
	s := schema.Schema{Attributes: map[string]schema.Attribute{
		"environment": schema.MapAttribute{Optional: true, ElementType: types.StringType},
	}}
	config := tfsdk.Config{
		Schema: s,
		Raw: tftypes.NewValue(s.Type().TerraformType(ctx), map[string]tftypes.Value{
			"environment": env,
		}),
	}

	res := &resource.ValidateConfigResponse{}
	ReservedEnvironment(AppFolderEnv).ValidateResource(ctx, resource.ValidateConfigRequest{Config: config}, res)

	var details []string
	for _, d := range res.Diagnostics.Errors() {
		details = append(details, d.Detail())
	}
	return details
}

func TestReservedEnvironment(t *testing.T) {
	errs := validateEnvironment(t, tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, map[string]tftypes.Value{
		app.APP_FOLDER:        tftypes.NewValue(tftypes.String, "web"),
		app.CC_CLAMAV:         tftypes.NewValue(tftypes.String, "true"),
		app.CC_PRE_BUILD_HOOK: tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		"MY_VARIABLE":         tftypes.NewValue(tftypes.String, "free"),
	}))

	// in list order: the resource mappings, then the hooks, then the integrations
	want := []string{app.APP_FOLDER, app.CC_PRE_BUILD_HOOK, app.CC_CLAMAV}
	if len(errs) != len(want) {
		t.Fatalf("got %d errors %v, want %d", len(errs), errs, len(want))
	}
	for i, name := range want {
		if !strings.HasPrefix(errs[i], name+" is managed by an attribute") {
			t.Errorf("error %d = %q, want it about %s", i, errs[i], name)
		}
	}
}

func TestReservedEnvironment_nullOrUnknown(t *testing.T) {
	for name, env := range map[string]tftypes.Value{
		"null":    tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, nil),
		"unknown": tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, tftypes.UnknownValue),
	} {
		if errs := validateEnvironment(t, env); len(errs) != 0 {
			t.Errorf("%s environment: got errors %v", name, errs)
		}
	}
}

// testPlan is the smallest RuntimePlan: a Runtime and no variable of its own.
type testPlan struct{ Runtime }

func (p *testPlan) ToEnv(context.Context, *diag.Diagnostics) map[string]string { return nil }
func (p *testPlan) FromEnv(context.Context, *maps.Map[string, string], *diag.Diagnostics) {
}

func TestOnRuntime(t *testing.T) {
	m := OnRuntime[*testPlan](AppFolderEnv)

	p := &testPlan{}
	p.AppFolder = types.StringValue("web")
	if name, value := m.ToEnv(p); name != app.APP_FOLDER || value != "web" {
		t.Errorf("ToEnv = %s, %s", name, value)
	}

	env := maps.NewMap(map[string]string{app.APP_FOLDER: "api"})
	pkg.ReadEnv(env, p, m)
	if p.AppFolder.ValueString() != "api" || env.Has(app.APP_FOLDER) {
		t.Errorf("ReadEnv: app_folder = %s, env = %v", p.AppFolder, env)
	}
}
