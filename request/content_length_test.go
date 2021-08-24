package request

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"
	"testing"

	"github.com/go-camp/httpc"
)

func TestContentLengthBuilder(t *testing.T) {
	var cases = map[string]struct {
		Body io.Reader

		ExpectContentLength int64
		ExpectBodyLength    int
	}{
		"body is nil": {
			Body:                nil,
			ExpectContentLength: 0,
			ExpectBodyLength:    0,
		},
		"has Len method": {
			Body: bytes.NewReader([]byte(`message digest`)),

			ExpectContentLength: 14,
			ExpectBodyLength:    14,
		},
		"unseekable": {
			Body: io.NopCloser(bytes.NewReader([]byte(`123`))),

			ExpectContentLength: -1,
			ExpectBodyLength:    3,
		},
		"seekable": {
			Body: io.NewSectionReader(bytes.NewReader([]byte(`123`)), 1, 2),

			ExpectContentLength: 2,
			ExpectBodyLength:    2,
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			var buildCallCount int
			build := ContentLengthBuilder{}.Builder(
				func(ctx context.Context, req *httpc.Request) (output interface{}, md httpc.Metadata, err error) {
					buildCallCount++
					return
				},
			)
			req := &httpc.Request{
				Request: &http.Request{
					Header:        http.Header{},
					URL:           &url.URL{},
					ContentLength: -1,
				},
				Body: c.Body,
			}
			_, _, err := build(ctx, req)
			if err != nil {
				t.Fatalf("expect no err, got %v", err)
			}

			if buildCallCount != 1 {
				t.Fatalf("expect build call count is %d, got %d", 1, buildCallCount)
			}
			if c.ExpectContentLength != req.ContentLength {
				t.Fatalf("expect request content length is %d, got %d", c.ExpectContentLength, req.ContentLength)
			}

			var bodyLength int
			if c.Body != nil {
				bodyContent, err := io.ReadAll(c.Body)
				if err != nil {
					t.Fatalf("expect no err, got %v", err)
				}
				bodyLength = len(bodyContent)
			}
			if c.ExpectBodyLength != bodyLength {
				t.Fatalf("expect body length is %d, got %d", c.ExpectBodyLength, bodyLength)
			}
		})
	}
}
