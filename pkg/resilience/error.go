// Package resilience contains shared reliability primitives for OFM services.
package resilience

import "errors"

// ErrCircuitOpen indicates that a dependency is currently failing fast.
var ErrCircuitOpen = errors.New("resilience: circuit breaker is open")

// PermanentError marks an error that must not be retried.
type PermanentError struct{ Err error }

// Error implements error.
func (e PermanentError) Error() string { return e.Err.Error() }

// Unwrap exposes the underlying application error.
func (e PermanentError) Unwrap() error { return e.Err }

// Permanent wraps an error so retry runners route it directly to their DLQ
// or caller instead of retrying it.
func Permanent(err error) error {
	if err == nil {
		return nil
	}
	return PermanentError{Err: err}
}
