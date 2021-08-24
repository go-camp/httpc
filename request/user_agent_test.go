package request

import (
	"context"
	"net/http"
	"net/url"
	"testing"

	"github.com/go-camp/httpc"
)

func TestUserAgentBuilder(t *testing.T) {
	ctx := context.Background()
	build := UserAgentBuilder{
		UserAgent: UserAgent{
			Name:    "test",
			Version: "0.1",
		},
	}.Builder(
		func(ctx context.Context, req *httpc.Request) (output interface{}, md httpc.Metadata, err error) {
			return
		},
	)
	req := &httpc.Request{
		Request: &http.Request{
			Header:        http.Header{},
			URL:           &url.URL{},
			ContentLength: -1,
		},
	}
	_, _, err := build(ctx, req)
	if err != nil {
		t.Fatalf("except no err, got %v", err)
	}

	userAgent := req.Header.Get("User-Agent")
	expectUserAgent := "test/0.1"
	if expectUserAgent != userAgent {
		t.Fatalf("expect user agent is %s, got %s", expectUserAgent, userAgent)
	}
}
