package response

import (
	"net/http"
	"testing"

	"github.com/go-camp/httpc"
)

func TestRequestIDDeserializer(t *testing.T) {
	expectRequestID := "12ca095b1798410e91b0f5f2ec6b6e05"
	deserialize := RequestIDDeserializer{}.Deserializer(
		func(req *http.Request) (output httpc.DeserializeOutput, md httpc.Metadata, err error) {
			output.Response = &http.Response{
				Header: http.Header{
					"X-Request-Id": []string{expectRequestID},
				},
			}
			return output, md, nil
		},
	)
	_, md, err := deserialize(newNopHTTPRequest())
	if err != nil {
		t.Fatalf("expect no err, got %s", err)
	}
	requestID := GetRequestID(md)
	if expectRequestID != requestID {
		t.Fatalf("expect request id is %s, got %s", expectRequestID, requestID)
	}
}
