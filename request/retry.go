package request

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/go-camp/httpc"
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

var DefaultRetryer Retryer

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
