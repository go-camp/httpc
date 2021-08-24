package request

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/go-camp/httpc"
)

type testRetryer struct {
	M int
	D func(attempt int) time.Duration
	C func(err error) Retryable
}

func (r *testRetryer) MaxAttempts() int {
	return r.M
}
func (r *testRetryer) Delay(attempt int) time.Duration {
	return r.D(attempt)
}
func (r *testRetryer) Check(err error) Retryable {
	return r.C(err)
}

func TestRetryBuilder(t *testing.T) {
	testCases := []struct {
		Name               string
		MaxAttempts        int
		AttemptsErrorCount int
		DelayDuration      time.Duration
		Retryable          Retryable
		Body               io.Reader

		ExpectError string
	}{
		{
			Name:               "max attempts exhausted",
			MaxAttempts:        3,
			AttemptsErrorCount: 3,
			DelayDuration:      0,
			Retryable:          RetryableYes,
			Body:               bytes.NewReader([]byte("1")),

			ExpectError: "request retry finalizer, attempts 3, max attempts 3 exhausted, err 3",
		},
		{
			Name:               "no retryable",
			MaxAttempts:        3,
			AttemptsErrorCount: 3,
			DelayDuration:      0,
			Retryable:          RetryableNo,
			Body:               bytes.NewReader([]byte("1")),

			ExpectError: "request retry finalizer, attempts 1, retryable No, err 1",
		},
		{
			Name:               "ctx canceled",
			MaxAttempts:        3,
			AttemptsErrorCount: 3,
			DelayDuration:      110 * time.Millisecond,
			Retryable:          RetryableYes,
			Body:               bytes.NewReader([]byte("1")),

			ExpectError: "request retry finalizer, attempts 2, sleep canceled, context deadline exceeded",
		},
		{
			Name:               "ctx canceled",
			MaxAttempts:        3,
			AttemptsErrorCount: 3,
			DelayDuration:      0,
			Retryable:          RetryableYes,
			Body:               bytes.NewBuffer([]byte("1")),

			ExpectError: "request retry finalizer, attempts 2, body rewind failed, request body cannot be rewinded",
		},
		{
			Name:               "final success",
			MaxAttempts:        3,
			AttemptsErrorCount: 2,
			DelayDuration:      0,
			Retryable:          RetryableYes,
			Body:               bytes.NewReader([]byte("1")),

			ExpectError: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			var buildCount int
			build := RetryBuilder{
				Retryer: &testRetryer{
					M: tc.MaxAttempts,
					D: func(attempt int) time.Duration {
						return tc.DelayDuration
					},
					C: func(error) Retryable {
						return tc.Retryable
					},
				},
			}.Builder(
				func(ctx context.Context, req *httpc.Request) (output interface{}, md httpc.Metadata, err error) {
					buildCount++
					if buildCount <= tc.AttemptsErrorCount {
						err = fmt.Errorf("err %d", buildCount)
					}
					return
				},
			)

			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()
			req := &httpc.Request{
				Request: &http.Request{
					URL:           &url.URL{},
					Header:        http.Header{},
					ContentLength: -1,
				},
				Body: tc.Body,
			}
			_, _, err := build(ctx, req)
			if tc.ExpectError != "" {
				if err == nil {
					t.Fatalf("expect err is %s, got none", tc.ExpectError)
				} else {
					if tc.ExpectError != err.Error() {
						t.Fatalf("expect err is %s, got %s", tc.ExpectError, err)
					}
				}
			} else {
				if err != nil {
					t.Fatalf("expect no err, got %s", err)
				}
			}
		})
	}
}

type testRetryableError struct {
	Retryable bool
}

func (e *testRetryableError) Error() string {
	return "retryable error"
}
func (e *testRetryableError) RetryableError() bool {
	return e.Retryable
}

func TestBasicRetryer(t *testing.T) {
	retryer := DefaultRetryer
	maxAttempts := retryer.MaxAttempts()
	if maxAttempts != 3 {
		t.Fatalf("expect max attempts is %d, got %d", 3, maxAttempts)
	}

	delay := retryer.Delay(1)
	if delay <= 0 {
		t.Fatalf("expect delay greater than 0, got %d", delay)
	}

	testCases := []struct {
		Name string
		Err  error

		ExpectRetryable Retryable
	}{
		{
			Name: "errors new",
			Err:  errors.New("temp err"),

			ExpectRetryable: RetryableUnknown,
		},
		{
			Name: "retryable error",
			Err:  &testRetryableError{Retryable: false},

			ExpectRetryable: RetryableNo,
		},
		{
			Name: "connection error",
			Err:  errors.New("connection reset"),

			ExpectRetryable: RetryableYes,
		},
		{
			Name: "response error",
			Err: &httpc.ResponseError{
				Response: &http.Response{
					Header:     http.Header{},
					StatusCode: http.StatusServiceUnavailable,
				},
				Err: errors.New("temp unavailable"),
			},

			ExpectRetryable: RetryableYes,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			retryable := retryer.Check(tc.Err)
			if tc.ExpectRetryable != retryable {
				t.Fatalf("expect retryable is %s, got %s", tc.ExpectRetryable, retryable)
			}
		})
	}
}
