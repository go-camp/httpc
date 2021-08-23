package httpc

import (
	"fmt"
	"net/http"
)

// ResponseError wraps any errors that occur while processing a http.Response.
// Examples:
//   1. DeserializationError
//   2. APIError/GenericAPIError
type ResponseError struct {
	Response  *http.Response
	RequestID string
	Err       error
}

func (e *ResponseError) Error() string {
	if e.RequestID == "" {
		return fmt.Sprintf("http response error, status code: %d, %v",
			e.Response.StatusCode, e.Err)
	}
	return fmt.Sprintf("http response error, status code: %d, request id: %s, %v",
		e.Response.StatusCode, e.RequestID, e.Err)
}

func (e *ResponseError) Unwrap() error {
	return e.Err
}

func (e *ResponseError) HTTPResponse() *http.Response {
	return e.Response
}

func (e *ResponseError) HTTPStatusCode() int {
	return e.Response.StatusCode
}

func (e *ResponseError) HTTPRequestID() string {
	return e.RequestID
}

// ResponseError wraps any errors that occur while deserializing a http.Response.
type DeserializationError struct {
	Err      error
	Snapshot []byte
}

func (e *DeserializationError) Error() string {
	return fmt.Sprintf("deserialization failed, %v", e.Err)
}

func (e *DeserializationError) Unwrap() error { return e.Err }

//go:generate stringer -type=ErrorFault -trimprefix=ErrorFault
type ErrorFault int

const (
	ErrorFaultUnknown ErrorFault = iota
	ErrorFaultServer
	ErrorFaultClient
)

// APIError represents a kind of error deserialized from http.Response.
// Examples:
//   1. 404 -> NotFoundError
//   2. 401 -> UnauthorizedError
//   3. {"Code": "Expired", "Msg": "card token expired"} ->
//      GenericAPIError{Code: "Expired", Message: "card token expired", Fault: ErrorFaultClient}
type APIError interface {
	error

	ErrorCode() string
	ErrorMessage() string
	ErrorFault() ErrorFault
}

// GenericAPIError is a generic APIError implemention.
type GenericAPIError struct {
	Code    string
	Message string
	Fault   ErrorFault
}

var _ APIError = (*GenericAPIError)(nil)

func (e *GenericAPIError) Error() string {
	return fmt.Sprintf("api error: %s, %s", e.Code, e.Message)
}

func (e *GenericAPIError) ErrorCode() string {
	return e.Code
}

func (e *GenericAPIError) ErrorMessage() string {
	return e.Message
}

func (e *GenericAPIError) ErrorFault() ErrorFault {
	return e.Fault
}
