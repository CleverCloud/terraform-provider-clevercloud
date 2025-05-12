//go:generate tfplugindocs
package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/getsentry/sentry-go"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/provider/impl"
)

// these will be set by the goreleaser configuration
// to appropriate values for the compiled binary.
var version string = "dev"

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers")
	flag.Parse()

	ctx := context.Background()
	opts := providerserver.ServeOpts{
		Address:         "registry.terraform.io/hashicorp/clevercloud",
		Debug:           debug,
		ProtocolVersion: 6,
	}

	if os.Getenv("CC_TELEMETRY") == "true" {
		pkg.SetupSentry()

		defer func() {
			if r := recover(); r != nil {
				ev := sentry.NewEvent()
				ev.Message = "panic in provider"
				ev.Extra["panic"] = fmt.Sprintf("%+v", r)
				sentry.CaptureEvent(ev)
			}
		}()
	}

	err := providerserver.Serve(ctx, impl.New(version), opts)
	if err != nil {
		panic(err.Error())
	}
}
