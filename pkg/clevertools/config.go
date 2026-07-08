// Package clevertools parses the clever-tools CLI configuration file.
//
// clever-tools >= 4.6.0 moved credentials from the config file root into a
// "profiles" array (to support multiple accounts). The pinned client
// dependency go.clever-cloud.dev/client@v0.1.7 only understands the legacy
// root-level format, so it silently fails to find credentials against a
// modern clever-tools.json. This package understands both formats.
package clevertools

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/adrg/xdg"
)

const (
	configDir      = "clever-cloud"
	configFileName = "clever-tools.json"
)

// Overrides holds per-profile consumer key/secret overrides.
type Overrides struct {
	OAuthConsumerKey    string `json:"OAUTH_CONSUMER_KEY,omitempty"`
	OAuthConsumerSecret string `json:"OAUTH_CONSUMER_SECRET,omitempty"`
}

// Profile is one clever-tools account, as stored in clever-tools.json.
type Profile struct {
	Alias          string     `json:"alias"`
	Token          string     `json:"token"`
	Secret         string     `json:"secret"`
	ExpirationDate string     `json:"expirationDate,omitempty"`
	UserID         string     `json:"userId,omitempty"`
	Email          string     `json:"email,omitempty"`
	Overrides      *Overrides `json:"overrides,omitempty"`
}

// currentFormat is the clever-tools >= 4.6.0 layout.
type currentFormat struct {
	Profiles []Profile `json:"profiles"`
}

// legacyFormat is the flat, single-account layout written by older clever-tools.
type legacyFormat struct {
	Token          string `json:"token"`
	Secret         string `json:"secret"`
	ExpirationDate string `json:"expirationDate,omitempty"`
}

// ConfigFilePath resolves the clever-tools config file, mirroring the lookup
// done by go.clever-cloud.dev/client: XDG config search, falling back to
// $HOME/.config/... since clever-tools does not honor XDG on macOS. Returns
// "" if no file is found.
func ConfigFilePath() string {
	relPath := fmt.Sprintf("%s/%s", configDir, configFileName)

	path, _ := xdg.SearchConfigFile(relPath)
	if path != "" {
		return path
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	fallback := fmt.Sprintf("%s/.config/%s", home, relPath)
	if _, err := os.Stat(fallback); err == nil {
		return fallback
	}

	return ""
}

// ActiveProfile reads path and returns the active profile, following the
// clever-tools semantics. A missing file is not an error: (nil, nil).
func ActiveProfile(path string) (*Profile, error) {
	if path == "" {
		return nil, nil
	}

	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}

		return nil, fmt.Errorf("reading clever-tools config at %q: %w", path, err)
	}

	return ActiveProfileFrom(content)
}

// ActiveProfileFrom parses raw clever-tools.json content and returns the
// active profile: profiles[0] in the current format (clever-tools always
// moves the freshly authenticated profile there, regardless of alias), or
// the legacy flat object as a single "default" profile. Returns (nil, nil)
// when the content is valid JSON but carries no usable credentials, and an
// error only on malformed JSON.
func ActiveProfileFrom(content []byte) (*Profile, error) {
	var current currentFormat
	if err := json.Unmarshal(content, &current); err != nil {
		return nil, fmt.Errorf("parsing clever-tools config: %w", err)
	}

	if len(current.Profiles) > 0 {
		p := current.Profiles[0]
		if p.Token == "" || p.Secret == "" {
			return nil, nil
		}

		return &p, nil
	}

	var legacy legacyFormat
	if err := json.Unmarshal(content, &legacy); err != nil {
		return nil, fmt.Errorf("parsing clever-tools config: %w", err)
	}

	if legacy.Token == "" || legacy.Secret == "" {
		return nil, nil
	}

	return &Profile{
		Alias:          "default",
		Token:          legacy.Token,
		Secret:         legacy.Secret,
		ExpirationDate: legacy.ExpirationDate,
	}, nil
}

// Expired reports whether the profile's expiration date is in the past. A
// missing or unparseable date is never treated as expired: it's a cosmetic
// field the legacy format may lack, and we must not fail closed on it.
func (p *Profile) Expired() bool {
	if p == nil || p.ExpirationDate == "" {
		return false
	}

	t, err := time.Parse(time.RFC3339, p.ExpirationDate)
	if err != nil {
		return false
	}

	return t.Before(time.Now())
}
