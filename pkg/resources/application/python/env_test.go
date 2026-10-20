package python

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

	plan := Python{
		Runtime: application.Runtime{
			Environment: types.MapNull(types.StringType),
			AppFolder:   types.StringValue("web"),
		},
		PythonVersion:        types.StringValue("3.13"),
		PipRequirements:      types.StringValue("requirements/prod.txt"),
		Backend:              types.StringValue("gunicorn"),
		Module:               types.StringValue("app.server:app"),
		ManageTasks:          types.StringValue("migrate,collectstatic --noinput"),
		GunicornWorkerClass:  types.StringValue("gevent"),
		UwsgiExtraConfig:     types.StringValue("harakiri = 60"),
		CeleryModule:         types.StringValue("app.tasks"),
		CeleryLogfile:        types.StringValue("/tmp/celery.log"),
		UvRunCommand:         types.StringValue("uv run app.py"),
		EnvironmentMode:      types.StringValue("staging"),
		CeleryUseBeat:        types.BoolValue(true),
		UseGevent:            types.BoolValue(true),
		UwsgiInterceptErrors: types.BoolValue(true),
	}
	env := plan.ToEnv(ctx, &diags)
	for _, name := range pkg.EnvNames(envVars...) {
		if _, ok := env[name]; !ok {
			t.Errorf("ToEnv does not write %s", name)
		}
	}

	read := Python{}
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
	typ := schemaPythonV1.Type().TerraformType(ctx).(tftypes.Object)
	attrs := map[string]tftypes.Value{}
	for name, attrType := range typ.AttributeTypes {
		attrs[name] = tftypes.NewValue(attrType, nil)
	}
	attrs["environment"] = tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, map[string]tftypes.Value{
		app.CC_PYTHON_MODULE: tftypes.NewValue(tftypes.String, "app:app"),
		"MY_VARIABLE":        tftypes.NewValue(tftypes.String, "free"),
	})
	config := tfsdk.Config{Schema: schemaPythonV1, Raw: tftypes.NewValue(typ, attrs)}

	var errs []string
	for _, v := range (&ResourcePython{}).ConfigValidators(ctx) {
		res := &resource.ValidateConfigResponse{}
		v.ValidateResource(ctx, resource.ValidateConfigRequest{Config: config}, res)
		for _, d := range res.Diagnostics.Errors() {
			errs = append(errs, d.Detail())
		}
	}
	if len(errs) != 1 || !strings.HasPrefix(errs[0], "CC_PYTHON_MODULE is managed by an attribute") {
		t.Errorf("got %v, want one error about CC_PYTHON_MODULE", errs)
	}
}
