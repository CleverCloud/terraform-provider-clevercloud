package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	otel "go.opentelemetry.io/contrib/instrumentation/net/http/httptrace/otelhttptrace"
)

// Client is a wrapped HTTP client used to contact CleverCloud API.
type Client struct {
	httpClient    *http.Client
	authenticator Authenticator
	endpoint      string
	log           logrus.FieldLogger
}

// New instantiate a new CleverCloud client with options.
func New(options ...func(*Client)) *Client {
	discardLogger := logrus.New()
	discardLogger.Out = io.Discard
	discardLogger.Level = logrus.PanicLevel

	c := &Client{
		httpClient:    http.DefaultClient,
		authenticator: nil,
		endpoint:      API_ENDPOINT,
		log:           discardLogger,
	}

	for _, option := range options {
		option(c)
	}

	return c
}

func userAgent() string {
	return fmt.Sprintf(
		"Go/%s %s/%s cc-client/%s",
		strings.TrimPrefix(runtime.Version(), "go"),
		runtime.GOARCH, runtime.GOOS,
		CLIENT_VERSION,
	)
}

func request[T any](ctx context.Context, c *Client, method string, path string, payload interface{}) Response[T] {
	if c == nil {
		return fromError[T](errors.New("expect non nil client"))
	}

	url := fmt.Sprintf("%s%s", c.endpoint, path)
	ctx = mustContext(ctx)

	body := []byte{}

	if payload != nil {
		var err error
		body, err = json.Marshal(payload)

		if err != nil {
			return fromError[T](errors.Wrap(err, "failed to serialize request body"))
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(body))
	if err != nil {
		return fromError[T](errors.Wrap(err, "failed to build CleverCloud API request"))
	}

	otel.Inject(ctx, req)
	req.Header.Set("User-Agent", userAgent())

	if len(body) != 0 {
		req.Header.Set("Content-Type", "application/json")
	}

	if c.authenticator != nil {
		c.authenticator.Sign(req)
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		c.log.Warnf("RESPONSE:\t%s\t%s\t->\t%+v", req.Method, req.URL.String(), err.Error())

		return fromError[T](errors.Wrap(err, "failed to build CleverCloud API request"))
	}

	c.log.Infof("RESPONSE:\t%s\t%s\t->\t%s", req.Method, req.URL.String(), res.Status)
	defer res.Body.Close()

	return fromHTTPResponse[T](res)
}

func (c *Client) Authenticator() Authenticator {
	return c.authenticator
}

// Perform a GET request.
func Get[T any](ctx context.Context, c *Client, path string) Response[T] {
	return request[T](ctx, c, http.MethodGet, path, nil)
}

// Perform a POST request.
func Post[T any](ctx context.Context, c *Client, path string, payload interface{}) Response[T] {
	return request[T](ctx, c, http.MethodPost, path, payload)
}

// Perform a PUT request.
func Put[T any](ctx context.Context, c *Client, path string, payload interface{}) Response[T] {
	return request[T](ctx, c, http.MethodPut, path, payload)
}

// Perform a DELETE request.
func Delete[T any](ctx context.Context, c *Client, path string) Response[T] {
	return request[T](ctx, c, http.MethodDelete, path, nil)
}

// Perform a PATCH request.
func Patch[T any](ctx context.Context, c *Client, path string, payload interface{}) Response[T] {
	return request[T](ctx, c, http.MethodPatch, path, payload)
}

// Perform an SSE request.
func Stream[T any](ctx context.Context, c *Client, path string) StreamResponse[T] {
	url := fmt.Sprintf("%s%s", c.endpoint, path)
	ctx = mustContext(ctx)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fromErrorStream[T](errors.Wrap(err, "failed to build CleverCloud API request"))
	}

	otel.Inject(ctx, req)
	req.Header.Set("User-Agent", userAgent())

	if c.authenticator != nil {
		c.authenticator.Sign(req)
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		res.Body.Close()

		return fromErrorStream[T](errors.Wrap(err, "failed to build CleverCloud API request"))
	}

	return fromHTTPStream[T](res)
}
