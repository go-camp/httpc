package httpc

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
)

// Request same as http.Request, but the Body is treated as io.Reader instead of io.ReadCloser.
// Even if the Body implements the io.Closer interface, the body will not be closed.
type Request struct {
	*http.Request

	Body io.Reader
}

func NewRequest(ctx context.Context, method, url string, body io.Reader) (*Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return nil, err
	}
	return &Request{Request: req, Body: body}, nil
}

func (r *Request) Clone(ctx context.Context) *Request {
	rr := r.Request.Clone(ctx)
	r2 := *r
	r2.Request = rr
	return &r2
}

type safeBody struct {
	mux sync.Mutex
	r   io.Reader
}

func (b *safeBody) Close() error {
	b.mux.Lock()
	defer b.mux.Unlock()
	b.r = nil
	return nil
}

func (b *safeBody) Read(p []byte) (int, error) {
	b.mux.Lock()
	defer b.mux.Unlock()
	if b.r == nil {
		return 0, io.EOF
	}
	return b.r.Read(p)
}

func (r *Request) Build() *http.Request {
	rr := r.Request.Clone(r.Context())
	if r.Body == nil || r.Body == http.NoBody {
		rr.Body = nil
		rr.ContentLength = 0
	} else {
		rr.Body = &safeBody{r: r.Body}
	}
	return rr
}

// RequestSendError wraps the error returned by the Do() method.
type RequestSendError struct {
	Err error
}

func (e *RequestSendError) Error() string {
	return fmt.Sprintf("request send failed, %v", e.Err)
}

func (e *RequestSendError) Unwrap() error {
	return e.Err
}

// SerializationError wraps any errors that occur while serializing the request.
type SerializationError struct {
	Err error
}

func (e *SerializationError) Error() string {
	return fmt.Sprintf("request serialization failed, %v", e.Err)
}

func (e *SerializationError) Unwrap() error {
	return e.Err
}
