package static

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"go.clever-cloud.com/terraform-provider/pkg/resources/application"
)

var _ resource.ResourceWithConfigValidators = (*ResourceStatic)(nil)

// ConfigValidators rejects the environment variables an attribute manages.
func (r *ResourceStatic) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{application.ReservedEnvironment(envVars...)}
}
