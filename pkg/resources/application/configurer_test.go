package application_test

import (
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"go.clever-cloud.com/terraform-provider/pkg/resources/application/nodejs"
)

// Configurer's framework interfaces are inherited by every runtime resource
// through embedding, and their signatures do not depend on the type parameter:
// asserting one representative instantiation (nodejs) proves the wiring for
// all runtimes, notably the shared ValidateConfig.
var (
	_ resource.ResourceWithValidateConfig = &nodejs.ResourceNodeJS{}
	_ resource.ResourceWithConfigure      = &nodejs.ResourceNodeJS{}
	_ resource.ResourceWithImportState    = &nodejs.ResourceNodeJS{}
)
