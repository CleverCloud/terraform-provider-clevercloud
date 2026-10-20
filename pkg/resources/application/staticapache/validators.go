package staticapache

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"go.clever-cloud.com/terraform-provider/pkg/resources/application"
)

var _ resource.ResourceWithConfigValidators = (*ResourceStaticApache)(nil)

// ConfigValidators rejects the environment variables an attribute manages.
func (r *ResourceStaticApache) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{application.ReservedEnvironment(envVars...)}
}
