package nodejs

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/resources/application"
	"go.clever-cloud.dev/sdk/app"
)

var (
	DevDependenciesEnv = pkg.NewBoolEnv[NodeJS](app.CC_NODE_DEV_DEPENDENCIES, "install",
		func(n *NodeJS) types.Bool { return n.DevDependencies },
		func(n *NodeJS, v types.Bool) { n.DevDependencies = v })
	StartScriptEnv = pkg.NewStringEnv[NodeJS](app.CC_RUN_COMMAND,
		func(n *NodeJS) types.String { return n.StartScript },
		func(n *NodeJS, v types.String) { n.StartScript = v })
	PackageManagerEnv = pkg.NewStringEnv[NodeJS](app.CC_NODE_BUILD_TOOL,
		func(n *NodeJS) types.String { return n.PackageManager },
		func(n *NodeJS, v types.String) { n.PackageManager = v })
	RegistryEnv = pkg.NewStringEnv[NodeJS](app.CC_NPM_REGISTRY,
		func(n *NodeJS) types.String { return n.Registry },
		func(n *NodeJS, v types.String) { n.Registry = v })
	RegistryTokenEnv = pkg.NewStringEnv[NodeJS](app.NPM_TOKEN,
		func(n *NodeJS) types.String { return n.RegistryToken },
		func(n *NodeJS, v types.String) { n.RegistryToken = v })
	NodeVersionEnv = pkg.NewStringEnv[NodeJS](app.CC_NODE_VERSION,
		func(n *NodeJS) types.String { return n.NodeVersion },
		func(n *NodeJS, v types.String) { n.NodeVersion = v })
	CustomInstallCommandEnv = pkg.NewStringEnv[NodeJS](app.CC_CUSTOM_BUILD_TOOL,
		func(n *NodeJS) types.String { return n.CustomInstallCommand },
		func(n *NodeJS, v types.String) { n.CustomInstallCommand = v })
	RegistryBasicAuthEnv = pkg.NewStringEnv[NodeJS](app.CC_NPM_BASIC_AUTH,
		func(n *NodeJS) types.String { return n.RegistryBasicAuth },
		func(n *NodeJS, v types.String) { n.RegistryBasicAuth = v })
	BunInstallCacheDirEnv = pkg.NewStringEnv[NodeJS](app.CC_BUN_INSTALL_CACHE_DIR,
		func(n *NodeJS) types.String { return n.BunInstallCacheDir },
		func(n *NodeJS, v types.String) { n.BunInstallCacheDir = v })
)

// envVars lists every nodejs attribute backed by an environment variable.
var envVars = []pkg.EnvVarMapping[*NodeJS]{
	application.OnRuntime[*NodeJS](application.AppFolderEnv),
	DevDependenciesEnv,
	StartScriptEnv,
	PackageManagerEnv,
	RegistryEnv,
	RegistryTokenEnv,
	NodeVersionEnv,
	CustomInstallCommandEnv,
	RegistryBasicAuthEnv,
	BunInstallCacheDirEnv,
}
