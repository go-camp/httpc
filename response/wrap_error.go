package response

import (
	"net/http"

	"github.com/go-camp/httpc"
)

// WrapRequestErrorDeserializer wraps err as RequestSendError if Response is nil.
type WrapRequestErrorDeserializer struct {
}

func (d WrapRequestErrorDeserializer) Deserializer(deserialize httpc.DeserializeFunc) httpc.DeserializeFunc {
	return func(req *http.Request) (httpc.DeserializeOutput, httpc.Metadata, error) {
		return d.deserialize(req, deserialize)
	}
}

func (d WrapRequestErrorDeserializer) deserialize(req *http.Request, deserialize httpc.DeserializeFunc) (
	output httpc.DeserializeOutput, md httpc.Metadata, err error,
) {
	output, md, err = deserialize(req)
	if err == nil || output.Response != nil {
		return
	}

	err = &httpc.RequestSendError{Err: err}
	return
}

// WrapRequestErrorDeserializer wraps err as ResponseError, if Response is not nil.
type WrapResponseErrorDeserializer struct {
}

func (d WrapResponseErrorDeserializer) Deserializer(deserialize httpc.DeserializeFunc) httpc.DeserializeFunc {
	return func(req *http.Request) (httpc.DeserializeOutput, httpc.Metadata, error) {
		return d.deserialize(req, deserialize)
	}
}

func (d WrapResponseErrorDeserializer) deserialize(req *http.Request, deserialize httpc.DeserializeFunc) (
	output httpc.DeserializeOutput, md httpc.Metadata, err error,
) {
	output, md, err = deserialize(req)
	if err == nil || output.Response == nil {
		return
	}

	reqID := GetRequestID(md)
	err = &httpc.ResponseError{
		Response:  output.Response,
		RequestID: reqID,
		Err:       err,
	}
	return
}
