package php

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

	plan := PHP{
		Runtime: application.Runtime{
			Environment: types.MapNull(types.StringType),
			AppFolder:   types.StringValue("web"),
		},
		PHPVersion:                   types.StringValue("8.4"),
		WebRoot:                      types.StringValue("/public"),
		RedisSessions:                types.BoolValue(true),
		DevDependencies:              types.BoolValue(true),
		ComposerVersion:              types.StringValue("2"),
		CGIImplementation:            types.StringValue("fastcgi"),
		ApacheHeadersSize:            types.Int64Value(16),
		HTTPTimeout:                  types.Int64Value(300),
		MaxInputVars:                 types.Int64Value(5000),
		MemoryLimit:                  types.Int64Value(256),
		RealpathCacheTTL:             types.Int64Value(600),
		OpcacheMemory:                types.Int64Value(128),
		OpcacheMaxAcceleratedFiles:   types.Int64Value(20000),
		OpcacheInternedStringsBuffer: types.Int64Value(16),
		OpcachePreload:               types.StringValue("config/preload.php"),
		AsyncAppBucket:               types.BoolValue(true),
		DisableAppBucket:             types.BoolValue(true),
		APMAgentEnabled:              types.BoolValue(false),
	}
	env := plan.ToEnv(ctx, &diags)
	for _, name := range pkg.EnvNames(envVars...) {
		if _, ok := env[name]; !ok {
			t.Errorf("ToEnv does not write %s", name)
		}
	}

	read := PHP{}
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
	typ := schemaPHP.Type().TerraformType(ctx).(tftypes.Object)
	attrs := map[string]tftypes.Value{}
	for name, attrType := range typ.AttributeTypes {
		attrs[name] = tftypes.NewValue(attrType, nil)
	}
	attrs["environment"] = tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, map[string]tftypes.Value{
		app.MEMORY_LIMIT: tftypes.NewValue(tftypes.String, "256"),
		"MY_VARIABLE":    tftypes.NewValue(tftypes.String, "free"),
	})
	config := tfsdk.Config{Schema: schemaPHP, Raw: tftypes.NewValue(typ, attrs)}

	var errs []string
	for _, v := range (&ResourcePHP{}).ConfigValidators(ctx) {
		res := &resource.ValidateConfigResponse{}
		v.ValidateResource(ctx, resource.ValidateConfigRequest{Config: config}, res)
		for _, d := range res.Diagnostics.Errors() {
			errs = append(errs, d.Detail())
		}
	}
	if len(errs) != 1 || !strings.HasPrefix(errs[0], "MEMORY_LIMIT is managed by an attribute") {
		t.Errorf("got %v, want one error about MEMORY_LIMIT", errs)
	}
}
