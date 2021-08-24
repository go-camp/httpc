package response

import (
	"net/http"

	"github.com/go-camp/httpc"
)

// BodyCloseDeserializer closes response body, if the wrapped DeserializeFunc returns a nil error.
type BodyCloseDeserializer struct {
}

func (d BodyCloseDeserializer) Deserializer(deserialize httpc.DeserializeFunc) httpc.DeserializeFunc {
	return func(req *http.Request) (httpc.DeserializeOutput, httpc.Metadata, error) {
		return d.deserialize(req, deserialize)
	}
}

func (d BodyCloseDeserializer) deserialize(req *http.Request, deserialize httpc.DeserializeFunc) (
	output httpc.DeserializeOutput, md httpc.Metadata, err error,
) {
	output, md, err = deserialize(req)
	if err != nil {
		return
	}

	resp := output.Response
	if resp == nil || resp.Body == nil {
		return
	}
	resp.Body.Close()
	return
}

// BodyCloseDeserializer closes response body, if the wrapped DeserializeFunc returns a error.
type BodyCloseErrorDeserializer struct {
}

func (d BodyCloseErrorDeserializer) Deserializer(deserialize httpc.DeserializeFunc) httpc.DeserializeFunc {
	return func(req *http.Request) (httpc.DeserializeOutput, httpc.Metadata, error) {
		return d.deserialize(req, deserialize)
	}
}

func (d BodyCloseErrorDeserializer) deserialize(req *http.Request, deserialize httpc.DeserializeFunc) (
	output httpc.DeserializeOutput, md httpc.Metadata, err error,
) {
	output, md, err = deserialize(req)
	if err == nil {
		return
	}

	resp := output.Response
	if resp == nil || resp.Body == nil {
		return
	}
	resp.Body.Close()
	return
}
