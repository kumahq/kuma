package log_test

import (
	"bytes"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	kuma_log "github.com/kumahq/kuma/v2/pkg/log"
)

var _ = Describe("ComponentAwareSink", func() {
	var (
		registry *kuma_log.ComponentLevelRegistry
		buf      *bytes.Buffer
	)

	BeforeEach(func() {
		registry = kuma_log.NewComponentLevelRegistry()
		buf = &bytes.Buffer{}
	})

	newLogger := func() logr.Logger {
		return kuma_log.NewLoggerToWithRegistry(buf, kuma_log.InfoLevel, registry)
	}

	It("should accumulate dot-separated names", func() {
		logger := newLogger()
		named := logger.WithName("xds").WithName("server")

		// Without override, info level should work
		named.Info("test")
		Expect(buf.String()).To(ContainSubstring("test"))
	})

	It("should allow debug logs when component override is debug", func() {
		Expect(registry.SetLevel("xds", kuma_log.DebugLevel)).To(Succeed())
		logger := newLogger().WithName("xds")

		logger.V(1).Info("debug message")
		Expect(buf.String()).To(ContainSubstring("debug message"))
	})

	It("should block debug logs when component override is info", func() {
		Expect(registry.SetLevel("xds", kuma_log.InfoLevel)).To(Succeed())
		logger := newLogger().WithName("xds")

		logger.V(1).Info("debug message")
		Expect(buf.String()).To(BeEmpty())
	})

	It("should block all logs when component override is off", func() {
		Expect(registry.SetLevel("xds", kuma_log.OffLevel)).To(Succeed())
		logger := newLogger().WithName("xds")

		logger.Info("info message")
		Expect(buf.String()).To(BeEmpty())
	})

	It("should apply parent override to child component", func() {
		Expect(registry.SetLevel("xds", kuma_log.DebugLevel)).To(Succeed())
		logger := newLogger().WithName("xds").WithName("server")

		logger.V(1).Info("debug from child")
		Expect(buf.String()).To(ContainSubstring("debug from child"))
	})

	It("should fall back to global level when no override", func() {
		// No override set — global level is info
		logger := newLogger().WithName("xds")

		logger.V(1).Info("debug message")
		Expect(buf.String()).To(BeEmpty())

		logger.Info("info message")
		Expect(buf.String()).To(ContainSubstring("info message"))
	})

	It("should not affect unnamed logger with overrides", func() {
		Expect(registry.SetLevel("xds", kuma_log.OffLevel)).To(Succeed())
		logger := newLogger()

		logger.Info("root message")
		Expect(buf.String()).To(ContainSubstring("root message"))
	})

	It("should still emit error logs when component override is off", func() {
		Expect(registry.SetLevel("xds", kuma_log.OffLevel)).To(Succeed())
		logger := newLogger().WithName("xds")

		logger.Error(nil, "error message")
		Expect(buf.String()).To(ContainSubstring("error message"))
	})

	It("should allow error logs when component override is info", func() {
		Expect(registry.SetLevel("xds", kuma_log.InfoLevel)).To(Succeed())
		logger := newLogger().WithName("xds")

		logger.Error(nil, "error message")
		Expect(buf.String()).To(ContainSubstring("error message"))
	})
})
