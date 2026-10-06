package application_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg/resources/application"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

// stubResponse is the AppResponseProvider the generic CRUD hands to FromApp.
type stubResponse struct {
	app         *tmp.AppResponse
	buildFlavor types.String
}

func (s stubResponse) GetApp() *tmp.AppResponse     { return s.app }
func (s stubResponse) GetBuildFlavor() types.String { return s.buildFlavor }

func apiApp() *tmp.AppResponse {
	app := &tmp.AppResponse{
		Name:           "tf-test-app",
		Description:    "a description",
		Zone:           "par",
		DeployURL:      "https://push.clever-cloud.com/app_1.git",
		StickySessions: true,
		ForceHTTPS:     "ENABLED",
	}
	app.Instance.MinInstances = 1
	app.Instance.MaxInstances = 3
	app.Instance.MinFlavor.Name = "nano"
	app.Instance.MaxFlavor.Name = "XS"
	return app
}

func TestRuntimeFromAPI_PopulatesEveryAttribute(t *testing.T) {
	var diags diag.Diagnostics
	runtime := &application.Runtime{}

	runtime.FromApp(t.Context(), stubResponse{
		app:         apiApp(),
		buildFlavor: types.StringValue("M"),
	}, &diags)

	if diags.HasError() {
		t.Fatalf("mapping reported %v", diags.Errors())
	}
	for name, value := range map[string]types.String{
		"name": runtime.Name, "description": runtime.Description,
		"region": runtime.Region, "deploy_url": runtime.DeployURL,
		"smallest_flavor": runtime.SmallestFlavor, "biggest_flavor": runtime.BiggestFlavor,
		"build_flavor": runtime.BuildFlavor,
	} {
		if value.IsNull() || value.IsUnknown() {
			t.Errorf("%s is %v, want a value", name, value)
		}
	}
	if runtime.MinInstanceCount.ValueInt64() != 1 || runtime.MaxInstanceCount.ValueInt64() != 3 {
		t.Error("the instance counts must be mapped")
	}
	// Both are Computed with a false default: a null here is a diff on the next plan.
	for name, value := range map[string]types.Bool{
		"sticky_sessions": runtime.StickySessions, "redirect_https": runtime.RedirectHTTPS,
	} {
		if value.IsNull() || value.IsUnknown() {
			t.Errorf("%s is %v: a Computed attribute must never be null after a read", name, value)
		}
		if !value.ValueBool() {
			t.Errorf("%s = false, want true", name)
		}
	}
}

// ForceHTTPS is an enum on the API side and a bool in the schema.
func TestRuntimeFromAPI_RedirectHTTPSFollowsTheEnum(t *testing.T) {
	for enum, want := range map[string]bool{"ENABLED": true, "DISABLED": false, "": false} {
		t.Run(enum, func(t *testing.T) {
			var diags diag.Diagnostics
			runtime := &application.Runtime{}
			app := apiApp()
			app.ForceHTTPS = enum

			runtime.FromApp(t.Context(), stubResponse{app: app}, &diags)

			if runtime.RedirectHTTPS.ValueBool() != want {
				t.Errorf("redirect_https = %v, want %v", runtime.RedirectHTTPS, want)
			}
		})
	}
}

// A failed fetch hands the mapper no payload, and prior state must survive it.
func TestRuntimeFromAPI_NilPayloadIsANoOp(t *testing.T) {
	var diags diag.Diagnostics
	runtime := &application.Runtime{Name: types.StringValue("kept-name")}

	runtime.FromApp(t.Context(), nil, &diags)
	runtime.FromApp(t.Context(), stubResponse{app: nil}, &diags)

	if diags.HasError() {
		t.Fatalf("a nil payload must not report an error, got %v", diags.Errors())
	}
	if runtime.Name.ValueString() != "kept-name" {
		t.Errorf("name = %q, want it preserved", runtime.Name.ValueString())
	}
}
