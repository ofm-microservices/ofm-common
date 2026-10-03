package logging

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/ofm-microservices/ofm-common/pkg/observability/metadata"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

var defaultState struct {
	sync.RWMutex
	logger Logger
}

// IsDevelopment reports whether verbose transport logging is enabled for the
// supplied application environment. Only an explicit dev/development value
// enables request and message payload diagnostics.
func IsDevelopment(env string) bool {
	switch strings.ToLower(strings.TrimSpace(env)) {
	case "dev", "development":
		return true
	default:
		return false
	}
}

// Field is the structured logging field type accepted by Logger methods.
type Field = zap.Field

// Logger is the small structured logging contract shared across OFM services.
type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	With(fields ...Field) Logger
	Sync() error
}

type verboseLogger interface{ verbose() bool }

// IsVerbose reports whether a logger was constructed for development
// transport diagnostics. It keeps the mode decision inside logging rather than
// forcing adapters to receive application configuration.
func IsVerbose(lg Logger) bool {
	if l, ok := lg.(verboseLogger); ok {
		return l.verbose()
	}
	return false
}

// SetDefault registers the process logger for generic transport adapters that
// intentionally do not depend on service configuration or DI details.
func SetDefault(lg Logger) {
	defaultState.Lock()
	defaultState.logger = lg
	defaultState.Unlock()
}

// ProcessLogger returns the process logger registered during bootstrap.
func ProcessLogger() Logger {
	defaultState.RLock()
	defer defaultState.RUnlock()
	return defaultState.logger
}

// String creates a string-valued logging field.
func String(key, value string) Field { return zap.String(key, value) }

// Int creates an int-valued logging field.
func Int(key string, value int) Field { return zap.Int(key, value) }

// Int64 creates an int64-valued logging field.
func Int64(key string, value int64) Field { return zap.Int64(key, value) }

// Any creates a generic logging field.
func Any(key string, value any) Field { return zap.Any(key, value) }

// Err creates an error logging field.
func Err(err error) Field { return zap.Error(err) }

// Operation creates a field that names the action being performed.
func Operation(value string) Field { return String("operation", value) }

// Attempt creates a field that records the current retry attempt number.
func Attempt(value int) Field { return Int("attempt", value) }

// Retryable creates a field that indicates whether an error may be retried.
func Retryable(value bool) Field { return zap.Bool("retryable", value) }

// DurationMS creates a field that records elapsed time in milliseconds.
func DurationMS(value time.Duration) Field { return zap.Int64("duration_ms", value.Milliseconds()) }

// WithContext returns a child logger enriched with trace correlation fields
// extracted from the supplied context.
func WithContext(ctx context.Context, lg Logger) Logger {
	if lg == nil {
		return nil
	}
	if ctx == nil {
		return lg
	}
	values := metadata.FromContext(ctx)
	fields := make([]Field, 0, 4)
	if values.RequestID != "" {
		fields = append(fields, String("request_id", values.RequestID))
	}
	if values.CorrelationID != "" {
		fields = append(fields, String("correlation_id", values.CorrelationID))
	}
	if values.TestRunID != "" {
		fields = append(fields, String("test_run_id", values.TestRunID))
	}
	if values.TestScenarioID != "" {
		fields = append(fields, String("scenario_id", values.TestScenarioID))
	}
	sc := trace.SpanContextFromContext(ctx)
	if sc.IsValid() {
		fields = append(fields, String("trace_id", sc.TraceID().String()), String("span_id", sc.SpanID().String()))
	}
	if len(fields) == 0 {
		return lg
	}
	return lg.With(fields...)
}
