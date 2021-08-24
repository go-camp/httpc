package request

import (
	"context"
	"net/http"
	"net/url"
	"testing"

	"github.com/go-camp/httpc"
)

func TestRequestIDBuilder(t *testing.T) {
	expectReqID := "12ca095b1798410e91b0f5f2ec6b6e05"
	build := RequestIDBuilder{
		IDGenerator: func() (string, error) {
			return expectReqID, nil
		},
	}.Builder(
		func(ctx context.Context, req *httpc.Request) (output interface{}, md httpc.Metadata, err error) {
			return
		},
	)

	ctx := context.Background()
	req := &httpc.Request{
		Request: &http.Request{
			URL:           &url.URL{},
			Header:        http.Header{},
			ContentLength: 0,
		},
	}
	_, _, err := build(ctx, req)
	if err != nil {
		t.Fatalf("except no err, got %v", err)
	}

	reqID := req.Header.Get("X-Request-Id")
	if expectReqID != reqID {
		t.Fatalf("expect request id is %s, got %s", expectReqID, reqID)
	}
}
