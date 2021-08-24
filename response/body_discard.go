package response

import (
	"io"
	"net/http"

	"github.com/go-camp/httpc"
)

const maxBodyDiscardBytes = 2 << 10

// BodyDiscardDeserializer discards response body content,
// if the wrapped DeserializeFunc returns nil error.
type BodyDiscardDeserializer struct {
}

func (d BodyDiscardDeserializer) Deserializer(deserialize httpc.DeserializeFunc) httpc.DeserializeFunc {
	return func(req *http.Request) (httpc.DeserializeOutput, httpc.Metadata, error) {
		return d.deserialize(req, deserialize)
	}
}

func discardResponseBody(resp *http.Response) {
	if resp == nil || resp.Body == nil {
		return
	}

	if resp.ContentLength == -1 || resp.ContentLength <= maxBodyDiscardBytes {
		io.CopyN(io.Discard, resp.Body, maxBodyDiscardBytes)
	}
}

func (d BodyDiscardDeserializer) deserialize(req *http.Request, deserialize httpc.DeserializeFunc) (
	output httpc.DeserializeOutput, md httpc.Metadata, err error,
) {
	output, md, err = deserialize(req)
	if err != nil {
		return
	}

	discardResponseBody(output.Response)

	return
}

// BodyDiscardErrorDeserializer discards response body content,
// if the wrapped DeserializeFunc returns an error.
type BodyDiscardErrorDeserializer struct {
}

func (d BodyDiscardErrorDeserializer) Deserializer(deserialize httpc.DeserializeFunc) httpc.DeserializeFunc {
	return func(req *http.Request) (httpc.DeserializeOutput, httpc.Metadata, error) {
		return d.deserialize(req, deserialize)
	}
}

func (d BodyDiscardErrorDeserializer) deserialize(req *http.Request, deserialize httpc.DeserializeFunc) (
	output httpc.DeserializeOutput, md httpc.Metadata, err error,
) {
	output, md, err = deserialize(req)
	if err == nil {
		return
	}

	discardResponseBody(output.Response)

	return
}
