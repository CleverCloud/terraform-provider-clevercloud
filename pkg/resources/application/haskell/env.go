package haskell

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/resources/application"
	"go.clever-cloud.dev/sdk/app"
)

var (
	StackTargetEnv = pkg.NewStringEnv[Haskell](app.CC_HASKELL_STACK_TARGET,
		func(h *Haskell) types.String { return h.StackTarget },
		func(h *Haskell, v types.String) { h.StackTarget = v })
	StackSetupCommandEnv = pkg.NewStringEnv[Haskell](app.CC_HASKELL_STACK_SETUP_COMMAND,
		func(h *Haskell) types.String { return h.StackSetupCommand },
		func(h *Haskell, v types.String) { h.StackSetupCommand = v })
	StackInstallCommandEnv = pkg.NewStringEnv[Haskell](app.CC_HASKELL_STACK_INSTALL_COMMAND,
		func(h *Haskell) types.String { return h.StackInstallCommand },
		func(h *Haskell, v types.String) { h.StackInstallCommand = v })
	StackInstallDependenciesCommandEnv = pkg.NewStringEnv[Haskell](app.CC_HASKELL_STACK_INSTALL_DEPENDENCIES_COMMAND,
		func(h *Haskell) types.String { return h.StackInstallDependenciesCommand },
		func(h *Haskell, v types.String) { h.StackInstallDependenciesCommand = v })
)

// envVars lists every haskell attribute backed by an environment variable.
var envVars = []pkg.EnvVarMapping[*Haskell]{
	application.OnRuntime[*Haskell](application.AppFolderEnv),
	StackTargetEnv,
	StackSetupCommandEnv,
	StackInstallCommandEnv,
	StackInstallDependenciesCommandEnv,
}
