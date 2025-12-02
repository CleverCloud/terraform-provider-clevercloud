package application

import (
	"context"

	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"go.clever-cloud.com/terraform-provider/pkg/provider"
)

// RuntimeResource interface defines methods required by resources to use generic CRUD operations
type RuntimeResource interface {
	provider.Provider
	GetVariantSlug() string
}

// RuntimePlan interface defines methods required by plan types to use generic CRUD operations
type RuntimePlan interface {
	VHostsAsStrings(ctx context.Context, diags *diag.Diagnostics) []string
	DependenciesAsString(ctx context.Context, diags *diag.Diagnostics) []string
	ToEnv(ctx context.Context, diags *diag.Diagnostics) map[string]string
	ToDeployment(auth *http.BasicAuth) *Deployment
	GetRuntimePtr() *Runtime
	FromEnv(ctx context.Context, env map[string]string, diags *diag.Diagnostics)
}
