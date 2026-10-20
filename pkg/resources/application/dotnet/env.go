package dotnet

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/resources/application"
	"go.clever-cloud.dev/sdk/app"
)

var (
	ProfileEnv = pkg.NewStringEnv[Dotnet](app.CC_DOTNET_PROFILE,
		func(d *Dotnet) types.String { return d.DotnetProfile },
		func(d *Dotnet, v types.String) { d.DotnetProfile = v })
	ProjEnv = pkg.NewStringEnv[Dotnet](app.CC_DOTNET_PROJ,
		func(d *Dotnet) types.String { return d.DotnetProj },
		func(d *Dotnet, v types.String) { d.DotnetProj = v })
	TFMEnv = pkg.NewStringEnv[Dotnet](app.CC_DOTNET_TFM,
		func(d *Dotnet) types.String { return d.DotnetTFM },
		func(d *Dotnet, v types.String) { d.DotnetTFM = v })
	VersionEnv = pkg.NewStringEnv[Dotnet](app.CC_DOTNET_VERSION,
		func(d *Dotnet) types.String { return d.DotnetVersion },
		func(d *Dotnet, v types.String) { d.DotnetVersion = v })
)

// envVars lists every dotnet attribute backed by an environment variable.
var envVars = []pkg.EnvVarMapping[*Dotnet]{
	application.OnRuntime[*Dotnet](application.AppFolderEnv),
	ProfileEnv,
	ProjEnv,
	TFMEnv,
	VersionEnv,
}
