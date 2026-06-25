// Package apperror provides client-safe errors that handlers can surface
// directly to API callers, distinct from unexpected internal errors.
package apperror

// ClientError is an error safe to show to API clients.
type ClientError struct {
	Code    string
	Message string
	Status  int
}

func (e *ClientError) Error() string {
	return e.Message
}

// New creates a client-safe error.
func New(status int, code, message string) *ClientError {
	return &ClientError{Status: status, Code: code, Message: message}
}

// IsClient reports whether err is a *ClientError.
func IsClient(err error) (*ClientError, bool) {
	if ce, ok := err.(*ClientError); ok {
		return ce, true
	}
	return nil, false
}
