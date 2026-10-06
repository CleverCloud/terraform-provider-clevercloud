package fsbucket_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg/resources/database/fsbucket"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

func apiAddon() *tmp.AddonResponse {
	return &tmp.AddonResponse{
		ID: "addon_1", Name: "tf-test-fsbucket", RealID: "fsbucket_1", Region: "par",
		Plan: tmp.AddonPlan{Slug: "s"},
	}
}

func apiEnv() tmp.EnvVars {
	return tmp.EnvVars{
		{Name: "BUCKET_HOST", Value: "host.services.clever-cloud.com"},
		{Name: "BUCKET_FTP_USERNAME", Value: "ftp-user"},
		{Name: "BUCKET_FTP_PASSWORD", Value: "ftp-secret"},
	}
}

func TestFSBucketFromAPI_PopulatesEveryAttribute(t *testing.T) {
	var diags diag.Diagnostics
	state := &fsbucket.FSBucket{}

	state.
		FromAddon(t.Context(), apiAddon(), &diags).
		FromEnv(t.Context(), apiEnv(), &diags)

	if diags.HasError() {
		t.Fatalf("mapping reported %v", diags.Errors())
	}
	for name, value := range map[string]types.String{
		"name": state.Name, "region": state.Region, "host": state.Host,
		"ftp_username": state.FTPUsername, "ftp_password": state.FTPPassword,
	} {
		if value.IsNull() || value.IsUnknown() {
			t.Errorf("%s is %v, want a value", name, value)
		}
	}
}

func TestFSBucketFromAPI_PreservesWhatTheAPIDoesNotReturn(t *testing.T) {
	var diags diag.Diagnostics
	state := &fsbucket.FSBucket{FTPPassword: types.StringValue("kept-secret")}

	state.FromAddon(t.Context(), apiAddon(), &diags)

	if state.FTPPassword.ValueString() != "kept-secret" {
		t.Error("the add-on view carries no credential and must not clear one")
	}
}

func TestFSBucketFromAPI_NilPayloadIsANoOp(t *testing.T) {
	var diags diag.Diagnostics
	state := &fsbucket.FSBucket{Host: types.StringValue("kept-host")}

	state.
		FromAddon(t.Context(), nil, &diags).
		FromEnv(t.Context(), nil, &diags)

	if diags.HasError() {
		t.Fatalf("a nil payload must not report an error, got %v", diags.Errors())
	}
	if state.Host.ValueString() != "kept-host" {
		t.Error("host must be preserved")
	}
}
