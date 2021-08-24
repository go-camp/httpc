package request

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/go-camp/httpc"
)

// ContentLengthBuilder gets the bytes length from the request Body and
// sets the length to the request ContentLength.
type ContentLengthBuilder struct {
}

type contentLengthError struct {
	While string
	Err   error
}

func (e *contentLengthError) Error() string {
	return fmt.Sprintf("request content length builder, %s failed, %v", e.While, e.Err)
}

func (e *contentLengthError) Unwrap() error {
	return e.Err
}

func (b ContentLengthBuilder) Builder(build httpc.BuildFunc) httpc.BuildFunc {
	return func(ctx context.Context, req *httpc.Request) (interface{}, httpc.Metadata, error) {
		return b.build(ctx, req, build)
	}
}

func (b ContentLengthBuilder) build(ctx context.Context, req *httpc.Request, build httpc.BuildFunc) (
	output interface{}, md httpc.Metadata, err error,
) {
	var contentLength int64 = -1
	if req.Body == nil || req.Body == http.NoBody {
		contentLength = 0
	} else {
		if b, ok := req.Body.(interface{ Len() int }); ok {
			contentLength = int64(b.Len())
		} else if sr, ok := req.Body.(io.ReadSeeker); ok {
			startPos, err := sr.Seek(0, io.SeekCurrent)
			if err != nil {
				return output, md, &contentLengthError{
					While: "seek current",
					Err:   err,
				}
			}
			endPos, err := sr.Seek(0, io.SeekEnd)
			if err != nil {
				return output, md, &contentLengthError{
					While: "seek end",
					Err:   err,
				}
			}
			_, err = sr.Seek(startPos, io.SeekStart)
			if err != nil {
				return output, md, &contentLengthError{
					While: "seek start",
					Err:   err,
				}
			}
			contentLength = endPos - startPos
		}
		if contentLength == 0 {
			req.Body = http.NoBody
		}
	}
	req.ContentLength = contentLength

	return build(ctx, req)
}
