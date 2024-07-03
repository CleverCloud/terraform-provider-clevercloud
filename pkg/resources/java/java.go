package java

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"go.clever-cloud.dev/client"
)

type ResourceJava struct {
	// war /
	profile string
	cc      *client.Client
	org     string
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

// Convert a profile into product name
func (r *ResourceJava) toProductName() string {
	m := map[string]string{
		"war": "Java + WAR",
	}

	return m[r.profile]
}
