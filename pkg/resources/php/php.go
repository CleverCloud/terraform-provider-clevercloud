package php

import (
	"context"

	"github.com/go-git/go-git/v5/plumbing/transport/http"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"go.clever-cloud.dev/client"
)

type ResourcePHP struct {
	cc      *client.Client
	org     string
	gitAuth *http.BasicAuth
}

func NewResourcePHP() resource.Resource {
	return &ResourcePHP{}
}

func (r *ResourcePHP) Metadata(ctx context.Context, req resource.MetadataRequest, res *resource.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_php"
}
