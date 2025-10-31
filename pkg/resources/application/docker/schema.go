package docker

import (
	"context"
	_ "embed"
	"fmt"
	"net"
	"strconv"

	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/attributes"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/resources/application"
)

type Docker struct {
	application.Runtime
	Dockerfile        types.String `tfsdk:"dockerfile"`
	Buildx            types.Bool   `tfsdk:"buildx"`
	ContainerPort     types.Int64  `tfsdk:"container_port"`
	ContainerPortTCP  types.Int64  `tfsdk:"container_port_tcp"`
	EnableIPv6        types.Bool   `tfsdk:"enable_ipv6"`
	IPv6Cidr          types.String `tfsdk:"ipv6_cidr"`
	RegistryURL       types.String `tfsdk:"registry_url"`
	RegistryUser      types.String `tfsdk:"registry_user"`
	RegistryPassword  types.String `tfsdk:"registry_password"`
	DaemonSocketMount types.Bool   `tfsdk:"daemon_socket_mount"`
}

type DockerV0 struct {
	application.RuntimeV0
	Dockerfile        types.String `tfsdk:"dockerfile"`
	ContainerPort     types.Int64  `tfsdk:"container_port"`
	ContainerPortTCP  types.Int64  `tfsdk:"container_port_tcp"`
	EnableIPv6        types.Bool   `tfsdk:"enable_ipv6"`
	IPv6Cidr          types.String `tfsdk:"ipv6_cidr"`
	RegistryURL       types.String `tfsdk:"registry_url"`
	RegistryUser      types.String `tfsdk:"registry_user"`
	RegistryPassword  types.String `tfsdk:"registry_password"`
	DaemonSocketMount types.Bool   `tfsdk:"daemon_socket_mount"`
}

//go:embed doc.md
var dockerDoc string

func (r ResourceDocker) Schema(ctx context.Context, req resource.SchemaRequest, res *resource.SchemaResponse) {
	res.Schema = schemaDocker
}

var schemaDocker = schema.Schema{
	Version:             1,
	MarkdownDescription: dockerDoc,
	Attributes: application.WithRuntimeCommons(map[string]schema.Attribute{
		"dockerfile": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "The name of the Dockerfile to build",
		},
		"buildx": schema.BoolAttribute{
			Optional:            true,
			Computed:            true,
			Default:             booldefault.StaticBool(false),
			MarkdownDescription: "Set to true to use buildx to build the Docker image",
		},
		"container_port": schema.Int64Attribute{
			Optional:            true,
			MarkdownDescription: "Set to custom HTTP port if your Docker container runs on custom port",
		},
		"container_port_tcp": schema.Int64Attribute{
			Optional:            true,
			MarkdownDescription: "Set to custom TCP port if your Docker container runs on custom port.",
		},
		"enable_ipv6": schema.BoolAttribute{
			Optional:           true,
			DeprecationMessage: "never works, please use `ipv6_cidr`",
		},
		"ipv6_cidr": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Activate the support of IPv6 with an IPv6 subnet int the docker daemon",
			Validators: []validator.String{
				pkg.NewValidator("IPv6 CIDR 🍾", func(_ context.Context, req validator.StringRequest, res *validator.StringResponse) {
					if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
						return
					}

					str := req.ConfigValue.ValueString()
					ip, _, err := net.ParseCIDR(str)
					if err != nil {
						res.Diagnostics.AddAttributeError(req.Path, "invalid IPv6 CIDR provided", err.Error())
					}

					if len(ip) != net.IPv6len {
						res.Diagnostics.AddAttributeError(req.Path, "invalid IPv6 CIDR provided", "expect an IPv6 before the mask")
					}
				}),
			},
		},
		"registry_url": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "The server of your private registry (optional).	Docker’s public registry",
		},
		"registry_user": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "The username to login to a private registry",
		},
		"registry_password": schema.StringAttribute{
			Optional:            true,
			Sensitive:           true,
			MarkdownDescription: "The password of your username",
		},
		"daemon_socket_mount": schema.BoolAttribute{
			Optional:            true,
			Computed:            true,
			Default:             booldefault.StaticBool(false),
			MarkdownDescription: "Set to true to access the host Docker socket from inside your container",
		},
	}),
	Blocks: attributes.WithBlockRuntimeCommons(map[string]schema.Block{}),
}

var schemaDockerV0 = schema.Schema{
	Version:             0,
	MarkdownDescription: dockerDoc,
	Attributes: application.WithRuntimeCommonsV0(map[string]schema.Attribute{
		"dockerfile": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "The name of the Dockerfile to build",
		},
		"container_port": schema.Int64Attribute{
			Optional:            true,
			MarkdownDescription: "Set to custom HTTP port if your Docker container runs on custom port",
		},
		"container_port_tcp": schema.Int64Attribute{
			Optional:            true,
			MarkdownDescription: "Set to custom TCP port if your Docker container runs on custom port.",
		},
		"enable_ipv6": schema.BoolAttribute{
			Optional:           true,
			DeprecationMessage: "never works, please use `ipv6_cidr`",
		},
		"ipv6_cidr": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Activate the support of IPv6 with an IPv6 subnet int the docker daemon",
			Validators: []validator.String{
				pkg.NewValidator("IPv6 CIDR 🍾", func(_ context.Context, req validator.StringRequest, res *validator.StringResponse) {
					if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
						return
					}

					str := req.ConfigValue.ValueString()
					ip, _, err := net.ParseCIDR(str)
					if err != nil {
						res.Diagnostics.AddAttributeError(req.Path, "invalid IPv6 CIDR provided", err.Error())
					}

					if len(ip) != net.IPv6len {
						res.Diagnostics.AddAttributeError(req.Path, "invalid IPv6 CIDR provided", "expect an IPv6 before the mask")
					}
				}),
			},
		},
		"registry_url": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "The server of your private registry (optional).	Docker’s public registry",
		},
		"registry_user": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "The username to login to a private registry",
		},
		"registry_password": schema.StringAttribute{
			Optional:            true,
			Sensitive:           true,
			MarkdownDescription: "The password of your username",
		},
		"daemon_socket_mount": schema.BoolAttribute{
			Optional:            true,
			Computed:            true,
			Default:             booldefault.StaticBool(false),
			MarkdownDescription: "Set to true to access the host Docker socket from inside your container",
		},
	}),
	Blocks: attributes.WithBlockRuntimeCommons(map[string]schema.Block{}),
}

func (p *Docker) toEnv(ctx context.Context, diags *diag.Diagnostics) map[string]string {
	// Start with common runtime environment variables (APP_FOLDER, Hooks, Environment)
	env := p.ToEnv(ctx, diags)
	if diags.HasError() {
		return env
	}

	// Add Docker-specific environment variables
	pkg.IfIsSetStr(p.Dockerfile, func(s string) { env[CC_DOCKERFILE] = s })
	pkg.IfIsSetB(p.Buildx, func(b bool) { env[CC_DOCKER_BUILDX] = strconv.FormatBool(b) })
	pkg.IfIsSetI(p.ContainerPort, func(i int64) { env[CC_DOCKER_EXPOSED_HTTP_PORT] = fmt.Sprintf("%d", i) })
	pkg.IfIsSetI(p.ContainerPortTCP, func(i int64) { env[CC_DOCKER_EXPOSED_TCP_PORT] = fmt.Sprintf("%d", i) })
	pkg.IfIsSetStr(p.IPv6Cidr, func(s string) { env[CC_DOCKER_FIXED_CIDR_V6] = s })
	pkg.IfIsSetStr(p.RegistryURL, func(s string) { env[CC_DOCKER_LOGIN_SERVER] = s })
	pkg.IfIsSetStr(p.RegistryUser, func(s string) { env[CC_DOCKER_LOGIN_USERNAME] = s })
	pkg.IfIsSetStr(p.RegistryPassword, func(s string) { env[CC_DOCKER_LOGIN_PASSWORD] = s })
	pkg.IfIsSetB(p.DaemonSocketMount, func(e bool) { env[CC_MOUNT_DOCKER_SOCKET] = strconv.FormatBool(e) })

	return env
}

// fromEnv iter on environment set on the clever application and
// handle language specific env vars
// put the others on Environment field
func (d *Docker) fromEnv(ctx context.Context, env map[string]string) diag.Diagnostics {
	diags := diag.Diagnostics{}
	m := helper.NewEnvMap(env)

	d.Dockerfile = pkg.FromStr(m.Pop(CC_DOCKERFILE))
	d.Buildx = pkg.FromBool(m.Pop(CC_DOCKER_BUILDX) == "true")

	if port, err := strconv.ParseInt(m.Pop(CC_DOCKER_EXPOSED_HTTP_PORT), 10, 64); err == nil {
		d.ContainerPort = pkg.FromI(port)
	}
	if port, err := strconv.ParseInt(m.Pop(CC_DOCKER_EXPOSED_TCP_PORT), 10, 64); err == nil {
		d.ContainerPortTCP = pkg.FromI(port)
	}

	d.IPv6Cidr = pkg.FromStr(m.Pop(CC_DOCKER_FIXED_CIDR_V6))
	d.RegistryURL = pkg.FromStr(m.Pop(CC_DOCKER_LOGIN_SERVER))
	d.RegistryUser = pkg.FromStr(m.Pop(CC_DOCKER_LOGIN_USERNAME))
	d.RegistryPassword = pkg.FromStr(m.Pop(CC_DOCKER_LOGIN_PASSWORD))
	d.DaemonSocketMount = pkg.FromBool(m.Pop(CC_MOUNT_DOCKER_SOCKET) == "true")

	d.FromEnvironment(ctx, m)
	return diags
}

func (p *Docker) toDeployment(gitAuth *http.BasicAuth) *application.Deployment {
	if p.Deployment == nil || p.Deployment.Repository.IsNull() {
		return nil
	}

	return &application.Deployment{
		Repository:    p.Deployment.Repository.ValueString(),
		Commit:        p.Deployment.Commit.ValueStringPointer(),
		CleverGitAuth: gitAuth,
	}
}
