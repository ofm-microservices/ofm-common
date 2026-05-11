package tracing

import "errors"

// ErrNilTracerProvider is returned when tracing setup is requested without an
// initialized tracer provider.
var ErrNilTracerProvider = errors.New("tracer provider is nil")
