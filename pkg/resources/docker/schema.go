package docker

import (
	"context"
	_ "embed"
	"fmt"
	"net"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/application"
	"go.clever-cloud.com/terraform-provider/pkg/attributes"
)

type Docker struct {
	attributes.Runtime
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
	res.Schema = schema.Schema{
		Version:             0,
		MarkdownDescription: dockerDoc,
		Attributes: attributes.WithRuntimeCommons(map[string]schema.Attribute{
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
				MarkdownDescription: "The password of your username",
			},
			"daemon_socket_mount": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Set to true to access the host Docker socket from inside your container",
			},
		}),
		Blocks: attributes.WithBlockRuntimeCommons(map[string]schema.Block{}),
	}
}

// https://developer.hashicorp.com/terraform/plugin/framework/resources/state-upgrade#implementing-state-upgrade-support
func (p *Docker) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{}
}

func (p *Docker) toEnv(ctx context.Context, diags diag.Diagnostics) map[string]string {
	env := map[string]string{}

	// do not use the real map since ElementAs can nullish it
	// https://github.com/hashicorp/terraform-plugin-framework/issues/698
	customEnv := map[string]string{}
	diags.Append(p.Environment.ElementsAs(ctx, &customEnv, false)...)
	if diags.HasError() {
		return env
	}
	env = pkg.Merge(env, customEnv)

	pkg.IfIsSet(p.AppFolder, func(s string) { env["APP_FOLDER"] = s })

	// Docker specific
	pkg.IfIsSet(p.Dockerfile, func(s string) { env["CC_DOCKERFILE"] = s })
	pkg.IfIsSetI(p.ContainerPort, func(i int64) { env["CC_DOCKER_EXPOSED_HTTP_PORT"] = fmt.Sprintf("%d", i) })
	pkg.IfIsSetI(p.ContainerPortTCP, func(i int64) { env["CC_DOCKER_EXPOSED_TCP_PORT"] = fmt.Sprintf("%d", i) })
	pkg.IfIsSet(p.IPv6Cidr, func(s string) { env["CC_DOCKER_FIXED_CIDR_V6"] = s })
	pkg.IfIsSet(p.RegistryURL, func(s string) { env["CC_DOCKER_LOGIN_SERVER"] = s })
	pkg.IfIsSet(p.RegistryUser, func(s string) { env["CC_DOCKER_LOGIN_USERNAME"] = s })
	pkg.IfIsSet(p.RegistryPassword, func(s string) { env["CC_DOCKER_LOGIN_PASSWORD"] = s })
	pkg.IfIsSetB(p.DaemonSocketMount, func(e bool) { env["CC_MOUNT_DOCKER_SOCKET"] = strconv.FormatBool(e) })

	env = pkg.Merge(env, p.Hooks.ToEnv())

	return env
}

func (p *Docker) toDeployment() *application.Deployment {
	if p.Deployment == nil || p.Deployment.Repository.IsNull() {
		return nil
	}

	return &application.Deployment{
		Repository: p.Deployment.Repository.ValueString(),
		Commit:     p.Deployment.Commit.ValueStringPointer(),
	}
}
