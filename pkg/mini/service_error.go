package mini

import (
	"errors"
	"fmt"
)

func (e *NATSError) Error() string {
	return fmt.Sprintf("%q: %s", e.Subject, e.Description)
}

// Unwrap returns the underlying error if any.
func (e *NATSError) Unwrap() error {
	return e.err
}

// Is reports whether the target error is equal to this error.
func (e *NATSError) Is(target error) bool {
	if e == nil {
		return false
	}
	if t, ok := target.(*NATSError); ok {
		return e.Subject == t.Subject && e.Description == t.Description
	}
	return e.err != nil && errors.Is(e.err, target)
}
