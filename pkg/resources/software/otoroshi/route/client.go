package route

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// otoroshiClient talks to the admin API exposed by an Otoroshi add-on. It is
// deliberately separate from the Clever Cloud client: everything this resource
// needs is served by Otoroshi itself, the underlying Java application is never
// touched
type otoroshiClient struct {
	baseURL string
	user    string
	secret  string
	http    *http.Client
}

func newOtoroshiClient(baseURL, user, secret string) *otoroshiClient {
	return &otoroshiClient{
		baseURL: baseURL,
		user:    user,
		secret:  secret,
		http:    &http.Client{Timeout: 30 * time.Second},
	}
}

// do issues an authenticated call and decodes the response into out when not nil.
// A 404 is reported through the found return value rather than as an error, since
// callers need to tell "gone" apart from "broken"
func (c *otoroshiClient) do(ctx context.Context, method, path string, body any, out any) (found bool, err error) {
	var payload io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return false, fmt.Errorf("failed to encode %s %s payload: %w", method, path, err)
		}
		payload = bytes.NewReader(raw)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, payload)
	if err != nil {
		return false, err
	}
	req.SetBasicAuth(c.user, c.secret)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	res, err := c.http.Do(req)
	if err != nil {
		return false, err
	}
	defer res.Body.Close()

	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return false, err
	}

	if res.StatusCode == http.StatusNotFound {
		return false, nil
	}
	if res.StatusCode >= 300 {
		return false, fmt.Errorf("%s %s: status %d: %s", method, path, res.StatusCode, string(raw))
	}

	if out != nil && len(raw) > 0 {
		if err := json.Unmarshal(raw, out); err != nil {
			return false, fmt.Errorf("failed to decode %s %s response: %w", method, path, err)
		}
	}

	return true, nil
}

// WaitReady blocks until the admin API answers. A freshly created add-on takes a
// while to boot and replies 503 in the meantime, so writing a route right after
// creating the Otoroshi would fail without this
func (c *otoroshiClient) WaitReady(ctx context.Context, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	var last error
	for {
		// The template endpoint is read-only and always present, it makes a good probe
		_, last = c.do(ctx, http.MethodGet, "/api/routes/_template", nil, nil)
		if last == nil {
			return nil
		}

		if time.Now().After(deadline) {
			return fmt.Errorf("Otoroshi admin API still unavailable after %s: %w", timeout, last)
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(3 * time.Second):
		}
	}
}

// ngRoute is the subset of an Otoroshi route this resource drives. Every other
// field of the entity is filled in by Otoroshi itself on write, and ignored on read
type ngRoute struct {
	// Otoroshi happily accepts an empty id and creates an unaddressable route,
	// so let it generate one on create by omitting the field
	ID       string     `json:"id,omitempty"`
	Name     string     `json:"name"`
	Enabled  bool       `json:"enabled"`
	Frontend ngFrontend `json:"frontend"`
	Backend  ngBackend  `json:"backend"`
}

type ngFrontend struct {
	// Domains carry the path when there is one, as "host/path"
	Domains   []string `json:"domains"`
	StripPath bool     `json:"strip_path"`
}

type ngBackend struct {
	Targets []ngTarget `json:"targets"`
}

type ngTarget struct {
	Hostname string `json:"hostname"`
	Port     int64  `json:"port"`
	TLS      bool   `json:"tls"`
}

func (c *otoroshiClient) GetRoute(ctx context.Context, id string) (*ngRoute, bool, error) {
	route := &ngRoute{}
	found, err := c.do(ctx, http.MethodGet, "/api/routes/"+id, nil, route)
	if err != nil || !found {
		return nil, found, err
	}
	return route, true, nil
}

func (c *otoroshiClient) CreateRoute(ctx context.Context, route ngRoute) (*ngRoute, error) {
	created := &ngRoute{}
	if _, err := c.do(ctx, http.MethodPost, "/api/routes", route, created); err != nil {
		return nil, err
	}
	return created, nil
}

func (c *otoroshiClient) UpdateRoute(ctx context.Context, route ngRoute) error {
	_, err := c.do(ctx, http.MethodPut, "/api/routes/"+route.ID, route, nil)
	return err
}

func (c *otoroshiClient) DeleteRoute(ctx context.Context, id string) error {
	_, err := c.do(ctx, http.MethodDelete, "/api/routes/"+id, nil, nil)
	return err
}
