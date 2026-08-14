package application

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	helpermaps "github.com/miton18/helper/maps"
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
