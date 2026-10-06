package oauth_consumer_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	oauthconsumer "go.clever-cloud.com/terraform-provider/pkg/resources/oauth_consumer"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

func apiConsumer() *tmp.OAuthConsumerResponse {
	return &tmp.OAuthConsumerResponse{
		Name: "tf-test-consumer", Description: "a description",
		BaseURL: "https://example.com", LogoURL: "https://example.com/logo.png",
		WebsiteURL: "https://example.com/site",
		Rights:     tmp.OAuthConsumerRightsResponse{AccessOrganisations: true},
	}
}

func TestOAuthConsumerFromAPI_PopulatesEveryAttribute(t *testing.T) {
	var diags diag.Diagnostics
	state := &oauthconsumer.OAuthConsumer{}

	state.FromConsumer(t.Context(), apiConsumer(), &diags)
	state.FromSecret(t.Context(), &tmp.OAuthConsumerSecretResponse{Secret: "secret"}, &diags)

	if diags.HasError() {
		t.Fatalf("mapping reported %v", diags.Errors())
	}
	for name, value := range map[string]types.String{
		"name": state.Name, "description": state.Description,
		"base_url": state.BaseURL, "logo_url": state.LogoURL,
		"website_url": state.WebsiteURL, "secret": state.Secret,
	} {
		if value.IsNull() || value.IsUnknown() {
			t.Errorf("%s is %v, want a value", name, value)
		}
	}
	if state.Rights.IsNull() {
		t.Error("rights must be set")
	}
}

// The secret comes from its own endpoint, so the consumer view must not clear it.
func TestOAuthConsumerFromAPI_PreservesWhatTheAPIDoesNotReturn(t *testing.T) {
	var diags diag.Diagnostics
	state := &oauthconsumer.OAuthConsumer{Secret: types.StringValue("kept-secret")}

	state.FromConsumer(t.Context(), apiConsumer(), &diags)

	if state.Secret.ValueString() != "kept-secret" {
		t.Error("the consumer view has no secret and must not clear it")
	}
}

func TestOAuthConsumerFromAPI_NilPayloadIsANoOp(t *testing.T) {
	var diags diag.Diagnostics
	state := &oauthconsumer.OAuthConsumer{Name: types.StringValue("kept-name")}

	state.FromConsumer(t.Context(), nil, &diags)
	state.FromSecret(t.Context(), nil, &diags)

	if diags.HasError() {
		t.Fatalf("a nil payload must not report an error, got %v", diags.Errors())
	}
	if state.Name.ValueString() != "kept-name" {
		t.Error("name must be preserved")
	}
}
