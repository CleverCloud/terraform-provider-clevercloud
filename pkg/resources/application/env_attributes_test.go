package application

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"go.clever-cloud.dev/sdk/app"
)

func validateEnvironment(t *testing.T, env tftypes.Value) []string {
	t.Helper()
	ctx := context.Background()
	s := schema.Schema{Attributes: map[string]schema.Attribute{
		"environment": schema.MapAttribute{Optional: true, ElementType: types.StringType},
	}}
	config := tfsdk.Config{
		Schema: s,
		Raw: tftypes.NewValue(s.Type().TerraformType(ctx), map[string]tftypes.Value{
			"environment": env,
		}),
	}

	res := &resource.ValidateConfigResponse{}
	ReservedEnvironment(map[string]string{app.CC_NODE_VERSION: "node_version"}).
		ValidateResource(ctx, resource.ValidateConfigRequest{Config: config}, res)

	var details []string
	for _, d := range res.Diagnostics.Errors() {
		details = append(details, d.Detail())
	}
	return details
}

func envValue(values map[string]tftypes.Value) tftypes.Value {
	return tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, values)
}

func TestReservedEnvironment(t *testing.T) {
	errs := validateEnvironment(t, envValue(map[string]tftypes.Value{
		app.CC_NODE_VERSION:   tftypes.NewValue(tftypes.String, "22"),
		app.APP_FOLDER:        tftypes.NewValue(tftypes.String, "web"),
		app.CC_PRE_BUILD_HOOK: tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		"MY_VARIABLE":         tftypes.NewValue(tftypes.String, "free"),
	}))

	want := []string{`APP_FOLDER is managed by the "app_folder"`, `CC_NODE_VERSION is managed by the "node_version"`, `CC_PRE_BUILD_HOOK is managed by the "hooks.pre_build"`}
	if len(errs) != len(want) {
		t.Fatalf("got %d errors %v, want %d", len(errs), errs, len(want))
	}
	for i, prefix := range want {
		if !strings.HasPrefix(errs[i], prefix) {
			t.Errorf("error %d = %q, want it to start with %q", i, errs[i], prefix)
		}
	}
}

func TestReservedEnvironment_nullOrUnknown(t *testing.T) {
	for name, env := range map[string]tftypes.Value{
		"null":    tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, nil),
		"unknown": tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, tftypes.UnknownValue),
	} {
		if errs := validateEnvironment(t, env); len(errs) != 0 {
			t.Errorf("%s environment: got errors %v", name, errs)
		}
	}
}
