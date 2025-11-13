package attributes

import (
	"context"
	"maps"
	"strings"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg"
)

// used to identify when the repo deployment is handled by a Github hook
// In this case, we must not deploy with Terraform
const GITHUB_COMMIT_PREFIX = "github_hook"

// Deployment block
type Deployment struct {
	Repository          types.String `tfsdk:"repository"`
	Commit              types.String `tfsdk:"commit"`
	BasicAuthentication types.String `tfsdk:"authentication_basic"`
}

// Hooks block
type Hooks struct {
	PreBuild   types.String `tfsdk:"pre_build"`
	PostBuild  types.String `tfsdk:"post_build"`
	PreRun     types.String `tfsdk:"pre_run"`
	RunSucceed types.String `tfsdk:"run_succeed"`
	RunFailed  types.String `tfsdk:"run_failed"`
}

var blocks = map[string]schema.Block{
	"deployment": schema.SingleNestedBlock{
		Attributes: map[string]schema.Attribute{
			"repository": schema.StringAttribute{
				Optional:            true, // If "deployment" attribute is defined, then repository is required
				Description:         "The repository URL to deploy, can be 'https://...', 'file://...'",
				MarkdownDescription: "The repository URL to deploy, can be 'https://...', 'file://...'",
			},
			"commit": schema.StringAttribute{
				Optional:            true,
				Description:         "The git reference you want to deploy",
				MarkdownDescription: "Support multiple syntax like `refs/heads/[BRANCH]`, `github_hook` or `[COMMIT]`, when using the special value `github_hook`, we will link the application to the Github repository",
				Validators: []validator.String{
					pkg.NewValidator(
						"if reference (not commit hash) is provided test it's syntax",
						func(ctx context.Context, req validator.StringRequest, res *validator.StringResponse) {
							if req.ConfigValue.IsNull() || !strings.Contains(req.ConfigValue.ValueString(), "/") {
								return
							}

							if req.ConfigValue.ValueString() == GITHUB_COMMIT_PREFIX {
								return
							}

							ref := plumbing.ReferenceName(req.ConfigValue.ValueString())
							if err := ref.Validate(); err != nil {
								res.Diagnostics.AddAttributeError(
									req.Path,
									"invalid Git reference",
									err.Error(),
								)
							}
						}),
				},
			},
			"authentication_basic": schema.StringAttribute{
				Sensitive: true,
				Optional:  true,
				// TODO: investigate behaviour (not available in plan, but available in config ?)
				//WriteOnly:           true,
				MarkdownDescription: "user ans password ':' separated, (PersonalAccessToken in Github case)",
				Validators:          []validator.String{UserPasswordInput},
			},
		},
	},
	"hooks": schema.SingleNestedBlock{
		Attributes: map[string]schema.Attribute{
			"pre_build": schema.StringAttribute{
				Optional:            true,
				Description:         "This hook is ran before the dependencies are fetched. If it fails, the deployment fails",
				MarkdownDescription: "[CC_PRE_BUILD_HOOK](https://www.clever.cloud/developers/doc/develop/build-hooks/#pre-build)",
			},
			"post_build": schema.StringAttribute{
				Optional:            true,
				Description:         "This hook is ran after the project is built, and before the cache archive is generated. If it fails, the deployment fails",
				MarkdownDescription: "[CC_POST_BUILD_HOOK](https://www.clever.cloud/developers/doc/develop/build-hooks/#post-build)",
			},
			"pre_run": schema.StringAttribute{
				Optional:            true,
				Description:         "This hook is ran before the application is started, but after the cache archive has been generated. If it fails, the deployment fails.",
				MarkdownDescription: "[CC_PRE_RUN_HOOK](https://www.clever.cloud/developers/doc/develop/build-hooks/#pre-run)",
			},
			"run_succeed": schema.StringAttribute{
				Optional:            true,
				Description:         "This hook are ran once the application has started. Their failure doesn't cause the deployment to fail.",
				MarkdownDescription: "[CC_RUN_SUCCEEDED_HOOK](https://www.clever.cloud/developers/doc/develop/build-hooks/#run-successfail)",
			},
			"run_failed": schema.StringAttribute{
				Optional:            true,
				Description:         "This hook are ran once the application has failed starting.",
				MarkdownDescription: "[CC_RUN_FAILED_HOOK](https://www.clever.cloud/developers/doc/develop/build-hooks/#run-successfail)",
			},
		},
	},
}

var UserPasswordInput = pkg.NewStringValidator("format must be 'user:password'", func(ctx context.Context, req validator.StringRequest, res *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	v := req.ConfigValue.ValueString()
	splits := strings.SplitN(v, ":", 2)
	if len(splits) != 2 {
		res.Diagnostics.AddError("invalid authenication value", "expect user:password")
		return
	}
	if splits[0] == "" && splits[1] == "" {
		res.Diagnostics.AddError("invalid authenication value", "user and password cannot be both empty")
	}
})

func WithBlockRuntimeCommons(runtimeSpecifics map[string]schema.Block) map[string]schema.Block {
	m := map[string]schema.Block{}
	maps.Copy(m, blocks)
	maps.Copy(m, runtimeSpecifics)
	return m
}

func (hooks *Hooks) ToEnv() map[string]string {
	m := map[string]string{}

	if hooks == nil {
		return m
	}

	pkg.IfIsSetStr(hooks.PreBuild, func(script string) { m["CC_PRE_BUILD_HOOK"] = script })
	pkg.IfIsSetStr(hooks.PostBuild, func(script string) { m["CC_POST_BUILD_HOOK"] = script })
	pkg.IfIsSetStr(hooks.PreRun, func(script string) { m["CC_PRE_RUN_HOOK"] = script })
	pkg.IfIsSetStr(hooks.RunFailed, func(script string) { m["CC_RUN_FAILED_HOOK"] = script })
	pkg.IfIsSetStr(hooks.RunSucceed, func(script string) { m["CC_RUN_SUCCEEDED_HOOK"] = script })

	return m
}
