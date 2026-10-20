package scala

import (
	"context"
	"reflect"
	"strings"
	"testing"

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

	plan := Scala{
		Runtime: application.Runtime{
			Environment: types.MapNull(types.StringType),
			AppFolder:   types.StringValue("web"),
		},
		JavaVersion:         types.StringValue("21"),
		SbtBuildGoal:        types.StringValue("clean stage"),
		SbtDeployArgs:       types.StringValue("-Dhttp.port=8080"),
		SbtTargetBin:        types.StringValue("myapp"),
		SbtTargetDir:        types.StringValue("server"),
		DisableMaxMetaspace: types.BoolValue(true),
		APMAgentEnabled:     types.BoolValue(true),
		APMAgentContinuous:  types.BoolValue(true),
	}
	env := plan.ToEnv(ctx, &diags)
	for _, name := range pkg.EnvNames(envVars...) {
		if _, ok := env[name]; !ok {
			t.Errorf("ToEnv does not write %s", name)
		}
	}

	read := Scala{}
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
	typ := schemaScala.Type().TerraformType(ctx).(tftypes.Object)
	attrs := map[string]tftypes.Value{}
	for name, attrType := range typ.AttributeTypes {
		attrs[name] = tftypes.NewValue(attrType, nil)
	}
	attrs["environment"] = tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, map[string]tftypes.Value{
		app.CC_SBT_TARGET_BIN: tftypes.NewValue(tftypes.String, "myapp"),
		"MY_VARIABLE":         tftypes.NewValue(tftypes.String, "free"),
	})
	config := tfsdk.Config{Schema: schemaScala, Raw: tftypes.NewValue(typ, attrs)}

	var errs []string
	for _, v := range (&ResourceScala{}).ConfigValidators(ctx) {
		res := &resource.ValidateConfigResponse{}
		v.ValidateResource(ctx, resource.ValidateConfigRequest{Config: config}, res)
		for _, d := range res.Diagnostics.Errors() {
			errs = append(errs, d.Detail())
		}
	}
	if len(errs) != 1 || !strings.HasPrefix(errs[0], "CC_SBT_TARGET_BIN is managed by an attribute") {
		t.Errorf("got %v, want one error about CC_SBT_TARGET_BIN", errs)
	}
}
