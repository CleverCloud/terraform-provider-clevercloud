package clevertools

import (
	"path/filepath"
	"testing"
)

func TestActiveProfileFrom(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		wantNil     bool
		wantErr     bool
		wantAlias   string
		wantToken   string
		wantSecret  string
		wantCK      string
		wantCS      string
		wantOverrid bool
	}{
		{
			name:      "current format, one profile",
			content:   `{"version":1,"profiles":[{"alias":"perso","token":"tok1","secret":"sec1"}]}`,
			wantAlias: "perso",
			wantToken: "tok1",
			wantSecret: "sec1",
		},
		{
			name: "current format, several profiles returns profiles[0] regardless of alias",
			content: `{"version":1,"profiles":[
				{"alias":"work","token":"tok-work","secret":"sec-work"},
				{"alias":"default","token":"tok-default","secret":"sec-default"}
			]}`,
			wantAlias:  "work",
			wantToken:  "tok-work",
			wantSecret: "sec-work",
		},
		{
			name:       "legacy flat format",
			content:    `{"token":"tok-legacy","secret":"sec-legacy","expirationDate":"2099-01-01T00:00:00.000Z"}`,
			wantAlias:  "default",
			wantToken:  "tok-legacy",
			wantSecret: "sec-legacy",
		},
		{
			name:    "empty profiles array means no credentials",
			content: `{"version":1,"profiles":[]}`,
			wantNil: true,
		},
		{
			name:    "profile missing token",
			content: `{"version":1,"profiles":[{"alias":"default","secret":"sec1"}]}`,
			wantNil: true,
		},
		{
			name:    "profile missing secret",
			content: `{"version":1,"profiles":[{"alias":"default","token":"tok1"}]}`,
			wantNil: true,
		},
		{
			name:    "malformed json",
			content: `{not json`,
			wantErr: true,
		},
		{
			name:        "profile with overrides",
			content:     `{"version":1,"profiles":[{"alias":"default","token":"tok1","secret":"sec1","overrides":{"OAUTH_CONSUMER_KEY":"ck","OAUTH_CONSUMER_SECRET":"cs"}}]}`,
			wantAlias:   "default",
			wantToken:   "tok1",
			wantSecret:  "sec1",
			wantCK:      "ck",
			wantCS:      "cs",
			wantOverrid: true,
		},
		{
			name:      "future version with profiles still parses",
			content:   `{"version":2,"profiles":[{"alias":"default","token":"tok1","secret":"sec1"}]}`,
			wantAlias: "default",
			wantToken: "tok1",
			wantSecret: "sec1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profile, err := ActiveProfileFrom([]byte(tt.content))

			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected an error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.wantNil {
				if profile != nil {
					t.Fatalf("expected nil profile, got %+v", profile)
				}
				return
			}

			if profile == nil {
				t.Fatalf("expected a profile, got nil")
			}
			if profile.Alias != tt.wantAlias {
				t.Errorf("Alias = %q, want %q", profile.Alias, tt.wantAlias)
			}
			if profile.Token != tt.wantToken {
				t.Errorf("Token = %q, want %q", profile.Token, tt.wantToken)
			}
			if profile.Secret != tt.wantSecret {
				t.Errorf("Secret = %q, want %q", profile.Secret, tt.wantSecret)
			}

			if tt.wantOverrid {
				if profile.Overrides == nil {
					t.Fatalf("expected overrides, got nil")
				}
				if profile.Overrides.OAuthConsumerKey != tt.wantCK {
					t.Errorf("OAuthConsumerKey = %q, want %q", profile.Overrides.OAuthConsumerKey, tt.wantCK)
				}
				if profile.Overrides.OAuthConsumerSecret != tt.wantCS {
					t.Errorf("OAuthConsumerSecret = %q, want %q", profile.Overrides.OAuthConsumerSecret, tt.wantCS)
				}
			}
		})
	}
}

func TestActiveProfile_MissingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "does-not-exist.json")

	profile, err := ActiveProfile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if profile != nil {
		t.Fatalf("expected nil profile for missing file, got %+v", profile)
	}
}

func TestProfile_Expired(t *testing.T) {
	tests := []struct {
		name string
		date string
		want bool
	}{
		{"past date", "2000-01-01T00:00:00.000Z", true},
		{"future date", "2099-01-01T00:00:00.000Z", false},
		{"empty", "", false},
		{"garbage", "not-a-date", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Profile{ExpirationDate: tt.date}
			if got := p.Expired(); got != tt.want {
				t.Errorf("Expired() = %v, want %v", got, tt.want)
			}
		})
	}
}
