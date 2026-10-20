package docker

import (
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/resources/application"
	"go.clever-cloud.dev/sdk/app"
)

// The runtime logs in to a private registry with these variables, which the
// SDK does not list yet.
const (
	dockerLoginServer   = "CC_DOCKER_LOGIN_SERVER"
	dockerLoginUsername = "CC_DOCKER_LOGIN_USERNAME"
	dockerLoginPassword = "CC_DOCKER_LOGIN_PASSWORD"
)

// boolToEnv writes "true" or "false" when the attribute is set, nothing
// otherwise.
func boolToEnv(b types.Bool) string {
	if b.IsNull() || b.IsUnknown() {
		return ""
	}
	return strconv.FormatBool(b.ValueBool())
}

// boolFromEnv parses the variable value, null when it is absent or invalid.
func boolFromEnv(value string) types.Bool {
	if value == "" {
		return types.BoolNull()
	}
	return pkg.FromBoolPtr(&value)
}

var (
	DockerfileEnv = pkg.NewStringEnv[Docker](app.CC_DOCKERFILE,
		func(d *Docker) types.String { return d.Dockerfile },
		func(d *Docker, v types.String) { d.Dockerfile = v })
	ContainerPortEnv = pkg.NewInt64Env[Docker](app.CC_DOCKER_EXPOSED_HTTP_PORT,
		func(d *Docker) types.Int64 { return d.ContainerPort },
		func(d *Docker, v types.Int64) { d.ContainerPort = v })
	ContainerPortTCPEnv = pkg.NewInt64Env[Docker](app.CC_DOCKER_EXPOSED_TCP_PORT,
		func(d *Docker) types.Int64 { return d.ContainerPortTCP },
		func(d *Docker, v types.Int64) { d.ContainerPortTCP = v })
	IPv6CidrEnv = pkg.NewStringEnv[Docker](app.CC_DOCKER_FIXED_CIDR_V6,
		func(d *Docker) types.String { return d.IPv6Cidr },
		func(d *Docker, v types.String) { d.IPv6Cidr = v })
	RegistryURLEnv = pkg.NewStringEnv[Docker](dockerLoginServer,
		func(d *Docker) types.String { return d.RegistryURL },
		func(d *Docker, v types.String) { d.RegistryURL = v })
	RegistryUserEnv = pkg.NewStringEnv[Docker](dockerLoginUsername,
		func(d *Docker) types.String { return d.RegistryUser },
		func(d *Docker, v types.String) { d.RegistryUser = v })
	RegistryPasswordEnv = pkg.NewStringEnv[Docker](dockerLoginPassword,
		func(d *Docker) types.String { return d.RegistryPassword },
		func(d *Docker, v types.String) { d.RegistryPassword = v })
	DaemonSocketMountEnv = pkg.NewBoolEnv[Docker](app.CC_MOUNT_DOCKER_SOCKET,
		func(d *Docker) string { return boolToEnv(d.DaemonSocketMount) },
		func(d *Docker, value string) { d.DaemonSocketMount = boolFromEnv(value) })
	BuildxEnv = pkg.NewBoolEnv[Docker](app.CC_DOCKER_BUILDX,
		func(d *Docker) string { return boolToEnv(d.Buildx) },
		func(d *Docker, value string) { d.Buildx = boolFromEnv(value) })
)

// envVars lists every docker attribute backed by an environment variable.
var envVars = []pkg.EnvVarMapping[*Docker]{
	application.OnRuntime[*Docker](application.AppFolderEnv),
	DockerfileEnv,
	ContainerPortEnv,
	ContainerPortTCPEnv,
	IPv6CidrEnv,
	RegistryURLEnv,
	RegistryUserEnv,
	RegistryPasswordEnv,
	DaemonSocketMountEnv,
	BuildxEnv,
}
