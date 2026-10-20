package ruby

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

	plan := Ruby{
		Runtime: application.Runtime{
			Environment: types.MapNull(types.StringType),
			AppFolder:   types.StringValue("web"),
		},
		RubyVersion:           types.StringValue("3.3"),
		EnableSidekiq:         types.BoolValue(true),
		RackupServer:          types.StringValue("puma"),
		RakeGoals:             types.StringValue("db:migrate"),
		SidekiqFiles:          types.StringValue("./config/sidekiq.yml"),
		HTTPBasicAuth:         types.StringValue("login:password"),
		NginxProxyBuffers:     types.StringValue("16 32k"),
		NginxProxyBufferSize:  types.StringValue("64k"),
		EnableGzipCompression: types.BoolValue(true),
		GzipTypes:             types.StringValue("text/html"),
		NginxReadTimeout:      types.Int64Value(600),
		RackEnv:               types.StringValue("production"),
		RailsEnv:              types.StringValue("production"),
		StaticFilesPath:       types.StringValue("public"),
		StaticURLPrefix:       types.StringValue("/assets"),
		StaticWebroot:         types.StringValue("public"),
	}
	env := plan.ToEnv(ctx, &diags)
	for _, name := range pkg.EnvNames(envVars...) {
		if _, ok := env[name]; !ok {
			t.Errorf("ToEnv does not write %s", name)
		}
	}

	read := Ruby{}
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
	typ := schemaRuby.Type().TerraformType(ctx).(tftypes.Object)
	attrs := map[string]tftypes.Value{}
	for name, attrType := range typ.AttributeTypes {
		attrs[name] = tftypes.NewValue(attrType, nil)
	}
	attrs["environment"] = tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, map[string]tftypes.Value{
		app.CC_RUBY_VERSION: tftypes.NewValue(tftypes.String, "3.3"),
		"STATIC_WEBROOT":    tftypes.NewValue(tftypes.String, "public"),
		"MY_VARIABLE":       tftypes.NewValue(tftypes.String, "free"),
	})
	config := tfsdk.Config{Schema: schemaRuby, Raw: tftypes.NewValue(typ, attrs)}

	var errs []string
	for _, v := range (&ResourceRuby{}).ConfigValidators(ctx) {
		res := &resource.ValidateConfigResponse{}
		v.ValidateResource(ctx, resource.ValidateConfigRequest{Config: config}, res)
		for _, d := range res.Diagnostics.Errors() {
			errs = append(errs, d.Detail())
		}
	}
	if len(errs) != 1 || !strings.HasPrefix(errs[0], "CC_RUBY_VERSION is managed by an attribute") {
		t.Errorf("got %v, want one error about CC_RUBY_VERSION", errs)
	}
}
