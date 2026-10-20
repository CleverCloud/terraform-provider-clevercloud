package nodejs

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/miton18/helper/maps"
	"go.clever-cloud.com/terraform-provider/pkg/resources/application"
	"go.clever-cloud.dev/sdk/app"
)

// Every variable listed in envAttributes is written by ToEnv and read back by
// FromEnv into the attribute it names, so the reserved list cannot drift
// from the code.
func TestEnvAttributesRoundTrip(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	for _, attribute := range envAttributes {
		if _, ok := schemaNodeJS.Attributes[attribute]; !ok {
			t.Errorf("envAttributes names %q, which is not a nodejs attribute", attribute)
		}
	}

	node := NodeJS{
		Runtime:              application.Runtime{Environment: types.MapNull(types.StringType)},
		DevDependencies:      types.BoolValue(true),
		StartScript:          types.StringValue("node server.js"),
		PackageManager:       types.StringValue("custom"),
		Registry:             types.StringValue("npm.example.com"),
		RegistryToken:        types.StringValue("token"),
		NodeVersion:          types.StringValue("22"),
		CustomInstallCommand: types.StringValue("make deps"),
		RegistryBasicAuth:    types.StringValue("user:password"),
		BunInstallCacheDir:   types.StringValue("/tmp/bun"),
	}
	env := node.ToEnv(ctx, &diags)
	for variable := range envAttributes {
		if _, ok := env[variable]; !ok {
			t.Errorf("ToEnv does not write %s", variable)
		}
	}

	read := NodeJS{}
	envMap := maps.NewMap(env)
	read.FromEnv(ctx, envMap, &diags)
	for variable := range envAttributes {
		if envMap.Has(variable) {
			t.Errorf("FromEnv does not read %s back", variable)
		}
	}
	if diags.HasError() {
		t.Fatal(diags)
	}
	read.Integrations = node.Integrations
	if read.NodeVersion != node.NodeVersion || read.CustomInstallCommand != node.CustomInstallCommand ||
		read.RegistryBasicAuth != node.RegistryBasicAuth || read.BunInstallCacheDir != node.BunInstallCacheDir {
		t.Errorf("round trip lost a value: got %+v", read)
	}
}

func validate(t *testing.T, values map[string]tftypes.Value) []string {
	t.Helper()
	ctx := context.Background()
	typ := schemaNodeJS.Type().TerraformType(ctx).(tftypes.Object)
	attrs := map[string]tftypes.Value{}
	for name, attrType := range typ.AttributeTypes {
		attrs[name] = tftypes.NewValue(attrType, nil)
	}
	for name, v := range values {
		attrs[name] = v
	}
	config := tfsdk.Config{Schema: schemaNodeJS, Raw: tftypes.NewValue(typ, attrs)}

	var errs []string
	for _, v := range (&ResourceNodeJS{}).ConfigValidators(ctx) {
		res := &resource.ValidateConfigResponse{}
		v.ValidateResource(ctx, resource.ValidateConfigRequest{Config: config}, res)
		for _, d := range res.Diagnostics.Errors() {
			errs = append(errs, d.Detail())
		}
	}
	return errs
}

func str(s string) tftypes.Value { return tftypes.NewValue(tftypes.String, s) }

func TestConfigValidators(t *testing.T) {
	envType := tftypes.Map{ElementType: tftypes.String}
	for name, tc := range map[string]struct {
		values map[string]tftypes.Value
		want   []string
	}{
		"free variables": {
			values: map[string]tftypes.Value{"environment": tftypes.NewValue(envType, map[string]tftypes.Value{"PORT_NAME": str("x")})},
		},
		"node variable in environment": {
			values: map[string]tftypes.Value{"environment": tftypes.NewValue(envType, map[string]tftypes.Value{app.CC_NODE_VERSION: str("22")})},
			want:   []string{`CC_NODE_VERSION is managed by the "node_version" attribute`},
		},
		"common variable in environment": {
			values: map[string]tftypes.Value{"environment": tftypes.NewValue(envType, map[string]tftypes.Value{app.CC_CLAMAV: str("true")})},
			want:   []string{`CC_CLAMAV is managed by the "integrations.clamav.enabled" attribute`},
		},
		"custom command with the custom manager": {
			values: map[string]tftypes.Value{"custom_install_command": str("make deps"), "package_manager": str("custom")},
		},
		"custom command with another manager": {
			values: map[string]tftypes.Value{"custom_install_command": str("make deps"), "package_manager": str("npm")},
			want:   []string{"custom_install_command is only used when package_manager is \"custom\""},
		},
		"custom command without manager": {
			values: map[string]tftypes.Value{"custom_install_command": str("make deps")},
			want:   []string{"custom_install_command is only used when package_manager is \"custom\""},
		},
		"custom command with an unknown manager": {
			values: map[string]tftypes.Value{"custom_install_command": str("make deps"), "package_manager": tftypes.NewValue(tftypes.String, tftypes.UnknownValue)},
		},
	} {
		got := validate(t, tc.values)
		if len(got) != len(tc.want) {
			t.Errorf("%s: got %v, want %v", name, got, tc.want)
			continue
		}
		for i := range got {
			if !strings.HasPrefix(got[i], tc.want[i]) {
				t.Errorf("%s: got %q, want it to start with %q", name, got[i], tc.want[i])
			}
		}
	}
}
