package staticapache

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/resources/application"
	"go.clever-cloud.dev/sdk/app"
)

var WebrootEnv = pkg.NewStringEnv[StaticApache](app.CC_WEBROOT,
	func(s *StaticApache) types.String { return s.Webroot },
	func(s *StaticApache, v types.String) { s.Webroot = v })

// envVars lists every static_apache attribute backed by an environment
// variable.
var envVars = []pkg.EnvVarMapping[*StaticApache]{
	application.OnRuntime[*StaticApache](application.AppFolderEnv),
	WebrootEnv,
}
