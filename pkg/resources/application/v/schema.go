package v

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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg"
)

type V struct {
	application.Runtime
	Binary           types.String `tfsdk:"binary"`
	DevelopmentBuild types.Bool   `tfsdk:"development_build"`
}

//go:embed doc.md
var vDoc string

func (r ResourceV) Schema(ctx context.Context, req resource.SchemaRequest, res *resource.SchemaResponse) {

	res.Schema = schema.Schema{
		Version:             0,
		MarkdownDescription: vDoc,
		Attributes: application.WithRuntimeCommons(map[string]schema.Attribute{
			"binary": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The name of the output binary file. Default: `${APP_HOME}/v_bin_${APP_ID}`",
			},
			"development_build": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Set to true to compile without the `-prod` flag.",
			},
		}),
		Blocks: attributes.WithBlockRuntimeCommons(map[string]schema.Block{}),
	}
}

func (vapp V) toEnv(ctx context.Context, diags *diag.Diagnostics) map[string]string {
	env := map[string]string{}

	// do not use the real map since ElementAs can nullish it
	// https://github.com/hashicorp/terraform-plugin-framework/issues/698
	customEnv := map[string]string{}
	diags.Append(vapp.Environment.ElementsAs(ctx, &customEnv, false)...)
	if diags.HasError() {
		return env
	}
	env = pkg.Merge(env, customEnv)

	pkg.IfIsSetStr(vapp.Binary, func(s string) { env["CC_V_BINARY"] = s })
	pkg.IfIsSetB(vapp.DevelopmentBuild, func(devBuild bool) {
		if devBuild {
			env["ENVIRONMENT"] = "development"
		}
	})

	env = pkg.Merge(env, vapp.Hooks.ToEnv())

	return env
}

func (vapp V) toDeployment(gitAuth *http.BasicAuth) *application.Deployment {
	if vapp.Deployment == nil || vapp.Deployment.Repository.IsNull() {
		return nil
	}

	d := &application.Deployment{
		Repository:    vapp.Deployment.Repository.ValueString(),
		Commit:        vapp.Deployment.Commit.ValueStringPointer(),
		CleverGitAuth: gitAuth,
	}

	if !vapp.Deployment.BasicAuthentication.IsNull() && !vapp.Deployment.BasicAuthentication.IsUnknown() {
		// Expect validation to be done in the schema valisation step
		userPass := vapp.Deployment.BasicAuthentication.ValueString()
		splits := strings.SplitN(userPass, ":", 2)
		d.Username = &splits[0]
		d.Password = &splits[1]
	}

	return d
}
