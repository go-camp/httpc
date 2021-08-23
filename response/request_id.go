package response

import (
	"net/http"

	"github.com/go-camp/httpc"
)

type mdRequestIDKey struct{}

func GetRequestID(md httpc.Metadata) (string, bool) {
	v, ok := md.Get(mdRequestIDKey{})
	if !ok {
		return "", false
	}
	id, ok := v.(string)
	return id, ok
}

// RequestIDDeserializer sets the response's request id to metadata.
type RequestIDDeserializer struct {
	// Default: []string{"X-Request-Id"}
	Headers []string
}

func (d RequestIDDeserializer) Deserializer(deserialize httpc.DeserializeFunc) httpc.DeserializeFunc {
	return func(req *http.Request) (httpc.DeserializeOutput, httpc.Metadata, error) {
		return d.deserialize(req, deserialize)
	}
}

func (d RequestIDDeserializer) headers() []string {
	if len(d.Headers) == 0 {
		return []string{headerXRequestID}
	}
	return d.Headers
}

func (d RequestIDDeserializer) deserialize(req *http.Request, deserialize httpc.DeserializeFunc) (
	output httpc.DeserializeOutput, md httpc.Metadata, err error,
) {
	output, md, err = deserialize(req)

	resp := output.Response
	if resp == nil {
		return
	}

	for _, h := range d.headers() {
		id := resp.Header.Get(h)
		if id != "" {
			md.Set(mdRequestIDKey{}, id)
			break
		}
	}

	return
}
