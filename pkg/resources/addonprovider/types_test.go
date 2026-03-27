package addonprovider

import (
	"context"
	"net/url"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func TestURLType_ValueFromString(t *testing.T) {
	ctx := context.Background()
	urlType := URLType{}

	tests := []struct {
		name        string
		input       string
		wantError   bool
		errorPrefix string
	}{
		{
			name:      "valid https url",
			input:     "https://api.example.com/resource",
			wantError: false,
		},
		{
			name:      "valid https url with path",
			input:     "https://example.com/path/to/resource",
			wantError: false,
		},
		{
			name:      "valid https url with port",
			input:     "https://example.com:8443/resource",
			wantError: false,
		},
		{
			name:      "http url should parse (validation is done by validator)",
			input:     "http://api.example.com/resource",
			wantError: false,
		},
		{
			name:      "ftp url should parse (validation is done by validator)",
			input:     "ftp://api.example.com/resource",
			wantError: false,
		},
		{
			name:      "url without scheme should parse as relative URL",
			input:     "not-a-valid-url",
			wantError: false, // Valid relative URL
		},
		{
			name:        "malformed url should fail",
			input:       "://example.com",
			wantError:   true,
			errorPrefix: "Invalid URL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stringValue := basetypes.NewStringValue(tt.input)
			_, diags := urlType.ValueFromString(ctx, stringValue)

			hasError := diags.HasError()
			if hasError != tt.wantError {
				t.Errorf("ValueFromString() hasError = %v, wantError %v", hasError, tt.wantError)
				if hasError {
					t.Logf("Diagnostics: %v", diags)
				}
			}

			if tt.wantError && hasError && tt.errorPrefix != "" {
				found := false
				for _, diag := range diags.Errors() {
					if len(diag.Summary()) >= len(tt.errorPrefix) && diag.Summary()[:len(tt.errorPrefix)] == tt.errorPrefix {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected error with prefix %q, but got: %v", tt.errorPrefix, diags)
				}
			}
		})
	}
}

func TestURLType_NullAndUnknown(t *testing.T) {
	ctx := context.Background()
	urlType := URLType{}

	t.Run("null value", func(t *testing.T) {
		stringValue := basetypes.NewStringNull()
		val, diags := urlType.ValueFromString(ctx, stringValue)

		if diags.HasError() {
			t.Errorf("Null value should not produce errors: %v", diags)
		}

		urlValue, ok := val.(URLValue)
		if !ok {
			t.Fatal("Value is not URLValue")
		}

		if !urlValue.IsNull() {
			t.Error("Expected null value")
		}
	})

	t.Run("unknown value", func(t *testing.T) {
		stringValue := basetypes.NewStringUnknown()
		val, diags := urlType.ValueFromString(ctx, stringValue)

		if diags.HasError() {
			t.Errorf("Unknown value should not produce errors: %v", diags)
		}

		urlValue, ok := val.(URLValue)
		if !ok {
			t.Fatal("Value is not URLValue")
		}

		if !urlValue.IsUnknown() {
			t.Error("Expected unknown value")
		}
	})
}

func TestURLValue_URL(t *testing.T) {
	url1, _ := url.Parse("https://api.example.com/resource")
	url2, _ := url.Parse("https://api.example.com:8443/path")

	tests := []struct {
		name     string
		input    *url.URL
		wantHost string
		wantPath string
	}{
		{
			name:     "simple url",
			input:    url1,
			wantHost: "api.example.com",
			wantPath: "/resource",
		},
		{
			name:     "url with port",
			input:    url2,
			wantHost: "api.example.com:8443",
			wantPath: "/path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val := NewURLValue(tt.input)
			url := val.URL()

			if url == nil {
				t.Fatal("URL() returned nil")
			}

			if url.Host != tt.wantHost {
				t.Errorf("URL().Host = %q, want %q", url.Host, tt.wantHost)
			}

			if url.Path != tt.wantPath {
				t.Errorf("URL().Path = %q, want %q", url.Path, tt.wantPath)
			}

			if url.Scheme != "https" {
				t.Errorf("URL().Scheme = %q, want %q", url.Scheme, "https")
			}
		})
	}
}
