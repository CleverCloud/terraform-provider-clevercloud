package client

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/pkg/errors"
)

type Response[T any] interface {
	Error() error
	HasError() bool

	StatusCode() int
	IsNotFoundError() bool

	SozuID() string

	Equal(anotherResponse Response[T]) bool

	Payload() *T
}

type response[T any] struct {
	*http.Response
	rawBody []byte
	err     error
	payload T
}

func fromHTTPResponse[T any](httpRes *http.Response) Response[T] {
	res := &response[T]{Response: httpRes}

	var readBodyErr error
	res.rawBody, readBodyErr = io.ReadAll(res.Body)

	if httpRes.StatusCode >= 300 {
		err := errors.New(string(res.rawBody))
		res.err = errors.Wrapf(err, "invalid response from CleverCloud API (status=%d)", httpRes.StatusCode)

		return res
	}

	if readBodyErr != nil {
		res.err = errors.Wrap(readBodyErr, "cannot read response body")

		return res
	}

	switch any(res.payload).(type) {
	case Nothing:
		// Do not try to parse an empty body
	case Raw:
		res.payload = any(res.rawBody).(T)
		return res
	case String:
		res.payload = any(string(res.rawBody)).(T)
	default:
		err := json.Unmarshal(res.rawBody, &res.payload)
		if err != nil {
			res.err = errors.Wrap(err, "cannot parse response body")

			return res
		}
	}

	return res
}

func fromError[T any](err error) Response[T] {
	return &response[T]{err: err}
}

func (r *response[T]) StatusCode() int {
	if r.Response == nil {
		return 0
	}

	return r.Response.StatusCode
}

func (r *response[T]) SozuID() string {
	if r.Response == nil {
		return ""
	}

	return r.Response.Header.Get("Sozu-Id")
}

func (r *response[T]) Error() error {
	return r.err
}

func (r *response[T]) HasError() bool {
	return r.err != nil
}

func (r *response[T]) IsNotFoundError() bool {
	return r.Response.StatusCode == http.StatusNotFound
}

func (r *response[T]) Equal(anotherResponse Response[T]) bool {
	return r.StatusCode() == anotherResponse.StatusCode() &&
		r.SozuID() == anotherResponse.SozuID()
}

func (r *response[T]) Payload() *T {
	return &r.payload
}
