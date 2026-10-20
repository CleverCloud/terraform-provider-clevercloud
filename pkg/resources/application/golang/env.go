package golang

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/resources/application"
	"go.clever-cloud.dev/sdk/app"
)

var (
	BinaryEnv = pkg.NewStringEnv[Go](app.CC_GO_BINARY,
		func(g *Go) types.String { return g.Binary },
		func(g *Go, v types.String) { g.Binary = v })
	BuildToolEnv = pkg.NewStringEnv[Go](app.CC_GO_BUILD_TOOL,
		func(g *Go) types.String { return g.BuildTool },
		func(g *Go, v types.String) { g.BuildTool = v })
	PackageEnv = pkg.NewStringEnv[Go](app.CC_GO_PKG,
		func(g *Go) types.String { return g.Package },
		func(g *Go, v types.String) { g.Package = v })
	RunDirEnv = pkg.NewStringEnv[Go](app.CC_GO_RUNDIR,
		func(g *Go) types.String { return g.RunDir },
		func(g *Go, v types.String) { g.RunDir = v })
)

// envVars lists every go attribute backed by an environment variable.
var envVars = []pkg.EnvVarMapping[*Go]{
	application.OnRuntime[*Go](application.AppFolderEnv),
	BinaryEnv,
	BuildToolEnv,
	PackageEnv,
	RunDirEnv,
}
