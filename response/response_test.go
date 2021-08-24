package response

import (
	"net/http"
	"testing"

	"github.com/go-camp/httpc"
)

func TestResponseDeserializer(t *testing.T) {
	expectResp := &http.Response{}
	deserialize := ResponseDeserializer{}.Deserializer(
		func(req *http.Request) (output httpc.DeserializeOutput, md httpc.Metadata, err error) {
			output.Response = expectResp
			return output, md, nil
		},
	)
	_, md, err := deserialize(newNopHTTPRequest())
	if err != nil {
		t.Fatalf("expect no err, got %s", err)
	}
	resp, ok := GetResponse(md)
	if !ok {
		t.Fatal("expect ok is true")
	}
	if expectResp != resp {
		t.Fatalf("expect response is %p, got %p", expectResp, resp)
	}
}
