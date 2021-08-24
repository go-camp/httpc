package request

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/go-camp/httpc"
)

func DefaultIDGenerator() (string, error) {
	var id [16]byte
	if _, err := rand.Read(id[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(id[:]), nil
}

// RequestIDBuilder sets a unique id to the value of specified request header.
type RequestIDBuilder struct {
	// default: X-Request-ID
	Header      string
	IDGenerator func() (string, error)
}

type requestIDError struct {
	While string
	Err   error
}

func (e *requestIDError) Error() string {
	return fmt.Sprintf("request id builder, %s failed, %v", e.While, e.Err)
}

func (e *requestIDError) Unwrap() error {
	return e.Err
}

func (b RequestIDBuilder) Builder(build httpc.BuildFunc) httpc.BuildFunc {
	return func(ctx context.Context, req *httpc.Request) (interface{}, httpc.Metadata, error) {
		return b.build(ctx, req, build)
	}
}

func (b RequestIDBuilder) header() string {
	if b.Header == "" {
		return headerXRequestID
	}
	return b.Header
}

func (b RequestIDBuilder) idGenerator() func() (string, error) {
	if b.IDGenerator == nil {
		return DefaultIDGenerator
	}
	return b.IDGenerator
}

func (b RequestIDBuilder) build(ctx context.Context, req *httpc.Request, build httpc.BuildFunc) (
	output interface{}, md httpc.Metadata, err error,
) {
	if req.Header.Get(b.header()) != "" {
		return build(ctx, req)
	}

	var id string
	id, err = b.idGenerator()()
	if err != nil {
		err = &requestIDError{While: "id generate", Err: err}
		return
	}
	if id != "" {
		req.Header.Set(b.header(), id)
	}

	return build(ctx, req)
}
