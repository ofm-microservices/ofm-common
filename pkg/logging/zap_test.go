package logging

import (
	"errors"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"
)

func TestLogging(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Logging Suite")
}

var _ = Describe("Logger", func() {
	Describe("New", func() {
		It("creates a production logger with the requested level", func() {
			lg, err := New("ofm-common", "test", "debug")

			Expect(err).NotTo(HaveOccurred())
			Expect(lg).NotTo(BeNil())

			Expect(func() { lg.Debug("debug message", String("scope", "test")) }).NotTo(Panic())
			Expect(func() { lg.Info("info message", Int("count", 1)) }).NotTo(Panic())
			Expect(func() { lg.Warn("warn message", Any("active", true)) }).NotTo(Panic())
			Expect(func() { lg.Error("error message", Err(errors.New("boom"))) }).NotTo(Panic())
			Expect(lg.Sync()).NotTo(MatchError(ErrNilZapLogger))
		})

		It("returns an error for an invalid log level", func() {
			lg, err := New("ofm-common", "test", "definitely-not-a-level")

			Expect(lg).To(BeNil())
			Expect(err).To(MatchError(ContainSubstring(`parse log level "definitely-not-a-level"`)))
		})

		It("wraps zap build failures", func() {
			original := buildZapLogger
			buildZapLogger = func(zap.Config, ...zap.Option) (*zap.Logger, error) {
				return nil, errors.New("forced build failure")
			}
			DeferCleanup(func() {
				buildZapLogger = original
			})

			lg, err := New("ofm-common", "test", "info")

			Expect(lg).To(BeNil())
			Expect(err).To(MatchError(ContainSubstring("build logger")))
			Expect(err).To(MatchError(ContainSubstring("forced build failure")))
		})
	})
	Describe("With", func() {
		It("returns a child logger that accepts additional fields", func() {
			lg, err := New("ofm-common", "test", "info")
			Expect(err).NotTo(HaveOccurred())

			child := lg.With(String("module", "child"))

			Expect(child).NotTo(BeNil())
			Expect(func() {
				child.Info("child message", String("request_id", "req-1"))
			}).NotTo(Panic())
			Expect(child.Sync()).NotTo(MatchError(ErrNilZapLogger))
		})
	})
	Describe("Sync", func() {
		It("returns ErrNilZapLogger for a nil concrete logger", func() {
			var lg *zapLogger

			Expect(lg.Sync()).To(MatchError(ErrNilZapLogger))
		})

		It("returns ErrNilZapLogger when the embedded zap logger is nil", func() {
			lg := &zapLogger{}

			Expect(lg.Sync()).To(MatchError(ErrNilZapLogger))
		})
	})

	Describe("Field helpers", func() {
		It("builds zap fields with the expected keys", func() {
			errField := Err(errors.New("boom"))

			Expect(String("service", "common").Key).To(Equal("service"))
			Expect(Int("attempt", 2).Key).To(Equal("attempt"))
			Expect(Any("payload", map[string]any{"ok": true}).Key).To(Equal("payload"))
			Expect(errField.Key).To(Equal("error"))
		})
	})

	Describe("Error wrappers", func() {
		It("wraps parse-level errors with the provided level", func() {
			err := WrapParseLogLevelError("broken", errors.New("parse failed"))

			Expect(err).To(MatchError(ContainSubstring(`parse log level "broken"`)))
			Expect(err).To(MatchError(ContainSubstring("parse failed")))
		})

		It("wraps logger build failures", func() {
			err := WrapBuildLoggerError(errors.New("build failed"))

			Expect(err).To(MatchError(ContainSubstring("build logger")))
			Expect(err).To(MatchError(ContainSubstring("build failed")))
		})
	})
})
