package nodejs

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/resources/application"
	"go.clever-cloud.dev/sdk/app"
)

// envVars maps every nodejs attribute backed by an environment variable.
var envVars = []pkg.EnvVarMapping[*NodeJS]{
	application.OnRuntime[*NodeJS](application.AppFolder{}),
	DevDependencies{},
	StartScript{},
	PackageManager{},
	Registry{},
	RegistryToken{},
	NodeVersion{},
	CustomInstallCommand{},
	RegistryBasicAuth{},
	BunInstallCacheDir{},
}

// DevDependencies maps dev_dependencies to CC_NODE_DEV_DEPENDENCIES, written
// as "install" when true. An absent variable leaves the attribute as is.
type DevDependencies struct{}

func (DevDependencies) FromEnv(n *NodeJS, v string) {
	if v == "install" {
		n.DevDependencies = types.BoolValue(true)
	}
}

func (DevDependencies) ToEnv(n *NodeJS) (string, string) {
	if n.DevDependencies.ValueBool() {
		return app.CC_NODE_DEV_DEPENDENCIES, "install"
	}
	return app.CC_NODE_DEV_DEPENDENCIES, ""
}

type StartScript struct{}

func (StartScript) FromEnv(n *NodeJS, v string) { n.StartScript = pkg.FromStr(v) }
func (StartScript) ToEnv(n *NodeJS) (string, string) {
	return app.CC_RUN_COMMAND, n.StartScript.ValueString()
}

type PackageManager struct{}

func (PackageManager) FromEnv(n *NodeJS, v string) { n.PackageManager = pkg.FromStr(v) }
func (PackageManager) ToEnv(n *NodeJS) (string, string) {
	return app.CC_NODE_BUILD_TOOL, n.PackageManager.ValueString()
}

type Registry struct{}

func (Registry) FromEnv(n *NodeJS, v string) { n.Registry = pkg.FromStr(v) }
func (Registry) ToEnv(n *NodeJS) (string, string) {
	return app.CC_NPM_REGISTRY, n.Registry.ValueString()
}

type RegistryToken struct{}

func (RegistryToken) FromEnv(n *NodeJS, v string) { n.RegistryToken = pkg.FromStr(v) }
func (RegistryToken) ToEnv(n *NodeJS) (string, string) {
	return app.NPM_TOKEN, n.RegistryToken.ValueString()
}

type NodeVersion struct{}

func (NodeVersion) FromEnv(n *NodeJS, v string) { n.NodeVersion = pkg.FromStr(v) }
func (NodeVersion) ToEnv(n *NodeJS) (string, string) {
	return app.CC_NODE_VERSION, n.NodeVersion.ValueString()
}

type CustomInstallCommand struct{}

func (CustomInstallCommand) FromEnv(n *NodeJS, v string) { n.CustomInstallCommand = pkg.FromStr(v) }
func (CustomInstallCommand) ToEnv(n *NodeJS) (string, string) {
	return app.CC_CUSTOM_BUILD_TOOL, n.CustomInstallCommand.ValueString()
}

type RegistryBasicAuth struct{}

func (RegistryBasicAuth) FromEnv(n *NodeJS, v string) { n.RegistryBasicAuth = pkg.FromStr(v) }
func (RegistryBasicAuth) ToEnv(n *NodeJS) (string, string) {
	return app.CC_NPM_BASIC_AUTH, n.RegistryBasicAuth.ValueString()
}

type BunInstallCacheDir struct{}

func (BunInstallCacheDir) FromEnv(n *NodeJS, v string) { n.BunInstallCacheDir = pkg.FromStr(v) }
func (BunInstallCacheDir) ToEnv(n *NodeJS) (string, string) {
	return app.CC_BUN_INSTALL_CACHE_DIR, n.BunInstallCacheDir.ValueString()
}
