package redis_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg/resources/database/redis"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

func apiAddon() *tmp.AddonResponse {
	return &tmp.AddonResponse{
		ID: "addon_1", Name: "tf-test-redis", RealID: "redis_1", Region: "par",
		CreationDate: 1759000000000, Plan: tmp.AddonPlan{Slug: "m_mono"},
	}
}

func apiEnv() tmp.EnvVars {
	return tmp.EnvVars{
		{Name: "REDIS_HOST", Value: "host.services.clever-cloud.com"},
		{Name: "REDIS_PORT", Value: "6379"},
		{Name: "REDIS_PASSWORD", Value: "secret"},
	}
}

func TestRedisFromAPI_PopulatesEveryAttribute(t *testing.T) {
	var diags diag.Diagnostics
	state := &redis.Redis{}

	state.FromAddon(t.Context(), apiAddon(), &diags)
	state.FromEnv(t.Context(), apiEnv(), &diags)

	if diags.HasError() {
		t.Fatalf("mapping reported %v", diags.Errors())
	}
	for name, value := range map[string]types.String{
		"name": state.Name, "plan": state.Plan, "region": state.Region,
		"host": state.Host, "token": state.Token,
	} {
		if value.IsNull() || value.IsUnknown() {
			t.Errorf("%s is %v, want a value", name, value)
		}
	}
	if state.Port.ValueInt64() != 6379 {
		t.Errorf("port = %d, want 6379", state.Port.ValueInt64())
	}
	// creation_date is Computed and used to be left unset by Read, so an import
	// kept it null forever.
	if state.CreationDate.IsNull() {
		t.Error("creation_date must be set")
	}
}

func TestRedisFromAPI_PreservesWhatTheAPIDoesNotReturn(t *testing.T) {
	var diags diag.Diagnostics
	state := &redis.Redis{
		Host:  types.StringValue("kept-host"),
		Token: types.StringValue("kept-token"),
	}

	state.FromAddon(t.Context(), apiAddon(), &diags)

	if state.Host.ValueString() != "kept-host" || state.Token.ValueString() != "kept-token" {
		t.Error("the add-on view carries no connection detail and must not clear one")
	}
}

// A port the API answers as something other than a number is reported as absent
// rather than failing the whole read.
func TestRedisFromAPI_UnparseablePortIsNull(t *testing.T) {
	var diags diag.Diagnostics
	state := &redis.Redis{}

	state.FromEnv(t.Context(), tmp.EnvVars{{Name: "REDIS_PORT", Value: "not-a-port"}}, &diags)

	if diags.HasError() {
		t.Fatalf("an unparseable port must not fail the read, got %v", diags.Errors())
	}
	if !state.Port.IsNull() {
		t.Errorf("port = %v, want null", state.Port)
	}
}

func TestRedisFromAPI_NilPayloadIsANoOp(t *testing.T) {
	var diags diag.Diagnostics
	state := &redis.Redis{Host: types.StringValue("kept-host")}

	state.FromAddon(t.Context(), nil, &diags)
	state.FromEnv(t.Context(), nil, &diags)

	if diags.HasError() {
		t.Fatalf("a nil payload must not report an error, got %v", diags.Errors())
	}
	if state.Host.ValueString() != "kept-host" {
		t.Error("host must be preserved")
	}
}
