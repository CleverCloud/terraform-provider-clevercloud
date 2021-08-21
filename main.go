package main

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"

	"github.com/clevercloud/terraform-provider-clevercloud/clevercloud"
)

func main() {
	tfsdk.Serve(context.Background(), clevercloud.New, tfsdk.ServeOpts{
		Name: "clevercloud",
	})
}
