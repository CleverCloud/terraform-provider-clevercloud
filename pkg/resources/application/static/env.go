package static

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/resources/application"
	"go.clever-cloud.dev/sdk/app"
)

var (
	WebrootEnv = pkg.NewStringEnv[Static](app.CC_WEBROOT,
		func(s *Static) types.String { return s.Webroot },
		func(s *Static, v types.String) { s.Webroot = v })
	BuildCommandEnv = pkg.NewStringEnv[Static](app.CC_BUILD_COMMAND,
		func(s *Static) types.String { return s.BuildCommand },
		func(s *Static, v types.String) { s.BuildCommand = v })
	HugoVersionEnv = pkg.NewStringEnv[Static](app.CC_HUGO_VERSION,
		func(s *Static) types.String { return s.HugoVersion },
		func(s *Static, v types.String) { s.HugoVersion = v })
	AutobuildOutdirEnv = pkg.NewStringEnv[Static](app.CC_STATIC_AUTOBUILD_OUTDIR,
		func(s *Static) types.String { return s.AutobuildOutdir },
		func(s *Static, v types.String) { s.AutobuildOutdir = v })
	StaticServerEnv = pkg.NewStringEnv[Static](app.CC_STATIC_SERVER,
		func(s *Static) types.String { return s.StaticServer },
		func(s *Static, v types.String) { s.StaticServer = v })
	StaticServerFlagsEnv = pkg.NewStringEnv[Static](app.CC_STATIC_FLAGS,
		func(s *Static) types.String { return s.StaticServerFlags },
		func(s *Static, v types.String) { s.StaticServerFlags = v })
	CaddyfileEnv = pkg.NewStringEnv[Static](app.CC_STATIC_CADDYFILE,
		func(s *Static) types.String { return s.Caddyfile },
		func(s *Static, v types.String) { s.Caddyfile = v })
)

// envVars lists every static attribute backed by an environment variable.
var envVars = []pkg.EnvVarMapping[*Static]{
	application.OnRuntime[*Static](application.AppFolderEnv),
	WebrootEnv,
	BuildCommandEnv,
	HugoVersionEnv,
	AutobuildOutdirEnv,
	StaticServerEnv,
	StaticServerFlagsEnv,
	CaddyfileEnv,
}
