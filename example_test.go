package httpc_test

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/go-camp/httpc"
	"github.com/go-camp/httpc/request"
	"github.com/go-camp/httpc/response"
)

func ExampleHandler_order() {
	type Input struct{ Name, Password string }
	type Output struct{}

	randPassword := func() (string, error) {
		var buf [6]byte
		_, err := io.ReadFull(rand.Reader, buf[:])
		if err != nil {
			return "", err
		}
		return hex.EncodeToString(buf[:]), nil
	}

	validateInput := func(input *Input) error {
		var verr = &httpc.InvalidParamsError{
			Context: []string{"Input"},
		}

		if input.Name == "" {
			verr.AddInvalid(httpc.NewParamRequiredError("Name"))
		}
		if input.Password == "" {
			verr.AddInvalid(httpc.NewParamRequiredError("Password"))
		}

		return verr.Err()
	}

	h := httpc.Handler{
		Initializer: httpc.ComposeInitializer(
			request.ServiceOperationNameInitializer{
				ServiceName:   "example.client",
				OperationName: "Get",
			}.Initializer,

			// http.OperationError
			request.WrapOperationErrorInitializer{}.Initializer,

			// initialize input here.
			func(initialize httpc.InitializeFunc) httpc.InitializeFunc {
				return func(ctx context.Context, input interface{}) (output interface{}, md httpc.Metadata, err error) {
					in, ok := input.(*Input)
					if !ok {
						return output, md, fmt.Errorf("expect %T type, got %T", &Input{}, input)
					}

					if in == nil {
						in = &Input{}
					}

					if in.Password == "" {
						password, err := randPassword()
						if err != nil {
							return output, md, err
						}
						in.Password = password
					}

					return initialize(ctx, in)
				}
			},

			// validate input here.
			// httpc.InvalidParamsError
			func(initialize httpc.InitializeFunc) httpc.InitializeFunc {
				return func(ctx context.Context, input interface{}) (output interface{}, md httpc.Metadata, err error) {
					in, ok := input.(*Input)
					if !ok {
						return output, md, fmt.Errorf("expect %T type, got %T", &Input{}, input)
					}

					if err := validateInput(in); err != nil {
						return output, md, err
					}

					return initialize(ctx, input)
				}
			},
		),

		Serializer: httpc.ComposeSerializer(
			// build url and serialize input here.
			// httpc.SerializationError
			func(serialize httpc.SerializeFunc) httpc.SerializeFunc {
				return func(ctx context.Context, input httpc.SerializeInput) (output interface{}, md httpc.Metadata, err error) {
					input.Request = &httpc.Request{
						Request: &http.Request{
							URL:           &url.URL{},
							Header:        http.Header{},
							ContentLength: 0,
						},
						Body: http.NoBody,
					}
					return serialize(ctx, input)
				}
			},
		),

		Builder: httpc.ComposeBuilder(
			request.UserAgentBuilder{
				UserAgent: request.UserAgent{
					Name:    "go",
					Version: "1.16",
				},
			}.Builder,

			request.RequestIDBuilder{}.Builder,
			request.ContentLengthBuilder{}.Builder,
			request.ContentMD5Builder{}.Builder,

			request.RetryBuilder{
				Retryer: request.DefaultRetryer,
			}.Builder,
		),

		Deserializer: httpc.ComposeDeserializer(
			// http.ResponseError
			response.WrapResponseErrorDeserializer{}.Deserializer,

			// http.RequestSendError
			response.WrapRequestErrorDeserializer{}.Deserializer,

			response.BodyCloseErrorDeserializer{}.Deserializer,
			response.BodyCloseDeserializer{}.Deserializer,

			response.BodyDiscardErrorDeserializer{}.Deserializer,
			response.BodyDiscardDeserializer{}.Deserializer,

			response.DateDeserializer{}.Deserializer,
			response.ResponseDeserializer{}.Deserializer,
			response.RequestIDDeserializer{}.Deserializer,

			// process response and build output here.
			func(deserialize httpc.DeserializeFunc) httpc.DeserializeFunc {
				return func(req *http.Request) (output httpc.DeserializeOutput, md httpc.Metadata, err error) {
					output, md, err = deserialize(req)
					if err != nil || output.Response == nil {
						return
					}

					response := output.Response
					if response.StatusCode < 200 || response.StatusCode >= 300 {
						// process api error here
						// httpc.APIError, httpc.GenericAPIError
						// httpc.DeserializationError
						return output, md, &httpc.GenericAPIError{Code: http.StatusText(response.StatusCode)}
					}

					// httpc.DeserializationError
					output.Output = &Output{}
					return
				}
			},
		),

		Do: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Header:     http.Header{},
				Body:       http.NoBody,
			}, nil
		},
	}

	output, _, err := h.Handle(context.Background(), &Input{Name: "name"})
	fmt.Println("output:", output)
	fmt.Println("err:", err)
	// output:
	// output: &{}
	// err: <nil>
}
