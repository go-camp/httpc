package request

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-camp/httpc"
	"github.com/go-camp/retry"
)

//go:generate stringer -type=Retryable -trimprefix=Retryable
type Retryable int

const (
	RetryableUnknown Retryable = iota
	RetryableYes
	RetryableNo
)

type Retryer interface {
	// MaxAttempts returns the maximum number of attempts.
	MaxAttempts() int

	// Delayer calculates the delay before the next attempt after this attempt.
	Delay(attempt int) time.Duration

	// Check checks if err can be retried.
	Check(err error) Retryable
}

type RetryableChecker interface {
	Check(err error) Retryable
}

type RetryableCheckers []RetryableChecker

func (r RetryableCheckers) Check(err error) Retryable {
	for _, c := range r {
		if retryable := c.Check(err); retryable != RetryableUnknown {
			return retryable
		}
	}
	return RetryableUnknown
}

var DefaultRetryableChecker = RetryableCheckers{
	RetryableErrorChecker{},
	RetryableConnectionErrorChecker{},
	RetryableHTTPStatusCodeChecker{
		Codes: DefaultRetryableHTTPStatusCodes,
	},
}

type RetryableErrorChecker struct{}

func (r RetryableErrorChecker) Check(err error) Retryable {
	var v interface{ RetryableError() bool }
	if !errors.As(err, &v) {
		return RetryableUnknown
	}
	if v.RetryableError() {
		return RetryableYes
	}
	return RetryableNo
}

type RetryableConnectionErrorChecker struct{}

func (r RetryableConnectionErrorChecker) Check(err error) Retryable {
	if err == nil {
		return RetryableUnknown
	}

	var tempErr interface{ Temporary() bool }
	var timeoutErr interface{ Timeout() bool }
	var urlErr *url.Error
	var netOpErr *net.OpError
	switch {
	case strings.Contains(err.Error(), "connection reset"):
		return RetryableYes
	case errors.As(err, &urlErr):
		if strings.Contains(urlErr.Error(), "connection refused") {
			return RetryableYes
		} else {
			return r.Check(errors.Unwrap(urlErr))
		}
	case errors.As(err, &netOpErr):
		if strings.EqualFold(netOpErr.Op, "dial") || netOpErr.Temporary() {
			return RetryableYes
		} else {
			return r.Check(errors.Unwrap(netOpErr))
		}
	case errors.As(err, &tempErr) && tempErr.Temporary():
		return RetryableYes
	case errors.As(err, &timeoutErr) && timeoutErr.Timeout():
		return RetryableYes
	default:
		return RetryableUnknown
	}
}

type RetryableHTTPStatusCodeChecker struct {
	Codes map[int]struct{}
}

func (r RetryableHTTPStatusCodeChecker) Check(err error) Retryable {
	var v interface{ HTTPStatusCode() int }
	if !errors.As(err, &v) {
		return RetryableUnknown
	}
	_, ok := r.Codes[v.HTTPStatusCode()]
	if ok {
		return RetryableYes
	}
	return RetryableUnknown
}

var DefaultRetryableHTTPStatusCodes = map[int]struct{}{
	http.StatusInternalServerError: {},
	http.StatusBadGateway:          {},
	http.StatusServiceUnavailable:  {},
	http.StatusGatewayTimeout:      {},
}

type RetryableErrorCodeChecker struct {
	Codes map[string]struct{}
}

func (r RetryableErrorCodeChecker) Check(err error) Retryable {
	var v interface{ ErrorCode() string }
	if !errors.As(err, &v) {
		return RetryableUnknown
	}
	_, ok := r.Codes[v.ErrorCode()]
	if ok {
		return RetryableYes
	}
	return RetryableUnknown
}

const DefaultRetryMaxAttempts = 3

var DefaultRetryDelayer = retry.ExpDelayer{
	Initial:    1 * time.Second,
	Multiplier: 2,
	Max:        20 * time.Second,
	Rand:       50,
}.Delay

var NopRetryDelayer = retry.NopDelayer{}.Delay

type BasicRetryerOptions struct {
	// If MaxAttempts is 0, then DefaultRetryMaxAttempts is used.
	MaxAttempts int
	// If Delayer is nil, then DefaultRetryDelayer is used.
	Delayer func(attempt int) time.Duration
	// If ErrorChecker is nil, then ErrorChecker is used.
	ErrorChecker RetryableChecker
}

func (o BasicRetryerOptions) Copy() BasicRetryerOptions {
	return o
}

type BasicRetryer struct {
	Options BasicRetryerOptions
}

func (r BasicRetryer) WithOptions(optFns ...func(*BasicRetryerOptions)) BasicRetryer {
	opts := r.Options.Copy()
	for _, optFn := range optFns {
		optFn(&opts)
	}
	return BasicRetryer{Options: opts}
}

func (r BasicRetryer) MaxAttempts() int {
	if r.Options.MaxAttempts == 0 {
		return DefaultRetryMaxAttempts
	}
	return r.Options.MaxAttempts
}

func (r BasicRetryer) delayer() func(int) time.Duration {
	if r.Options.Delayer == nil {
		return DefaultRetryDelayer
	}
	return r.Options.Delayer
}

func (r BasicRetryer) Delay(attempt int) time.Duration {
	return r.delayer()(attempt)
}

func (r BasicRetryer) checker() RetryableChecker {
	if r.Options.ErrorChecker == nil {
		return DefaultRetryableChecker
	}
	return r.Options.ErrorChecker
}

func (r BasicRetryer) Check(err error) Retryable {
	return r.checker().Check(err)
}

var DefaultRetryer Retryer = BasicRetryer{
	Options: BasicRetryerOptions{
		MaxAttempts:  DefaultRetryMaxAttempts,
		Delayer:      DefaultRetryDelayer,
		ErrorChecker: DefaultRetryableChecker,
	},
}

type RetryBuilder struct {
	Retryer Retryer
}

func (d RetryBuilder) retryer() Retryer {
	if d.Retryer == nil {
		return DefaultRetryer
	}
	return d.Retryer
}

func (d RetryBuilder) Builder(build httpc.BuildFunc) httpc.BuildFunc {
	return func(ctx context.Context, req *httpc.Request) (interface{}, httpc.Metadata, error) {
		return d.build(ctx, req, build)
	}
}

type retryError struct {
	Attempts int
	Message  string
	Err      error
}

func (e *retryError) Error() string {
	return fmt.Sprintf("request retry finalizer, attempts %d, %s, %v", e.Attempts, e.Message, e.Err)
}

func (e *retryError) Unwrap() error {
	return e.Err
}

func sleep(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

func (d RetryBuilder) build(ctx context.Context, req *httpc.Request, finalize httpc.BuildFunc) (
	output interface{}, md httpc.Metadata, err error,
) {
	var rewind func() error
	body := req.Body
	if body == nil || body == http.NoBody {
		rewind = func() error { return nil }
	} else if _, ok := body.(io.Seeker); ok {
		var rr *rewindReader
		rr, err = newRewindReader(body)
		if err != nil {
			err = &retryError{Message: "new rewind reader failed", Err: err}
			return
		}
		rewind = func() error {
			return rr.Rewind()
		}
	} else {
		rewind = func() error {
			return errors.New("request body cannot be rewinded")
		}
	}

	retryer := d.retryer()
	maxAttempts := retryer.MaxAttempts()
	attempts := 1
	var delay time.Duration
	for {
		output, md, err = finalize(ctx, req.Clone(req.Context()))
		if err == nil {
			return
		}
		if maxAttempts > 0 && attempts >= maxAttempts {
			err = &retryError{
				Attempts: attempts,
				Message:  fmt.Sprintf("max attempts %d exhausted", maxAttempts),
				Err:      err,
			}
			return
		}
		retryable := retryer.Check(err)
		if retryable != RetryableYes {
			err = &retryError{
				Attempts: attempts,
				Message:  fmt.Sprintf("retryable %s", retryable),
				Err:      err,
			}
			return
		}

		attempts++
		delay = retryer.Delay(attempts)
		if err = sleep(ctx, delay); err != nil {
			err = &retryError{
				Attempts: attempts,
				Message:  "sleep canceled",
				Err:      err,
			}
			return
		}
		if err = rewind(); err != nil {
			err = &retryError{Attempts: attempts, Message: "body rewind failed", Err: err}
			return
		}
	}
}
