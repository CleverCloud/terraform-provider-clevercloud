package rust

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/miton18/helper/maps"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/resources/application"
	"go.clever-cloud.dev/sdk/app"
)

// Every variable listed in envVars is written by ToEnv and read back by
// FromEnv, and the whole plan comes back: a field without its mapping fails.
func TestEnvVarsRoundTrip(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	plan := Rust{
		Runtime: application.Runtime{
			Environment: types.MapNull(types.StringType),
			AppFolder:   types.StringValue("web"),
		},
		Features:      types.SetValueMust(types.StringType, []attr.Value{types.StringValue("tls"), types.StringValue("metrics")}),
		RustupChannel: types.StringValue("nightly"),
		Bin:           types.StringValue("server"),
	}
	env := plan.ToEnv(ctx, &diags)
	for _, name := range pkg.EnvNames(envVars...) {
		if _, ok := env[name]; !ok {
			t.Errorf("ToEnv does not write %s", name)
		}
	}

	read := Rust{}
	envMap := maps.NewMap(env)
	read.FromEnv(ctx, envMap, &diags)
	if diags.HasError() {
		t.Fatal(diags)
	}
	if envMap.Size() != 0 {
		t.Errorf("FromEnv left variables behind: %v", envMap)
	}
	read.Environment = plan.Environment
	if !reflect.DeepEqual(read, plan) {
		t.Errorf("round trip lost a value:\ngot  %+v\nwant %+v", read, plan)
	}
}

func TestConfigValidators(t *testing.T) {
	ctx := context.Background()
	schemaRes := &resource.SchemaResponse{}
	(&ResourceRust{}).Schema(ctx, resource.SchemaRequest{}, schemaRes)
	typ := schemaRes.Schema.Type().TerraformType(ctx).(tftypes.Object)
	attrs := map[string]tftypes.Value{}
	for name, attrType := range typ.AttributeTypes {
		attrs[name] = tftypes.NewValue(attrType, nil)
	}
	attrs["environment"] = tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, map[string]tftypes.Value{
		app.CC_RUST_FEATURES: tftypes.NewValue(tftypes.String, "tls"),
		"MY_VARIABLE":        tftypes.NewValue(tftypes.String, "free"),
	})
	config := tfsdk.Config{Schema: schemaRes.Schema, Raw: tftypes.NewValue(typ, attrs)}

	var errs []string
	for _, v := range (&ResourceRust{}).ConfigValidators(ctx) {
		res := &resource.ValidateConfigResponse{}
		v.ValidateResource(ctx, resource.ValidateConfigRequest{Config: config}, res)
		for _, d := range res.Diagnostics.Errors() {
			errs = append(errs, d.Detail())
		}
	}
	if len(errs) != 1 || !strings.HasPrefix(errs[0], "CC_RUST_FEATURES is managed by an attribute") {
		t.Errorf("got %v, want one error about CC_RUST_FEATURES", errs)
	}
}

// An absent CC_RUST_FEATURES turns features back to null, so removing the
// variable outside Terraform shows up as drift.
func TestFeaturesAbsent(t *testing.T) {
	r := Rust{Features: types.SetValueMust(types.StringType, []attr.Value{types.StringValue("tls")})}
	FeaturesEnv.FromEnv(&r, "")
	if !r.Features.IsNull() {
		t.Errorf("features = %v, want null", r.Features)
	}
	if _, value := FeaturesEnv.ToEnv(&r); value != "" {
		t.Errorf("ToEnv of null features = %q, want nothing", value)
	}
}
