package attributes

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestUserPasswordInput(t *testing.T) {
	tests := []struct {
		name        string
		value       types.String
		expectError bool
		errorMsg    string
	}{{
		name:        "valid user:password format",
		value:       types.StringValue("user:password"),
		expectError: false,
	}, {
		name:        "valid with PAT token",
		value:       types.StringValue("github-user:ghp_1234567890abcdefghijklmnopqrstuvwxyz"),
		expectError: false,
	}, {
		name:        "valid with complex password containing colons",
		value:       types.StringValue("user:pass:word:with:colons"),
		expectError: false,
	}, {
		name:        "valid with email as username",
		value:       types.StringValue("user@example.com:password123"),
		expectError: false,
	}, {
		name:        "invalid - missing colon",
		value:       types.StringValue("userpassword"),
		expectError: true,
		errorMsg:    "expect user:password",
	}, {
		name:        "valid - only colon (empty user and password)",
		value:       types.StringValue(":"),
		expectError: true,
		errorMsg:    "user and password cannot be both empty",
	}, {
		name:        "invalid - empty string",
		value:       types.StringValue(""),
		expectError: true,
		errorMsg:    "expect user:password",
	}, {
		name:        "valid - null value (should not error)",
		value:       types.StringNull(),
		expectError: false,
	}, {
		name:        "valid - unknown value (should not error)",
		value:       types.StringUnknown(),
		expectError: false,
	}, {
		name:        "valid - empty password",
		value:       types.StringValue("user:"),
		expectError: false,
	}, {
		name:        "valid - empty username",
		value:       types.StringValue(":password"),
		expectError: false,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()

			req := validator.StringRequest{
				Path:           path.Root("test"),
				PathExpression: path.MatchRoot("test"),
				ConfigValue:    tt.value,
			}
			res := &validator.StringResponse{}

			UserPasswordInput.ValidateString(ctx, req, res)

			hasError := res.Diagnostics.HasError()
			if hasError != tt.expectError {
				t.Errorf("expected error: %v, got error: %v", tt.expectError, hasError)
				if hasError {
					t.Logf("Diagnostics: %v", res.Diagnostics)
				}
			}

			if tt.expectError && len(tt.errorMsg) > 0 {
				found := false
				for _, diag := range res.Diagnostics.Errors() {
					if diag.Detail() == tt.errorMsg {
						found = true
						break
					}
				}
				if !found {
					t.Errorf(
						"expected error message containing '%s', got diagnostics: %v",
						tt.errorMsg,
						res.Diagnostics)
				}
			}
		})
	}
}

func TestRepositoryURL(t *testing.T) {
	tests := []struct {
		name        string
		value       types.String
		expectError bool
	}{
		{name: "https", value: types.StringValue("https://github.com/CleverCloud/nodejs-example.git")},
		{name: "http", value: types.StringValue("http://git.example.com/repo.git")},
		{name: "https with credentials", value: types.StringValue("https://user:token@github.com/owner/repo.git")},
		{name: "absolute local path", value: types.StringValue("file:///home/user/repo")},
		{name: "relative local path", value: types.StringValue("file://./repo")},
		{name: "null", value: types.StringNull()},
		{name: "unknown", value: types.StringUnknown()},
		{name: "bare absolute path", value: types.StringValue("/home/user/repo"), expectError: true},
		{name: "bare relative path", value: types.StringValue("../repo"), expectError: true},
		{name: "host without scheme", value: types.StringValue("github.com/owner/repo.git"), expectError: true},
		{name: "scp-like ssh", value: types.StringValue("git@github.com:owner/repo.git"), expectError: true},
		{name: "ssh", value: types.StringValue("ssh://git@github.com/owner/repo.git"), expectError: true},
		{name: "empty", value: types.StringValue(""), expectError: true},
		{name: "file without path", value: types.StringValue("file://"), expectError: true},
		{name: "https without host", value: types.StringValue("https://"), expectError: true},
		{name: "uppercase scheme", value: types.StringValue("FILE:///home/user/repo"), expectError: true},
		{name: "scheme without slashes", value: types.StringValue("file:/home/user/repo"), expectError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := validator.StringRequest{
				Path:           path.Root("test"),
				PathExpression: path.MatchRoot("test"),
				ConfigValue:    tt.value,
			}
			res := &validator.StringResponse{}

			RepositoryURL.ValidateString(t.Context(), req, res)

			if res.Diagnostics.HasError() != tt.expectError {
				t.Errorf("expected error: %v, got diagnostics: %v", tt.expectError, res.Diagnostics)
			}
		})
	}
}
