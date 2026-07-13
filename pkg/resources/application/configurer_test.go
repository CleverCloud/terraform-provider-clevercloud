package application_test

import (
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"go.clever-cloud.com/terraform-provider/pkg/resources/application/nodejs"
)

var (
	_ resource.ResourceWithValidateConfig = &nodejs.ResourceNodeJS{}
	_ resource.ResourceWithConfigure      = &nodejs.ResourceNodeJS{}
	_ resource.ResourceWithImportState    = &nodejs.ResourceNodeJS{}
)
