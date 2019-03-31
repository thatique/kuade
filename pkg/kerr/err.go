package kerr

import (
	"context"
	"fmt"
	"io"
	"reflect"

	"golang.org/x/xerrors"
)

// An ErrorCode describes the error's category.
type ErrorCode int

const (
	// Returned by the Code function on a nil error. It is not a valid
	// code for an error.
	OK ErrorCode = iota

	// The error could not be categorized.
	Unknown

	// The resource was not found.
	NotFound

	// The resource exists, but it should not.
	AlreadyExists

	// A value given to an API is incorrect.
	InvalidArgument

	// Something unexpected happened. Internal errors always indicate
	// bugs in the API (or possibly the underlying provider).
	Internal

	// The feature is not implemented.
	Unimplemented

	// The system was in the wrong state.
	FailedPrecondition

	// The caller does not have permission to execute the specified operation.
	PermissionDenied

	// Some resource has been exhausted, typically because a service resource limit
	// has been reached.
	ResourceExhausted

	// The operation was canceled.
	Canceled

	// The operation timed out.
	DeadlineExceeded
)

type Error struct {
	Code  ErrorCode
	msg   string
	frame xerrors.Frame
	err   error
}

func (e *Error) Error() string {
	return fmt.Sprint(e)
}

func (e *Error) Format(s fmt.State, c rune) {
	xerrors.FormatError(e, s, c)
}

func (e *Error) FormatError(p xerrors.Printer) (next error) {
	if e.msg == "" {
		p.Printf("code=%v", e.Code)
	} else {
		p.Printf("%s (code=%v)", e.msg, e.Code)
	}
	e.frame.Format(p)
	return e.err
}

// Unwrap returns the error underlying the receiver, which may be nil.
func (e *Error) Unwrap() error {
	return e.err
}

// New returns a new error with the given code, underlying error and message. Pass 1
// for the call depth if New is called from the function raising the error; pass 2 if
// it is called from a helper function that was invoked by the original function; and
// so on.
func New(c ErrorCode, err error, callDepth int, msg string) *Error {
	return &Error{
		Code:  c,
		msg:   msg,
		frame: xerrors.Caller(callDepth),
		err:   err,
	}
}

// Code returns the ErrorCode of err if it, or some error it wraps, is an *Error.
// If err is context.Canceled or context.DeadlineExceeded, or wraps one of those errors,
// it returns the Canceled or DeadlineExceeded codes, respectively.
// If err is nil, it returns the special code OK.
// Otherwise, it returns Unknown.
func Code(err error) ErrorCode {
	if err == nil {
		return OK
	}
	var e *Error
	if xerrors.As(err, &e) {
		return e.Code
	}
	if xerrors.Is(err, context.Canceled) {
		return Canceled
	}
	if xerrors.Is(err, context.DeadlineExceeded) {
		return DeadlineExceeded
	}
	return Unknown
}

// Newf uses format and args to format a message, then calls New.
func Newf(c ErrorCode, err error, format string, args ...interface{}) *Error {
	return New(c, err, 2, fmt.Sprintf(format, args...))
}

// DoNotWrap reports whether an error should not be wrapped in the Error
// type from this package.
// It returns true if err is a retry error, a context error, io.EOF, or if it wraps
// one of those.
func DoNotWrap(err error) bool {
	if xerrors.Is(err, io.EOF) {
		return true
	}
	if xerrors.Is(err, context.Canceled) {
		return true
	}
	if xerrors.Is(err, context.DeadlineExceeded) {
		return true
	}

	return false
}

// ErrorAs is a helper for the ErrorAs method of an API's portable type.
// It performs some initial nil checks, and does a single level of unwrapping
// when err is a *Error. Then it calls its errorAs argument, which should
// be a driver implementation of ErrorAs.
func ErrorAs(err error, target interface{}, errorAs func(error, interface{}) bool) bool {
	if err == nil {
		return false
	}
	if target == nil {
		panic("ErrorAs target cannot be nil")
	}
	val := reflect.ValueOf(target)
	if val.Type().Kind() != reflect.Ptr || val.IsNil() {
		panic("ErrorAs target must be a non-nil pointer")
	}
	if e, ok := err.(*Error); ok {
		err = e.Unwrap()
	}
	return errorAs(err, target)
}
