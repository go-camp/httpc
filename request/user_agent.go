package request

import (
	"context"

	"github.com/go-camp/httpc"
)

type UserAgent struct {
	Name    string
	Version string
}

func (a UserAgent) IsEmpty() bool {
	return a.Name == "" && a.Version == ""
}

// UserAgentBuilder sets the specified user agent to the value of the User-Agent header.
type UserAgentBuilder struct {
	UserAgent UserAgent
}

func (b UserAgentBuilder) Builder(build httpc.BuildFunc) httpc.BuildFunc {
	return func(ctx context.Context, req *httpc.Request) (interface{}, httpc.Metadata, error) {
		return b.build(ctx, req, build)
	}
}

func userAgentFormatHeaderValue(userAgent UserAgent) string {
	if userAgent.Name == "" {
		return ""
	}
	if userAgent.Version == "" {
		return userAgent.Name
	}
	return userAgent.Name + "/" + userAgent.Version
}

func (b UserAgentBuilder) build(ctx context.Context, req *httpc.Request, build httpc.BuildFunc) (
	output interface{}, md httpc.Metadata, err error,
) {
	if req.Header.Get(headerUserAgent) != "" {
		return build(ctx, req)
	}

	userAgent := userAgentFormatHeaderValue(b.UserAgent)
	if userAgent != "" {
		req.Header.Set(headerUserAgent, userAgent)
	}

	return build(ctx, req)
}
