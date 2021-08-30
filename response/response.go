package response

import (
	"net/http"

	"github.com/go-camp/httpc"
)

type mdResponseKey struct{}

// GetResponse gets response from metadata.
func GetResponse(md httpc.Metadata) *http.Response {
	e := md.Get(mdResponseKey{})
	resp, _ := e.(*http.Response)
	return resp
}

// ResponseDeserializer sets response(*http.Response) to metadata.
type ResponseDeserializer struct {
}

func (d ResponseDeserializer) Deserializer(deserialize httpc.DeserializeFunc) httpc.DeserializeFunc {
	return func(req *http.Request) (httpc.DeserializeOutput, httpc.Metadata, error) {
		return d.deserialize(req, deserialize)
	}
}

func (d ResponseDeserializer) deserialize(req *http.Request, deserialize httpc.DeserializeFunc) (
	output httpc.DeserializeOutput, md httpc.Metadata, err error,
) {
	output, md, err = deserialize(req)
	resp := output.Response
	if resp == nil {
		return
	}
	md.Set(mdResponseKey{}, resp)
	return
}
