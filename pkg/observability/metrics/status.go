package metrics

import "fmt"

// StatusFromError returns success when err is nil and error otherwise.
func StatusFromError(err error) string {
	if err != nil {
		return "error"
	}
	return "success"
}

// HTTPStatusClass converts an HTTP status code into a low-cardinality label.
func HTTPStatusClass(code int) string {
	if code <= 0 {
		return "unknown"
	}
	return fmt.Sprintf("%dxx", code/100)
}
