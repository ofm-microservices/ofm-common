package logging

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type zapLogger struct {
	logger *zap.Logger
}

// New constructs the production JSON logger used by OFM services.
func New(service, env, level string) (Logger, error) {
	cfg := zap.NewProductionConfig()
	cfg.Encoding = "json"
	cfg.EncoderConfig.TimeKey = "ts"
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.OutputPaths = []string{"stdout"}
	cfg.ErrorOutputPaths = []string{"stderr"}

	if level != "" {
		zapLevel, err := zapcore.ParseLevel(level)
		if err != nil {
			return nil, WrapParseLogLevelError(level, err)
		}
		cfg.Level = zap.NewAtomicLevelAt(zapLevel)
	}

	zl, err := cfg.Build(zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	if err != nil {
		return nil, WrapBuildLoggerError(err)
	}

	return &zapLogger{logger: zl.With(
		zap.String("service", service),
		zap.String("env", env),
		zap.Int("pid", os.Getpid()),
	)}, nil
}

func (l *zapLogger) Debug(msg string, fields ...Field) { l.logger.Debug(msg, fields...) }
func (l *zapLogger) Info(msg string, fields ...Field)  { l.logger.Info(msg, fields...) }
func (l *zapLogger) Warn(msg string, fields ...Field)  { l.logger.Warn(msg, fields...) }
func (l *zapLogger) Error(msg string, fields ...Field) { l.logger.Error(msg, fields...) }

func (l *zapLogger) With(fields ...Field) Logger {
	return &zapLogger{logger: l.logger.With(fields...)}
}

func (l *zapLogger) Sync() error {
	if l == nil || l.logger == nil {
		return ErrNilZapLogger
	}
	return l.logger.Sync()
}
