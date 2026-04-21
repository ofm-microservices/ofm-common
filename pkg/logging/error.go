package logging

import (
	"errors"
	"fmt"
)

// ErrNilZapLogger is returned when logger operations are attempted on a nil
// zap logger.
var ErrNilZapLogger = errors.New("zap logger is nil")

// WrapParseLogLevelError annotates invalid log-level input.
func WrapParseLogLevelError(level string, err error) error {
	return fmt.Errorf("parse log level %q: %w", level, err)
}

// WrapBuildLoggerError annotates zap logger construction failures.
func WrapBuildLoggerError(err error) error {
	return fmt.Errorf("build logger: %w", err)
}
