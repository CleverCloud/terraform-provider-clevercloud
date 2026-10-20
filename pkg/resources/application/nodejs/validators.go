package nodejs

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/resources/application"
)

var _ resource.ResourceWithConfigValidators = (*ResourceNodeJS)(nil)

// ConfigValidators rejects the environment variables an attribute manages and
// a custom install command without the custom package manager.
func (r *ResourceNodeJS) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		application.ReservedEnvironment(pkg.EnvNames(envVars...)...),
		customInstallCommand{},
	}
}

// customInstallCommand requires package_manager = "custom" when
// custom_install_command is set: the platform ignores the command otherwise.
type customInstallCommand struct{}

func (v customInstallCommand) Description(ctx context.Context) string {
	return `custom_install_command requires package_manager = "custom"`
}

func (v customInstallCommand) MarkdownDescription(ctx context.Context) string {
	return "`custom_install_command` requires `package_manager = \"custom\"`"
}

func (v customInstallCommand) ValidateResource(ctx context.Context, req resource.ValidateConfigRequest, res *resource.ValidateConfigResponse) {
	var command, manager types.String
	res.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("custom_install_command"), &command)...)
	res.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("package_manager"), &manager)...)
	if res.Diagnostics.HasError() || command.IsNull() || command.IsUnknown() || manager.IsUnknown() {
		return
	}
	if manager.ValueString() != "custom" {
		res.Diagnostics.AddAttributeError(
			path.Root("custom_install_command"),
			"custom_install_command without the custom package manager",
			`custom_install_command is only used when package_manager is "custom": set package_manager = "custom" or remove custom_install_command.`,
		)
	}
}
