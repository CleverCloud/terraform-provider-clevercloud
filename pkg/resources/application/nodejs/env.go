package nodejs

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/resources/application"
	"go.clever-cloud.dev/sdk/app"
)

// EnvVar is a nodejs environment variable backing an attribute; its value is
// the variable name.
type EnvVar string

const (
	DevDependencies      EnvVar = app.CC_NODE_DEV_DEPENDENCIES
	StartScript          EnvVar = app.CC_RUN_COMMAND
	PackageManager       EnvVar = app.CC_NODE_BUILD_TOOL
	Registry             EnvVar = app.CC_NPM_REGISTRY
	RegistryToken        EnvVar = app.NPM_TOKEN
	NodeVersion          EnvVar = app.CC_NODE_VERSION
	CustomInstallCommand EnvVar = app.CC_CUSTOM_BUILD_TOOL
	RegistryBasicAuth    EnvVar = app.CC_NPM_BASIC_AUTH
	BunInstallCacheDir   EnvVar = app.CC_BUN_INSTALL_CACHE_DIR
)

// envVars lists every nodejs attribute backed by an environment variable.
var envVars = []pkg.EnvVarMapping[*NodeJS]{
	application.OnRuntime[*NodeJS](application.AppFolder),
	DevDependencies,
	StartScript,
	PackageManager,
	Registry,
	RegistryToken,
	NodeVersion,
	CustomInstallCommand,
	RegistryBasicAuth,
	BunInstallCacheDir,
}

func (v EnvVar) FromEnv(n *NodeJS, value string) {
	switch v {
	case DevDependencies:
		devDependenciesFromEnv(n, value)
	default:
		*v.attribute(n) = pkg.FromStr(value)
	}
}

func (v EnvVar) ToEnv(n *NodeJS) (string, string) {
	switch v {
	case DevDependencies:
		return string(v), devDependenciesToEnv(n)
	default:
		return string(v), v.attribute(n).ValueString()
	}
}

// devDependenciesFromEnv sets dev_dependencies to true when
// CC_NODE_DEV_DEPENDENCIES is "install". Any other value, or an absent
// variable, leaves the attribute as it is.
func devDependenciesFromEnv(n *NodeJS, value string) {
	if value == "install" {
		n.DevDependencies = types.BoolValue(true)
	}
}

// devDependenciesToEnv returns "install" when dev_dependencies is true, and
// nothing otherwise so the variable is not written.
func devDependenciesToEnv(n *NodeJS) string {
	if n.DevDependencies.ValueBool() {
		return "install"
	}
	return ""
}

// attribute returns the string attribute v backs; DevDependencies, a
// boolean, has its own functions.
func (v EnvVar) attribute(n *NodeJS) *types.String {
	switch v {
	case StartScript:
		return &n.StartScript
	case PackageManager:
		return &n.PackageManager
	case Registry:
		return &n.Registry
	case RegistryToken:
		return &n.RegistryToken
	case NodeVersion:
		return &n.NodeVersion
	case CustomInstallCommand:
		return &n.CustomInstallCommand
	case RegistryBasicAuth:
		return &n.RegistryBasicAuth
	case BunInstallCacheDir:
		return &n.BunInstallCacheDir
	}
	panic("no nodejs attribute for " + string(v))
}
