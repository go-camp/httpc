package response

import (
	"errors"
	"net/http"
	"net/url"
	"testing"

	"github.com/go-camp/httpc"
)

func newNopHTTPRequest() *http.Request {
	return &http.Request{
		URL:           &url.URL{},
		ContentLength: 0,
		Body:          http.NoBody,
	}
}

func TestWrapRequestErrorDeserializer(t *testing.T) {
	testCases := []struct {
		Name   string
		Output httpc.DeserializeOutput
		Err    error

		ExpectError string
	}{
		{
			Name: "without error",
		},
		{
			Name: "err without response",
			Err:  errors.New("request error"),

			ExpectError: "request send failed, request error",
		},
		{
			Name: "err with response",
			Err:  errors.New("response error"),
			Output: httpc.DeserializeOutput{
				Response: &http.Response{
					StatusCode: http.StatusNotFound,
				},
			},

			ExpectError: "response error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			deserialize := WrapRequestErrorDeserializer{}.Deserializer(
				func(req *http.Request) (output httpc.DeserializeOutput, md httpc.Metadata, err error) {
					return tc.Output, md, tc.Err
				},
			)
			_, _, err := deserialize(newNopHTTPRequest())
			if tc.ExpectError == "" {
				if err != nil {
					t.Fatalf("expect no err, got %s", err)
				}
			} else {
				if err == nil {
					t.Fatalf("expect err is %s, got nil", tc.ExpectError)
				} else {
					if tc.ExpectError != err.Error() {
						t.Fatalf("expect err is %s, got %s", tc.ExpectError, err.Error())
					}
				}
			}
		})
	}
}

func TestWrapResponseErrorDeserializer(t *testing.T) {
	testCases := []struct {
		Name   string
		SetMD  func(md *httpc.Metadata)
		Output httpc.DeserializeOutput
		Err    error

		ExpectError string
	}{
		{
			Name: "without error",
		},
		{
			Name: "err without response",
			Err:  errors.New("request error"),

			ExpectError: "request error",
		},
		{
			Name: "err with response",
			Err:  errors.New("response error"),
			Output: httpc.DeserializeOutput{
				Response: &http.Response{
					StatusCode: http.StatusNotFound,
				},
			},

			ExpectError: "http response error, status code: 404, response error",
		},
		{
			Name: "err with response request id",
			Err:  errors.New("response error"),
			SetMD: func(md *httpc.Metadata) {
				md.Set(mdRequestIDKey{}, "12ca095b1798410e91b0f5f2ec6b6e05")
			},
			Output: httpc.DeserializeOutput{
				Response: &http.Response{
					StatusCode: http.StatusNotFound,
				},
			},

			ExpectError: "http response error, status code: 404, request id: 12ca095b1798410e91b0f5f2ec6b6e05, response error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			deserialize := WrapResponseErrorDeserializer{}.Deserializer(
				func(req *http.Request) (output httpc.DeserializeOutput, md httpc.Metadata, err error) {
					if tc.SetMD != nil {
						tc.SetMD(&md)
					}
					return tc.Output, md, tc.Err
				},
			)
			_, _, err := deserialize(newNopHTTPRequest())
			if tc.ExpectError == "" {
				if err != nil {
					t.Fatalf("expect no err, got %s", err)
				}
			} else {
				if err == nil {
					t.Fatalf("expect err is %s, got nil", tc.ExpectError)
				} else {
					if tc.ExpectError != err.Error() {
						t.Fatalf("expect err is %s, got %s", tc.ExpectError, err.Error())
					}
				}
			}
		})
	}
}
