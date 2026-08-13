package impl

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.dev/client"
)

func TestResolveAPIToken_AttributeTakesPrecedenceOverEnv(t *testing.T) {
	t.Setenv("CLEVER_API_TOKEN", "token-from-env")

	config := ProviderData{APIToken: types.StringValue("token-from-attribute")}

	token, fromEnv := resolveAPIToken(config)
	if token != "token-from-attribute" {
		t.Errorf("expected the api_token attribute to win over the environment, got %q", token)
	}
	if fromEnv {
		t.Error("expected fromEnv to be false when the attribute is set")
	}
}

func TestResolveAPIToken_EnvFallback(t *testing.T) {
	t.Setenv("CLEVER_API_TOKEN", "token-from-env")

	config := ProviderData{APIToken: types.StringNull()}

	token, fromEnv := resolveAPIToken(config)
	if token != "token-from-env" {
		t.Errorf("expected the CLEVER_API_TOKEN environment variable to be used, got %q", token)
	}
	if !fromEnv {
		t.Error("expected fromEnv to be true when the token comes from the environment")
	}
}

func TestResolveAPIToken_Empty(t *testing.T) {
	t.Setenv("CLEVER_API_TOKEN", "")

	config := ProviderData{APIToken: types.StringNull()}

	token, _ := resolveAPIToken(config)
	if token != "" {
		t.Errorf("expected no token, got %q", token)
	}
}

func TestBearerClientOptions_SendsBearerAuthorization(t *testing.T) {
	var gotAuthorization string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuthorization = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte(`{}`)); err != nil {
			t.Errorf("failed to write response: %s", err)
		}
	}))
	defer server.Close()

	cc := client.New(bearerClientOptions(server.URL, "my-api-token")...)

	res := client.Get[map[string]any](context.Background(), cc, "/v2/self")
	if res.HasError() {
		t.Fatalf("unexpected error: %s", res.Error())
	}

	if gotAuthorization != "Bearer my-api-token" {
		t.Errorf("expected 'Bearer my-api-token' Authorization header, got %q", gotAuthorization)
	}
}

func TestBearerEndpoint_DefaultsToBridge(t *testing.T) {
	if got := bearerEndpoint(""); got != client.BRIDGE_API_ENDPOINT {
		t.Errorf("expected the API bridge endpoint by default, got %q", got)
	}
}

func TestBearerEndpoint_KeepsExplicitEndpoint(t *testing.T) {
	if got := bearerEndpoint("https://example.com"); got != "https://example.com" {
		t.Errorf("expected the explicit endpoint to be kept, got %q", got)
	}
}

func TestIsSetStr(t *testing.T) {
	testCases := []struct {
		name     string
		value    types.String
		expected bool
	}{
		{"null", types.StringNull(), false},
		{"unknown", types.StringUnknown(), false},
		{"empty", types.StringValue(""), false},
		{"set", types.StringValue("value"), true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := isSetStr(tc.value); got != tc.expected {
				t.Errorf("isSetStr(%s) = %t, expected %t", tc.name, got, tc.expected)
			}
		})
	}
}
