package java

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

type Java struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	MinInstanceCount types.Int64  `tfsdk:"min_instance_count"`
	MaxInstanceCount types.Int64  `tfsdk:"max_instance_count"`
	SmallestFlavor   types.String `tfsdk:"smallest_flavor"`
	BiggestFlavor    types.String `tfsdk:"biggest_flavor"`
	BuildFlavor      types.String `tfsdk:"build_flavor"`
	Region           types.String `tfsdk:"region"`
	StickySessions   types.Bool   `tfsdk:"sticky_sessions"`
	RedirectHTTPS    types.Bool   `tfsdk:"redirect_https"`
	VHost            types.String `tfsdk:"vhost"`
	AdditionalVHosts types.List   `tfsdk:"additional_vhosts"`
	DeployURL        types.String `tfsdk:"deploy_url"`
	Deployment       *Deployment  `tfsdk:"deployment"`

	// Env
	AppFolder   types.String `tfsdk:"app_folder"`
	Environment types.Map    `tfsdk:"environment"`

	// Java related
	JavaVersion types.String `tfsdk:"java_version"`
}

type Deployment struct {
	Repository types.String `tfsdk:"repository"`
	Commit     types.String `tfsdk:"commit"`
}

//go:embed resource_java.md
var javaDoc string

func (r ResourceJava) Schema(ctx context.Context, req resource.SchemaRequest, res *resource.SchemaResponse) {
	res.Schema = schema.Schema{
		Version:             0,
		MarkdownDescription: javaDoc,
		Attributes: attributes.WithRuntimeCommons(map[string]schema.Attribute{
			"java_version": schema.StringAttribute{
				Optional:    true,
				Description: "Choose the JVM version between 7 to 17 for OpenJDK or graalvm-ce for GraalVM 21.0.0.2 (based on OpenJDK 11.0).",
			},
		}),
		Blocks: map[string]schema.Block{
			"deployment": schema.SingleNestedBlock{ // TODO: factorize it
				Attributes: map[string]schema.Attribute{
					"repository": schema.StringAttribute{
						Optional:            true, // If "deployment" attribute is defined, then repository is required
						Description:         "",
						MarkdownDescription: "",
					},
					"commit": schema.StringAttribute{
						Optional:            true,
						Description:         "Support either '<branch>:<SHA>' or '<tag>'",
						MarkdownDescription: "Deploy application on the given commit/tag",
					},
				},
			},
		},
	}
}

// https://developer.hashicorp.com/terraform/plugin/framework/resources/state-upgrade#implementing-state-upgrade-support
func (plan *Java) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{}
}

func (plan *Java) toEnv(ctx context.Context, diags diag.Diagnostics) map[string]string {
	env := map[string]string{}

	// do not use the real map since ElementAs can nullish it
	// https://github.com/hashicorp/terraform-plugin-framework/issues/698
	customEnv := map[string]string{}
	diags.Append(plan.Environment.ElementsAs(ctx, &customEnv, false)...)
	if diags.HasError() {
		return env
	}
	for k, v := range customEnv {
		env[k] = v
	}

	pkg.IfIsSet(plan.AppFolder, func(s string) { env["APP_FOLDER"] = s })
	pkg.IfIsSet(plan.JavaVersion, func(s string) { env["CC_JAVA_VERSION"] = s })
	return env
}

func (java *Java) toDeployment() *application.Deployment {
	if java.Deployment == nil || java.Deployment.Repository.IsNull() {
		return nil
	}

	return &application.Deployment{
		Repository: java.Deployment.Repository.ValueString(),
		Commit:     java.Deployment.Commit.ValueStringPointer(),
	}
}
