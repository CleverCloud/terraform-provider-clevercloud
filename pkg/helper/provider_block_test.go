package helper

import "testing"

func TestProvider_String(t *testing.T) {
	tests := []struct {
		name   string
		fields *Provider
		want   string
	}{
		// TODO: Add test cases.

		{name: "test1", fields: NewProvider("clevercloud").OrganisationName("clevercloud"), want: `provider "clevercloud" {
	organisation = "clevercloud"
}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.fields.String(); got != tt.want {
				t.Errorf("Provider.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
