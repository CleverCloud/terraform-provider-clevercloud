package java

import (
	"context"
	_ "embed"
	"strings"

	"go.clever-cloud.com/terraform-provider/pkg/attributes"
	"go.clever-cloud.com/terraform-provider/pkg/resources/application"

	"maps"

	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg"
)

type Java struct {
	application.Runtime
	JavaVersion types.String `tfsdk:"java_version"`
}

type JavaV0 struct {
	application.RuntimeV0
	JavaVersion types.String `tfsdk:"java_version"`
}

//go:embed doc.md
var javaDoc string

func (r ResourceJava) Schema(ctx context.Context, req resource.SchemaRequest, res *resource.SchemaResponse) {
	res.Schema = schemaJava
}

var schemaJava = schema.Schema{
	Version:             1,
	MarkdownDescription: javaDoc,
	Attributes: application.WithRuntimeCommons(map[string]schema.Attribute{
		"java_version": schema.StringAttribute{
			Optional:    true,
			Description: "Choose the JVM version between 7 to 24 for OpenJDK or graalvm-ce for GraalVM 21.0.0.2 (based on OpenJDK 11.0).",
		},
	}),
	Blocks: attributes.WithBlockRuntimeCommons(map[string]schema.Block{}),
}

var schemaJavaV0 = schema.Schema{
	Version:             0,
	MarkdownDescription: javaDoc,
	Attributes: application.WithRuntimeCommonsV0(map[string]schema.Attribute{
		"java_version": schema.StringAttribute{
			Optional:    true,
			Description: "Choose the JVM version between 7 to 24 for OpenJDK or graalvm-ce for GraalVM 21.0.0.2 (based on OpenJDK 11.0).",
		},
	}),
	Blocks: attributes.WithBlockRuntimeCommons(map[string]schema.Block{}),
}

func (plan *Java) ToEnv(ctx context.Context, diags *diag.Diagnostics) map[string]string {
	env := map[string]string{}

	// do not use the real map since ElementAs can nullish it
	// https://github.com/hashicorp/terraform-plugin-framework/issues/698
	customEnv := map[string]string{}
	diags.Append(plan.Environment.ElementsAs(ctx, &customEnv, false)...)
	if diags.HasError() {
		return env
	}
	maps.Copy(env, customEnv)

	pkg.IfIsSetStr(plan.AppFolder, func(s string) { env["APP_FOLDER"] = s })
	pkg.IfIsSetStr(plan.JavaVersion, func(s string) { env["CC_JAVA_VERSION"] = s })
	return env
}

func (java *Java) FromEnv(ctx context.Context, env map[string]string, diags *diag.Diagnostics) {
	if val, ok := env["APP_FOLDER"]; ok {
		java.AppFolder = pkg.FromStr(val)
	}
	if val, ok := env["CC_JAVA_VERSION"]; ok {
		java.JavaVersion = pkg.FromStr(val)
	}
}

func (java *Java) ToDeployment(gitAuth *http.BasicAuth) *application.Deployment {
	if java.Deployment == nil || java.Deployment.Repository.IsNull() {
		return nil
	}

	d := &application.Deployment{
		Repository:    java.Deployment.Repository.ValueString(),
		Commit:        java.Deployment.Commit.ValueStringPointer(),
		CleverGitAuth: gitAuth,
	}

	if !java.Deployment.BasicAuthentication.IsNull() && !java.Deployment.BasicAuthentication.IsUnknown() {
		// Expect validation to be done in the schema valisation step
		userPass := java.Deployment.BasicAuthentication.ValueString()
		splits := strings.SplitN(userPass, ":", 2)
		d.Username = &splits[0]
		d.Password = &splits[1]
	}

	return d
}
