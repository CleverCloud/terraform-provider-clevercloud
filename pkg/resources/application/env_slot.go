package application

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	helpermaps "github.com/miton18/helper/maps"
	"go.clever-cloud.com/terraform-provider/pkg/attributes"
)

// withholdUserManagedEnv hides from the runtime, hook and integration readers
// every variable the practitioner declares in `environment`, so refresh cannot
// move it to its dedicated attribute and leave a permanent diff behind.
// Returned entries must be put back into env once those readers have run.
func withholdUserManagedEnv(env *helpermaps.Map[string, string], stateEnv types.Map) map[string]string {
	withheld := map[string]string{}

	if stateEnv.IsNull() || stateEnv.IsUnknown() {
		return withheld
	}

	for key := range stateEnv.Elements() {
		if value := env.PopPtr(key); value != nil {
			withheld[key] = *value
		}
	}

	return withheld
}

// ValidateEnvSlots reports variables put in `environment` although a dedicated
// attribute exists for them. Two cases, deliberately two distinct messages:
// declared in both slots the plan cannot settle, declared in `environment` alone
// it settles but gives up the attribute's typing and validation.
func ValidateEnvSlots[P any, T interface {
	*P
	RuntimePlan
}](ctx context.Context, plan T, diags *diag.Diagnostics) {
	runtime := plan.GetRuntimePtr()
	if runtime.Environment.IsNull() || runtime.Environment.IsUnknown() {
		return
	}

	// the check is best effort: an undecodable plan is left to the CRUD operations to report
	probe := diag.Diagnostics{}

	// emptying `environment` leaves ToEnv with the attribute-derived variables only
	configured := runtime.Environment
	defer func() { runtime.Environment = configured }()

	runtime.Environment = types.MapValueMust(types.StringType, map[string]attr.Value{})
	derived := plan.ToEnv(ctx, &probe)
	if probe.HasError() {
		return
	}

	backed := attributeBackedKeys[P, T](ctx, configured, &probe)

	for key := range configured.Elements() {
		_, isDerived := derived[key]

		switch {
		case isDerived:
			diags.AddAttributeWarning(
				path.Root("environment").AtMapKey(key),
				"Variable also managed by a dedicated attribute",
				key+" is set in `environment` and by the attribute dedicated to it. "+
					"The attribute wins when Terraform writes to Clever Cloud, `environment` wins on refresh, "+
					"so the plan cannot settle. Prefer the dedicated attribute and remove the variable from `environment`.",
			)
		case backed[key]:
			diags.AddAttributeWarning(
				path.Root("environment").AtMapKey(key),
				"Variable has a dedicated attribute",
				key+" is set in `environment` while a dedicated attribute exists for it. "+
					"Both reach Clever Cloud, but the attribute is typed and validated where `environment` "+
					"is a plain string map. Prefer the dedicated attribute.",
			)
		}
	}
}

// attributeBackedKeys returns the keys of `environment` that a dedicated attribute
// would claim on refresh. FromEnv pops what it consumes, so whatever survives has no
// attribute behind it. Values are irrelevant — every reader keys off presence — which
// is what keeps the check working when the practitioner interpolates an unknown value.
// FromEnv writes to its receiver, hence the scratch plan rather than the real one.
func attributeBackedKeys[P any, T interface {
	*P
	RuntimePlan
}](ctx context.Context, configured types.Map, diags *diag.Diagnostics) map[string]bool {
	env := helpermaps.NewMap(map[string]string{})
	for key := range configured.Elements() {
		env.Set(key, "")
	}

	scratch := T(new(P))
	scratch.FromEnv(ctx, env, diags)
	attributes.FromEnvHooks(env, nil)

	backed := map[string]bool{}
	for key := range configured.Elements() {
		if !env.Has(key) {
			backed[key] = true
		}
	}

	return backed
}
