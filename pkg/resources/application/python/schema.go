package python

import (
	"context"
	_ "embed"
	"strings"

	"go.clever-cloud.com/terraform-provider/pkg/attributes"
	"go.clever-cloud.com/terraform-provider/pkg/resources/application"

	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg"
)

type Python struct {
	application.Runtime
	PythonVersion   types.String `tfsdk:"python_version"`
	PipRequirements types.String `tfsdk:"pip_requirements"`
}

type PythonV0 struct {
	application.RuntimeV0
	PythonVersion   types.String `tfsdk:"python_version"`
	PipRequirements types.String `tfsdk:"pip_requirements"`
}

//go:embed doc.md
var pythonDoc string

func (r ResourcePython) Schema(ctx context.Context, req resource.SchemaRequest, res *resource.SchemaResponse) {
	res.Schema = schemaPythonV1
}

var schemaPythonV1 = schema.Schema{
	Version:             1,
	MarkdownDescription: pythonDoc,
	Attributes: application.WithRuntimeCommons(map[string]schema.Attribute{
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

var schemaPythonV0 = schema.Schema{
	Version:             0,
	MarkdownDescription: pythonDoc,
	Attributes: application.WithRuntimeCommonsV0(map[string]schema.Attribute{
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

func (py Python) ToEnv(ctx context.Context, diags *diag.Diagnostics) map[string]string {
	env := map[string]string{}

	// do not use the real map since ElementAs can nullish it
	// https://github.com/hashicorp/terraform-plugin-framework/issues/698
	customEnv := map[string]string{}
	diags.Append(py.Environment.ElementsAs(ctx, &customEnv, false)...)
	if diags.HasError() {
		return env
	}
	env = pkg.Merge(env, customEnv)

	pkg.IfIsSetStr(py.AppFolder, func(s string) { env["APP_FOLDER"] = s })
	pkg.IfIsSetStr(py.PythonVersion, func(version string) { env["CC_PYTHON_VERSION"] = version })
	pkg.IfIsSetStr(py.PipRequirements, func(pipReqFile string) { env["CC_PIP_REQUIREMENTS_FILE"] = pipReqFile })

	env = pkg.Merge(env, py.Hooks.ToEnv())
	env = pkg.Merge(env, py.Integrations.ToEnv(ctx, diags))
	return env
}

func (py *Python) FromEnv(ctx context.Context, env pkg.EnvMap, diags *diag.Diagnostics) {
	py.AppFolder = pkg.FromStrPtr(env.Get("APP_FOLDER"))
	py.PythonVersion = pkg.FromStrPtr(env.Get("CC_PYTHON_VERSION"))
	py.PipRequirements = pkg.FromStrPtr(env.Get("CC_PIP_REQUIREMENTS_FILE"))

	py.Integrations = attributes.FromEnvIntegrations(ctx, env, py.Integrations, diags)
}

func (py Python) ToDeployment(gitAuth *http.BasicAuth) *application.Deployment {
	if py.Deployment == nil || py.Deployment.Repository.IsNull() {
		return nil
	}

	d := &application.Deployment{
		Repository:    py.Deployment.Repository.ValueString(),
		Commit:        py.Deployment.Commit.ValueStringPointer(),
		CleverGitAuth: gitAuth,
	}

	if !py.Deployment.BasicAuthentication.IsNull() && !py.Deployment.BasicAuthentication.IsUnknown() {
		// Expect validation to be done in the schema valisation step
		userPass := py.Deployment.BasicAuthentication.ValueString()
		splits := strings.SplitN(userPass, ":", 2)
		d.Username = &splits[0]
		d.Password = &splits[1]
	}

	return d
}
