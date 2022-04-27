package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type NodeJS struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
	// TODO
}

func (r resourceNodejsType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			// customer provided
			"name":              {Type: types.StringType, Required: true},
			"description":       {Type: types.StringType, Required: true},
			"min_flavor":        {Type: types.StringType, Optional: true},
			"max_flavor":        {Type: types.StringType, Required: true},
			"region":            {Type: types.StringType, Required: true},
			"sticky_sessions":   {Type: types.BoolType, Optional: true},
			"build_flavor":      {Type: types.StringType, Optional: true},
			"redirect_https":    {Type: types.BoolType, Optional: true},
			"commit":            {Type: types.StringType, Optional: true, Description: "Support either full commit sha or tag"},
			"additional_vhosts": {Type: types.ListType{ElemType: types.StringType}, Computed: true},

			// provider
			"id":         {Type: types.StringType, Computed: true}, // for acceptance test retro compat
			"deploy_url": {Type: types.StringType, Computed: true},
			"vhost":      {Type: types.StringType, Computed: true}, // cleverapps one
		},
	}, nil
}
