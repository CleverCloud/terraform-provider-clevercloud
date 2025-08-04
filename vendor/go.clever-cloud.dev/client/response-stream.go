package client

import (
	"bufio"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

type StreamResponse[T any] interface {
	Error() error
	HasError() bool

	StatusCode() int
	IsNotFoundError() bool

	SozuID() string

	Equal(anotherResponse StreamResponse[T]) bool

	Close()
	Payload() <-chan *StreamEvent[T]
}

type streamResponse[T any] struct {
	*http.Response
	err      error
	close    chan struct{}
	payloads chan *StreamEvent[T]
}

// https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events/Using_server-sent_events#event_stream_format
type StreamEvent[T any] struct {
	Event string
	Data  []byte
	ID    []byte
	Retry int64
}

func (se *StreamEvent[T]) String() string {
	return fmt.Sprintf("Event=%s\tID=%s\t%s", se.Event, string(se.ID), string(se.Data))
}

func fromErrorStream[T any](err error) StreamResponse[T] {
	return &streamResponse[T]{
		err:   err,
		close: make(chan struct{}),
	}
}

func fromHTTPStream[T any](httpRes *http.Response) StreamResponse[T] {
	defer httpRes.Body.Close()

	res := &streamResponse[T]{
		Response: httpRes,
		close:    make(chan struct{}, 2),
		payloads: make(chan *StreamEvent[T], 10),
	}

	if httpRes.StatusCode >= 300 {
		res.err = errors.Errorf("invalid response from CleverCloud API (status=%d)", httpRes.StatusCode)

		return res
	}

	stream := bufio.NewScanner(res.Body)
	stream.Split(spliter)

	go res.loop(stream)

	return res
}

func (r *streamResponse[T]) StatusCode() int {
	if r.Response == nil {
		return 0
	}

	return r.Response.StatusCode
}

func (r *streamResponse[T]) SozuID() string {
	if r.Response == nil {
		return ""
	}

	return r.Response.Header.Get("Sozu-Id")
}

func (r *streamResponse[T]) Error() error {
	return r.err
}

func (r *streamResponse[T]) HasError() bool {
	return r.err != nil
}

func (r *streamResponse[T]) IsNotFoundError() bool {
	return r.Response.StatusCode == http.StatusNotFound
}

func (r *streamResponse[T]) Equal(anotherResponse StreamResponse[T]) bool {
	return r.StatusCode() == anotherResponse.StatusCode() &&
		r.SozuID() == anotherResponse.SozuID()
}

func (r *streamResponse[T]) Payload() <-chan *StreamEvent[T] {
	return r.payloads
}

func (r *streamResponse[T]) Close() {
	// non-blocking chan write
	select {
	case r.close <- struct{}{}:
	default:
	}
}

func (r *streamResponse[T]) loop(scan *bufio.Scanner) {
	defer func() {
		close(r.payloads)
	}()

	raws := make(chan string)
	errs := make(chan error)

	go func() {
		for scan.Scan() {
			if err := scan.Err(); err != nil {
				errs <- err

				break
			}

			raws <- scan.Text()
		}
		close(raws)
		close(errs)
	}()

	for {
		select {
		case <-r.close:
			return
		case er := <-errs:
			r.err = er

			return
		case <-r.Request.Context().Done():
			r.err = r.Request.Context().Err()

			return
		case raw := <-raws:
			r.payloads <- rawToStreamEvent[T](raw)
		}
	}
}

func rawToStreamEvent[T any](raw string) *StreamEvent[T] {
	ev := &StreamEvent[T]{Data: make([]byte, 0)}

	for _, split := range strings.Split(raw, "\n") {
		switch {
		case strings.HasPrefix(split, ":"):
			continue
		case strings.HasPrefix(split, "id:"):
			ev.ID = []byte(strings.TrimPrefix(split, "id:"))
		case strings.HasPrefix(split, "event:"):
			ev.Event = strings.TrimPrefix(split, "event:")
		case strings.HasPrefix(split, "retry:"):
			n, _ := strconv.ParseInt(strings.TrimPrefix(split, "retry:"), 10, 64)
			ev.Retry = n
		case strings.HasPrefix(split, "data:"):
			ev.Data = append(ev.Data, []byte(strings.TrimPrefix(split, "data:"))...)
		}
	}

	return ev
}

func spliter(data []byte, atEOF bool) (int, []byte, error) {
	size := len(data)

	data = []byte(strings.TrimSuffix(string(data), "\n\n"))

	if atEOF {
		return size, data, bufio.ErrFinalToken
	}
	return size, data, nil
}
