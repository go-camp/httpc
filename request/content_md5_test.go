package request

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/go-camp/httpc"
)

func TestContentMD5Builder(t *testing.T) {
	var cases = map[string]struct {
		Body io.Reader

		ExpectBodyLength int
		ExpectContentMD5 string
		ExpectError      string
	}{
		"body is nil": {
			Body: nil,
		},
		"unseekable": {
			Body: io.NopCloser(bytes.NewReader([]byte(`123`))),

			ExpectBodyLength: 3,
			ExpectError:      "new rewind reader failed, body doesn't implement io.Seeker interface",
		},
		"seekable": {
			Body: bytes.NewReader([]byte(`message digest`)),

			ExpectBodyLength: 14,
			ExpectContentMD5: "+WtpfXy3k41SWi8xqvFh0A==",
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			build := ContentMD5Builder{}.Builder(
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
				Body: c.Body,
			}
			_, _, err := build(ctx, req)
			if c.ExpectError != "" {
				if err == nil {
					t.Fatalf("expect error %v, got none", c.ExpectError)
				}
				if !strings.Contains(err.Error(), c.ExpectError) {
					t.Fatalf("expect error to contain %q, got %v", c.ExpectError, err)
				}
			} else if err != nil {
				t.Fatalf("except no err, got %v", err)
			}

			if m := req.Header.Get("Content-MD5"); m != c.ExpectContentMD5 {
				t.Fatalf("expect content md5 is %s, got %s", c.ExpectContentMD5, m)
			}

			var bodyLength int
			if c.Body != nil {
				bodyContent, err := io.ReadAll(c.Body)
				if err != nil {
					t.Fatalf("expect no err, got %v", err)
				}
				bodyLength = len(bodyContent)
			}
			if bodyLength != c.ExpectBodyLength {
				t.Fatalf("expect body length is %d, got %d", c.ExpectBodyLength, bodyLength)
			}
		})
	}
}
