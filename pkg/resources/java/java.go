package java

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
)

type ResourceJava struct {
	helper.Configurer
	profile string
}

func NewResourceJava(profile string) func() resource.Resource {
	return func() resource.Resource {
		return &ResourceJava{profile: profile}
	}
}

func (r *ResourceJava) Metadata(ctx context.Context, req resource.MetadataRequest, res *resource.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_java"

	if r.profile != "" {
		res.TypeName = res.TypeName + "_" + r.profile
	}
}
