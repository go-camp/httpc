package request

import (
	"context"
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"

	"github.com/go-camp/httpc"
)

// ContentMD5Builder calculates the md5 checksum of the request body
// and sets the checksum to the value of the Content-MD5 header.
//
// This builder requires the request Body to implement io.Seeker interface.
type ContentMD5Builder struct {
}

type contentMD5Error struct {
	While string
	Err   error
}

func (e *contentMD5Error) Error() string {
	return fmt.Sprintf("request content md5 builder, %s failed, %v", e.While, e.Err)
}

func (e *contentMD5Error) Unwrap() error {
	return e.Err
}

func calculateMD5(r io.Reader) (string, error) {
	h := md5.New()
	_, err := io.Copy(h, r)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(h.Sum(nil)), nil
}

func (b ContentMD5Builder) Builder(build httpc.BuildFunc) httpc.BuildFunc {
	return func(ctx context.Context, req *httpc.Request) (interface{}, httpc.Metadata, error) {
		return b.build(ctx, req, build)
	}
}

func (b ContentMD5Builder) build(ctx context.Context, req *httpc.Request, build httpc.BuildFunc) (
	output interface{}, md httpc.Metadata, err error,
) {
	if req.Header.Get(headerContentMD5) != "" {
		return build(ctx, req)
	}
	if req.Body == nil ||
		req.Body == http.NoBody {
		return build(ctx, req)
	}

	rr, err := newRewindReader(req.Body)
	if err != nil {
		return output, md, &contentMD5Error{
			While: "new rewind reader",
			Err:   err,
		}
	}
	m, err := calculateMD5(rr)
	if err != nil {
		return output, md, &contentMD5Error{
			While: "calculate body md5",
			Err:   err,
		}
	}
	if err = rr.Rewind(); err != nil {
		return output, md, &contentMD5Error{
			While: "body rewind",
			Err:   err,
		}
	}
	req.Header.Set(headerContentMD5, m)

	return build(ctx, req)
}
