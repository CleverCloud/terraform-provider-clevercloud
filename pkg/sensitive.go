package pkg

import (
	"os"
	"strconv"
	"strings"
)

// BypassSensitiveImportsEnvVar names the variable that turns off the redaction
// of sensitive resource attributes for one provider process.
const BypassSensitiveImportsEnvVar = "BYPASS_SENSITIVE_IMPORTS"

// BypassSensitiveImports reports whether the practitioner asked that resource
// attributes stop being reported as sensitive.
//
// Terraform never writes a sensitive value into the configuration it generates,
// so `terraform plan -generate-config-out` emits `environment = null` for an
// imported application, and the same for every other sensitive attribute it
// could have written. Applying that file then deletes what it could not carry.
// Setting this variable for the one command that generates the configuration
// makes Terraform write those values out instead.
//
// The provider's own credentials are deliberately left redacted: configuration
// generation never writes a provider block, so unmasking them would leak for no
// gain.
//
// The schema is built once per provider process, so the answer is fixed at
// start-up and holds for the whole run — including, while it is set, every plan
// the process prints. Which variables are secrets and belong in a secret
// manager rather than in the configuration is the practitioner's call.
//
// Anything strconv.ParseBool accepts is accepted, plus yes and no, and the
// value is read case-insensitively: a variable a human types by hand deserves
// to work whichever spelling they reach for. Anything else reads as false.
func BypassSensitiveImports() bool {
	value := strings.ToLower(strings.TrimSpace(os.Getenv(BypassSensitiveImportsEnvVar)))

	switch value {
	case "yes":
		return true
	case "no":
		return false
	}

	bypass, err := strconv.ParseBool(value)

	return err == nil && bypass
}
