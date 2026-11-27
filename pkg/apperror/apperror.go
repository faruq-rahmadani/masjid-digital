package apperror

// Code classifies an error independently of the transport.
// The HTTP delivery maps each code to a status code.
type Code string

const (
	CodeInvalidInput    Code = "INVALID_INPUT"   // request data is malformed or missing
	CodeUnauthenticated Code = "UNAUTHENTICATED" // no valid login
	CodeForbidden       Code = "FORBIDDEN"       // logged in, but not allowed
	CodeNotFound        Code = "NOT_FOUND"       // resource does not exist
	CodeAlreadyExists   Code = "ALREADY_EXISTS"  // resource is a duplicate, e.g. phone taken
	CodeBusinessRule    Code = "BUSINESS_RULE"   // valid data that breaks a rule, e.g. spending more than the balance
	CodeInternal        Code = "INTERNAL"        // unexpected failure
)

type AppError struct {
	Code          Code   // internal error code
	InternalError error  // error from the app, for logs only
	MessageError  string // message consumed by the frontend
}

func New(code Code, internalError error, messageError string) *AppError {
	return &AppError{
		Code:          code,
		InternalError: internalError,
		MessageError:  messageError,
	}
}

// Error returns the internal error for logs, falling back to the frontend message
func (e *AppError) Error() string {
	if e.InternalError == nil {
		return e.MessageError
	}
	return e.InternalError.Error()
}

// Unwrap lets errors.Is/As reach the internal error, e.g. a domain error
func (e *AppError) Unwrap() error {
	return e.InternalError
}
