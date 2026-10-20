package v

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/resources/application"
	"go.clever-cloud.dev/sdk/app"
)

var (
	BinaryEnv = pkg.NewStringEnv[V](app.CC_V_BINARY,
		func(v *V) types.String { return v.Binary },
		func(v *V, value types.String) { v.Binary = value })
	DevelopmentBuildEnv = pkg.NewBoolEnv[V](app.ENVIRONMENT,
		func(v *V) string {
			if v.DevelopmentBuild.ValueBool() {
				return "development"
			}
			return ""
		},
		func(v *V, value string) {
			// the attribute defaults to false: an absent variable reads as false
			v.DevelopmentBuild = types.BoolValue(value == "development")
		})
	ModulesPathEnv = pkg.NewStringEnv[V](app.VMODULES,
		func(v *V) types.String { return v.ModulesPath },
		func(v *V, value types.String) { v.ModulesPath = value })
)

// envVars lists every v attribute backed by an environment variable.
var envVars = []pkg.EnvVarMapping[*V]{
	application.OnRuntime[*V](application.AppFolderEnv),
	BinaryEnv,
	DevelopmentBuildEnv,
	ModulesPathEnv,
}
