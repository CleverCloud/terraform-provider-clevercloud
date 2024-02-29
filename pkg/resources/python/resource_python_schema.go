package python

import (
	"context"
	_ "embed"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/application"
	"go.clever-cloud.com/terraform-provider/pkg/attributes"
)

type Python struct {
	// Common
	ID               types.String           `tfsdk:"id"`
	Name             types.String           `tfsdk:"name"`
	Description      types.String           `tfsdk:"description"`
	MinInstanceCount types.Int64            `tfsdk:"min_instance_count"`
	MaxInstanceCount types.Int64            `tfsdk:"max_instance_count"`
	SmallestFlavor   types.String           `tfsdk:"smallest_flavor"`
	BiggestFlavor    types.String           `tfsdk:"biggest_flavor"`
	BuildFlavor      types.String           `tfsdk:"build_flavor"`
	Region           types.String           `tfsdk:"region"`
	StickySessions   types.Bool             `tfsdk:"sticky_sessions"`
	RedirectHTTPS    types.Bool             `tfsdk:"redirect_https"`
	VHost            types.String           `tfsdk:"vhost"`
	AdditionalVHosts types.List             `tfsdk:"additional_vhosts"`
	DeployURL        types.String           `tfsdk:"deploy_url"`
	Deployment       *attributes.Deployment `tfsdk:"deployment"`
	Hooks            *attributes.Hooks      `tfsdk:"hooks"`
	Dependencies     types.Set              `tfsdk:"dependencies"`

	// Env
	AppFolder   types.String `tfsdk:"app_folder"`
	Environment types.Map    `tfsdk:"environment"`

	// Python
	PythonVersion   types.String `tfsdk:"python_version"`
	PipRequirements types.String `tfsdk:"pip_requirements"`
}

//go:embed resource_python.md
var pythonDoc string

func (r ResourcePython) Schema(ctx context.Context, req resource.SchemaRequest, res *resource.SchemaResponse) {

	res.Schema = schema.Schema{
		Version:             0,
		MarkdownDescription: pythonDoc,
		Attributes: attributes.WithRuntimeCommons(map[string]schema.Attribute{
			// CC_PYTHON_VERSION
			"python_version": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Python version >= 2.7",
			},
			// CC_PIP_REQUIREMENTS_FILE
			"pip_requirements": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Define a custom requirements.txt file (default: requirements.txt)",
			},
		}),
		Blocks: attributes.WithBlockRuntimeCommons(map[string]schema.Block{}),
	}
}

// https://developer.hashicorp.com/terraform/plugin/framework/resources/state-upgrade#implementing-state-upgrade-support
func (r ResourcePython) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{}
}

func (py Python) toEnv(ctx context.Context, diags diag.Diagnostics) map[string]string {
	env := map[string]string{}

	// do not use the real map since ElementAs can nullish it
	// https://github.com/hashicorp/terraform-plugin-framework/issues/698
	customEnv := map[string]string{}
	diags.Append(py.Environment.ElementsAs(ctx, &customEnv, false)...)
	if diags.HasError() {
		return env
	}
	env = pkg.Merge(env, customEnv)

	pkg.IfIsSet(py.AppFolder, func(s string) { env["APP_FOLDER"] = s })
	pkg.IfIsSet(py.PythonVersion, func(version string) { env["CC_PYTHON_VERSION"] = version })
	pkg.IfIsSet(py.PipRequirements, func(pipReqFile string) { env["CC_PIP_REQUIREMENTS_FILE"] = pipReqFile })

	env = pkg.Merge(env, py.Hooks.ToEnv())
	return env
}

func (py Python) toDeployment() *application.Deployment {
	if py.Deployment == nil || py.Deployment.Repository.IsNull() {
		return nil
	}

	return &application.Deployment{
		Repository: py.Deployment.Repository.ValueString(),
		Commit:     py.Deployment.Commit.ValueStringPointer(),
	}
}
