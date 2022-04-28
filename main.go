//go:generate tfplugindocs
package main

import (
	"context"
	"flag"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"go.clever-cloud.com/terraform-provider/pkg/provider"
)

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers")
	flag.Parse()

	ctx := context.Background()
	opts := tfsdk.ServeOpts{
		Name:  "clevercloud",
		Debug: debug,
	}

	err := tfsdk.Serve(ctx, provider.New, opts)
	if err != nil {
		panic(err.Error())
	}
}
