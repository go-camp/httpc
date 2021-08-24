package response

import (
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/go-camp/httpc"
)

func TestBodyDiscardDeserializer(t *testing.T) {
	testCases := []struct {
		Name     string
		Response *http.Response
		Err      error

		ExpectRead bool
	}{
		{
			Name:     "response with error",
			Response: &http.Response{},
			Err:      errors.New("err"),

			ExpectRead: false,
		},
		{
			Name:     "response without error",
			Response: &http.Response{},

			ExpectRead: true,
		},
	}

	for _, tc := range testCases {
		var read bool
		deserialize := BodyDiscardDeserializer{}.Deserializer(
			func(req *http.Request) (output httpc.DeserializeOutput, md httpc.Metadata, err error) {
				output.Response = tc.Response
				if output.Response != nil {
					output.Response.Body = &testReadCloser{
						R: func(p []byte) (int, error) {
							read = true
							return 0, io.EOF
						},
					}
				}
				return output, md, tc.Err
			},
		)
		_, _, err := deserialize(newNopHTTPRequest())
		if err != tc.Err {
			t.Fatalf("expect err is %s, got %s", tc.Err, err)
		}
		if tc.ExpectRead != read {
			t.Fatalf("expect read is %v, got %v", tc.ExpectRead, read)
		}
	}
}

func TestBodyDiscardErrorDeserializer(t *testing.T) {
	testCases := []struct {
		Name     string
		Response *http.Response
		Err      error

		ExpectRead bool
	}{
		{
			Name:     "response with error",
			Response: &http.Response{},
			Err:      errors.New("err"),

			ExpectRead: true,
		},
		{
			Name:     "response without error",
			Response: &http.Response{},

			ExpectRead: false,
		},
	}

	for _, tc := range testCases {
		var read bool
		deserialize := BodyDiscardErrorDeserializer{}.Deserializer(
			func(req *http.Request) (output httpc.DeserializeOutput, md httpc.Metadata, err error) {
				output.Response = tc.Response
				if output.Response != nil {
					output.Response.Body = &testReadCloser{
						R: func(p []byte) (int, error) {
							read = true
							return 0, io.EOF
						},
					}
				}
				return output, md, tc.Err
			},
		)
		_, _, err := deserialize(newNopHTTPRequest())
		if err != tc.Err {
			t.Fatalf("expect err is %s, got %s", tc.Err, err)
		}
		if tc.ExpectRead != read {
			t.Fatalf("expect read is %v, got %v", tc.ExpectRead, read)
		}
	}
}
