package logging

import "go.uber.org/zap"

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

// String creates a string-valued logging field.
func String(key, value string) Field { return zap.String(key, value) }

// Int creates an int-valued logging field.
func Int(key string, value int) Field { return zap.Int(key, value) }

// Any creates a generic logging field.
func Any(key string, value any) Field { return zap.Any(key, value) }

// Err creates an error logging field.
func Err(err error) Field { return zap.Error(err) }
