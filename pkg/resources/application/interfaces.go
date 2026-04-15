package application

import (
	"context"

	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/miton18/helper/maps"
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
	FromEnv(ctx context.Context, env *maps.Map[string, string], diags *diag.Diagnostics)
}

// AppResponseProvider abstracts access to the underlying AppResponse for response mapping
type AppResponseProvider interface {
	GetApp() *tmp.AppResponse
	GetBuildFlavor() types.String
}

// VariantGuard is an optional interface implemented by runtimes that want a strict
// check at Read time: if the CC app's actual variant slug differs from the resource's
// expected slug, Read emits a clear error pointing the user to the right resource type.
// Used to protect users against resource renames (e.g. clevercloud_static was renamed
// to clevercloud_static_apache — a legacy state left on the new clevercloud_static
// runtime would silently mis-manage a static-apache app without this guard).
type VariantGuard interface {
	// MigrationHint is the remediation message shown when a variant mismatch is detected.
	// actualSlug is the slug returned by the CC API for the mis-managed app.
	MigrationHint(actualSlug string) string
}
