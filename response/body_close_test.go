package response

import (
	"errors"
	"net/http"
	"testing"

	"github.com/go-camp/httpc"
)

type testReadCloser struct {
	R func(p []byte) (int, error)
	C func() error
}

func (r *testReadCloser) Read(p []byte) (int, error) {
	return r.R(p)
}

func (r *testReadCloser) Close() error {
	return r.C()
}

func TestBodyCloseDeserializer(t *testing.T) {
	testCases := []struct {
		Name     string
		Response *http.Response
		Err      error

		ExpectClosed bool
	}{
		{
			Name:     "response with err",
			Response: &http.Response{},
			Err:      errors.New("err"),

			ExpectClosed: false,
		},
		{
			Name:     "response without err",
			Response: &http.Response{},

			ExpectClosed: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			var closed bool
			deserialize := BodyCloseDeserializer{}.Deserializer(
				func(req *http.Request) (output httpc.DeserializeOutput, md httpc.Metadata, err error) {
					output.Response = tc.Response
					if output.Response != nil {
						output.Response.Body = &testReadCloser{
							C: func() error {
								closed = true
								return nil
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
			if tc.ExpectClosed != closed {
				t.Fatalf("expect closed is %v, got %v", tc.ExpectClosed, closed)
			}
		})
	}
}

func TestBodyCloseErrorDeserializer(t *testing.T) {
	testCases := []struct {
		Name     string
		Response *http.Response
		Err      error

		ExpectClosed bool
	}{
		{
			Name:     "response with err",
			Response: &http.Response{},
			Err:      errors.New("err"),

			ExpectClosed: true,
		},
		{
			Name:     "response without err",
			Response: &http.Response{},

			ExpectClosed: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			var closed bool
			deserialize := BodyCloseErrorDeserializer{}.Deserializer(
				func(req *http.Request) (output httpc.DeserializeOutput, md httpc.Metadata, err error) {
					output.Response = tc.Response
					if output.Response != nil {
						output.Response.Body = &testReadCloser{
							C: func() error {
								closed = true
								return nil
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
			if tc.ExpectClosed != closed {
				t.Fatalf("expect closed is %v, got %v", tc.ExpectClosed, closed)
			}
		})
	}
}
