package httpc

import (
	"fmt"
	"strings"
)

type InvalidParamError struct {
	Context []string
	Reason  string
}

func (e *InvalidParamError) Error() string {
	var s strings.Builder
	e.writeTo(&s)
	return s.String()
}

func (e *InvalidParamError) writeTo(s *strings.Builder, context ...string) {
	context = append(context, e.Context...)
	s.WriteString(e.Reason)
	if len(context) == 0 {
		return
	}

	s.WriteString(", ")
	s.WriteString(context[0])
	for _, ctx := range context[1:] {
		if !strings.HasPrefix(ctx, "[") {
			s.WriteByte('.')
		}
		s.WriteString(ctx)
	}
	s.WriteByte('.')
}

type InvalidParamsError struct {
	Context []string
	Errs    []InvalidParamError
}

func (e *InvalidParamsError) Err() error {
	if len(e.Errs) == 0 {
		return nil
	}
	return e
}

func (e *InvalidParamsError) Error() string {
	var s strings.Builder
	fmt.Fprintf(&s, "%d validation error(s) found.\n", len(e.Errs))

	for _, err := range e.Errs {
		s.WriteString("- ")
		err.writeTo(&s, e.Context...)
		s.WriteByte('\n')
	}
	return s.String()
}

func (e *InvalidParamsError) AddInvalid(err error, context ...string) {
	if err == nil {
		return
	}

	switch ierr := err.(type) {
	case *InvalidParamError:
		e.addInvalidParamError(ierr, context...)
	case *InvalidParamsError:
		e.addInvalidParamsError(ierr, context...)
	default:
		e.Errs = append(e.Errs, InvalidParamError{
			Context: context,
			Reason:  err.Error(),
		})
	}
}

func (e *InvalidParamsError) AddInvalidWithIndex(err error, index int, context ...string) {
	e.AddInvalid(err, append(context, fmt.Sprintf("[%d]", index))...)
}

func (e *InvalidParamsError) addInvalidParamError(err *InvalidParamError, context ...string) {
	if err == nil {
		return
	}
	err2 := *err
	err2.Context = append(context, err2.Context...)
	e.Errs = append(e.Errs, err2)
}

func (e *InvalidParamsError) addInvalidParamsError(err *InvalidParamsError, context ...string) {
	if err == nil {
		return
	}
	for i := range err.Errs {
		e.addInvalidParamError(&err.Errs[i], context...)
	}
}

func NewParamRequiredError(context ...string) *InvalidParamError {
	return &InvalidParamError{Context: context, Reason: "missing required param"}
}
