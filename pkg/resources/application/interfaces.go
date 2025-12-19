package application

import (
	"context"

	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg/provider"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
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

// AppResponseProvider abstracts access to the underlying AppResponse for response mapping
type AppResponseProvider interface {
	GetApp() *tmp.AppResponse
	GetBuildFlavor() types.String
}
