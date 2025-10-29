package application

import (
	"context"
	"maps"
	"strings"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"go.clever-cloud.com/terraform-provider/pkg"
)

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
				MarkdownDescription: "Support multiple syntax like `refs/heads/[BRANCH]` or `[COMMIT]`, in most of the case, you can use `refs/heads/master`",
				Validators: []validator.String{
					pkg.NewValidator(
						"if reference (not commit hash) is provided test it's syntax",
						func(ctx context.Context, req validator.StringRequest, res *validator.StringResponse) {
							if req.ConfigValue.IsNull() || !strings.Contains(req.ConfigValue.ValueString(), "/") {
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

func WithBlockRuntimeCommons(runtimeSpecifics map[string]schema.Block) map[string]schema.Block {
	m := map[string]schema.Block{}

	maps.Copy(m, blocks)
	maps.Copy(m, runtimeSpecifics)

	return m
}
