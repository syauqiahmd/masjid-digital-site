// Package apperror classifies errors independently of the transport; the HTTP delivery maps
// each code to a status code and shows MessageError to the visitor.
package apperror

// Code classifies an error
type Code string

const (
	CodeInvalidInput    Code = "INVALID_INPUT"     // request data is malformed or missing
	CodeNotFound        Code = "NOT_FOUND"         // no such route or resource
	CodeTooManyRequests Code = "TOO_MANY_REQUESTS" // rate limited, try again later
	CodeInternal        Code = "INTERNAL"          // unexpected failure
	CodeUnavailable     Code = "UNAVAILABLE"       // a feature that is not configured, or a service it needs is down
)

type AppError struct {
	Code          Code   // internal error code
	InternalError error  // error from the app, for logs only
	MessageError  string // message shown to the visitor, in Indonesian
}

func New(code Code, internalError error, messageError string) *AppError {
	return &AppError{Code: code, InternalError: internalError, MessageError: messageError}
}

// Error returns the internal error for logs, falling back to the visitor's message
func (e *AppError) Error() string {
	if e.InternalError == nil {
		return e.MessageError
	}
	return e.InternalError.Error()
}

// Unwrap lets errors.Is/As reach the internal error
func (e *AppError) Unwrap() error {
	return e.InternalError
}
