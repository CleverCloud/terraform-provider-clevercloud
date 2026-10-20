package docker

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

	plan := Docker{
		Runtime: application.Runtime{
			Environment: types.MapNull(types.StringType),
			AppFolder:   types.StringValue("web"),
		},
		Dockerfile:        types.StringValue("Dockerfile.prod"),
		ContainerPort:     types.Int64Value(3000),
		ContainerPortTCP:  types.Int64Value(5432),
		IPv6Cidr:          types.StringValue("fd00::/64"),
		RegistryURL:       types.StringValue("registry.example.com"),
		RegistryUser:      types.StringValue("user"),
		RegistryPassword:  types.StringValue("secret"),
		DaemonSocketMount: types.BoolValue(true),
		Buildx:            types.BoolValue(false),
	}
	env := plan.ToEnv(ctx, &diags)
	for _, name := range pkg.EnvNames(envVars...) {
		if _, ok := env[name]; !ok {
			t.Errorf("ToEnv does not write %s", name)
		}
	}

	read := Docker{}
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
	typ := schemaDocker.Type().TerraformType(ctx).(tftypes.Object)
	attrs := map[string]tftypes.Value{}
	for name, attrType := range typ.AttributeTypes {
		attrs[name] = tftypes.NewValue(attrType, nil)
	}
	attrs["environment"] = tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, map[string]tftypes.Value{
		app.CC_DOCKER_EXPOSED_HTTP_PORT: tftypes.NewValue(tftypes.String, "3000"),
		"MY_VARIABLE":                   tftypes.NewValue(tftypes.String, "free"),
	})
	config := tfsdk.Config{Schema: schemaDocker, Raw: tftypes.NewValue(typ, attrs)}

	var errs []string
	for _, v := range (&ResourceDocker{}).ConfigValidators(ctx) {
		res := &resource.ValidateConfigResponse{}
		v.ValidateResource(ctx, resource.ValidateConfigRequest{Config: config}, res)
		for _, d := range res.Diagnostics.Errors() {
			errs = append(errs, d.Detail())
		}
	}
	if len(errs) != 1 || !strings.HasPrefix(errs[0], "CC_DOCKER_EXPOSED_HTTP_PORT is managed by an attribute") {
		t.Errorf("got %v, want one error about CC_DOCKER_EXPOSED_HTTP_PORT", errs)
	}
}
