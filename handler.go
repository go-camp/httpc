package httpc

import (
	"context"
	"fmt"
	"net/http"
)

// Metadata provides storing and reading metadata values.
// Metadata should be used to create features that are shared across HTTP clients.
// Most return values ​​should be stored in output rather than metadata.
type Metadata struct {
	kvs map[interface{}]interface{}
}

func (md Metadata) Get(key interface{}) (value interface{}, ok bool) {
	value, ok = md.kvs[key]
	return
}

func (md *Metadata) Set(key, value interface{}) {
	if md.kvs == nil {
		md.kvs = make(map[interface{}]interface{})
	}
	md.kvs[key] = value
}

// Clone returns a copy of md.
func (md Metadata) Clone() Metadata {
	var kvs map[interface{}]interface{}
	if md.kvs != nil {
		kvs = make(map[interface{}]interface{}, len(md.kvs))
		for k, v := range md.kvs {
			kvs[k] = v
		}
	}
	return Metadata{kvs: kvs}
}

type InitializeFunc func(ctx context.Context, input interface{}) (output interface{}, md Metadata, err error)

type Initializer func(InitializeFunc) InitializeFunc

func ComposeInitializer(initializer ...Initializer) Initializer {
	if len(initializer) == 1 {
		return initializer[0]
	}
	return func(initialize InitializeFunc) InitializeFunc {
		for i := len(initializer) - 1; i >= 0; i-- {
			initialize = initializer[i](initialize)
		}
		return initialize
	}
}

type SerializeInput struct {
	Input   interface{}
	Request *Request
}

type SerializeFunc func(ctx context.Context, input SerializeInput) (output interface{}, md Metadata, err error)

type Serializer func(SerializeFunc) SerializeFunc

func ComposeSerializer(serializer ...Serializer) Serializer {
	if len(serializer) == 1 {
		return serializer[0]
	}
	return func(serialize SerializeFunc) SerializeFunc {
		for i := len(serializer) - 1; i >= 0; i-- {
			serialize = serializer[i](serialize)
		}
		return serialize
	}
}

type BuildFunc func(ctx context.Context, req *Request) (output interface{}, md Metadata, err error)

type Builder func(BuildFunc) BuildFunc

func ComposeBuilder(builder ...Builder) Builder {
	if len(builder) == 1 {
		return builder[0]
	}
	return func(build BuildFunc) BuildFunc {
		for i := len(builder) - 1; i >= 0; i-- {
			build = builder[i](build)
		}
		return build
	}
}

type DeserializeOutput struct {
	Output   interface{}
	Response *http.Response
}

type DeserializeFunc func(req *http.Request) (output DeserializeOutput, md Metadata, err error)

type Deserializer func(DeserializeFunc) DeserializeFunc

func ComposeDeserializer(deserializer ...Deserializer) Deserializer {
	if len(deserializer) == 1 {
		return deserializer[0]
	}
	return func(deserialize DeserializeFunc) DeserializeFunc {
		for i := len(deserializer) - 1; i >= 0; i-- {
			deserialize = deserializer[i](deserialize)
		}
		return deserialize
	}
}

// Handler make a http request and process the response step by step.
//
// Call chain:
//   Initializer -> Serializer -> Builder -> Deserializer -> Do
//
// Return chain:
//   Initializer <- Serializer <- Builder <- Deserializer <- Do
//
// OperationError wraps any of:
//   1. SerializationError
//   2. InvalidParamsError
//   3. RequestSendError
//   4. ResponseError
//   5. others
//
// ResponseError wraps any of:
//   1. DeserializationError
//   2. APIError
//   3. GenericAPIError
//   4. others
type Handler struct {
	// Initializer initializes the input.
	// Examples:
	//   1. Assign default values to the input
	//   2. Validate the input.
	Initializer Initializer

	// Serializer builds the http request based on the input.
	// Exmaples:
	//   1. Encode the input into the request path/header/body.
	Serializer Serializer

	// Builder add extra headers to the request.
	// Exmaples:
	//   1. Set Content-Length header.
	//   2. Set User-Agent header.
	//   3. Retry the request.
	Builder Builder

	// Deserializer decode the response into the output/metadata/err.
	// Examples:
	//   1. Decode response body into the output.
	//   2. Decode response body into the GenericAPIError.
	//   3. Set raw response to the metadata.
	Deserializer Deserializer

	// Do is http.Client's Do method.
	Do func(req *http.Request) (*http.Response, error)
}

type deserializeWrapper struct {
	Do func(req *http.Request) (*http.Response, error)
}

func (w deserializeWrapper) Deserialize(req *http.Request) (output DeserializeOutput, md Metadata, err error) {
	var resp *http.Response
	resp, err = w.Do(req)
	output.Response = resp
	return
}

type buildWrapper struct {
	Deserialize DeserializeFunc
}

func (w buildWrapper) Build(ctx context.Context, req *Request) (output interface{}, md Metadata, err error) {
	var dout DeserializeOutput
	dout, md, err = w.Deserialize(req.Build())
	output = dout.Output
	return
}

type serializeWrapper struct {
	Build BuildFunc
}

func (w serializeWrapper) Serialize(ctx context.Context, input SerializeInput) (output interface{}, md Metadata, err error) {
	return w.Build(ctx, input.Request)
}

type initializeWrapper struct {
	Serialize SerializeFunc
}

func (w initializeWrapper) Initialize(ctx context.Context, input interface{}) (output interface{}, md Metadata, err error) {
	return w.Serialize(ctx, SerializeInput{Input: input})
}

func (h Handler) Handle(ctx context.Context, input interface{}) (output interface{}, md Metadata, err error) {
	deserialize := h.Deserializer(deserializeWrapper{Do: h.Do}.Deserialize)
	build := h.Builder(buildWrapper{Deserialize: deserialize}.Build)
	serialize := h.Serializer(serializeWrapper{Build: build}.Serialize)
	initialize := h.Initializer(initializeWrapper{Serialize: serialize}.Initialize)
	return initialize(ctx, input)
}

// OperationError wraps error returns by Handle() function.
type OperationError struct {
	Service   string
	Operation string
	Err       error
}

func (e *OperationError) ServiceName() string { return e.Service }

func (e *OperationError) OperationName() string { return e.Operation }

func (e *OperationError) Unwrap() error { return e.Err }

func (e *OperationError) Error() string {
	return fmt.Sprintf("%s operation error: %s, %v", e.Service, e.Operation, e.Err)
}
