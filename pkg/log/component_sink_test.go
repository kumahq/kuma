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

	// V(1) = debug logs: gated by component level override
	DescribeTable("debug-level gating via V(1)",
		func(overrideLevel kuma_log.LogLevel, names []string, expectOutput bool) {
			Expect(registry.SetLevel("xds", overrideLevel)).To(Succeed())
			logger := newLogger()
			for _, n := range names {
				logger = logger.WithName(n)
			}
			logger.V(1).Info("debug message")
			if expectOutput {
				Expect(buf.String()).To(ContainSubstring("debug message"))
			} else {
				Expect(buf.String()).To(BeEmpty())
			}
		},
		Entry("debug override allows V(1) logs", kuma_log.DebugLevel, []string{"xds"}, true),
		Entry("info override blocks V(1) logs", kuma_log.InfoLevel, []string{"xds"}, false),
		Entry("parent debug override applies to child", kuma_log.DebugLevel, []string{"xds", "server"}, true),
	)

	// Info logs: gated by component level override
	DescribeTable("info-level gating",
		func(overrideLevel kuma_log.LogLevel, expectOutput bool) {
			Expect(registry.SetLevel("xds", overrideLevel)).To(Succeed())
			newLogger().WithName("xds").Info("info message")
			if expectOutput {
				Expect(buf.String()).To(ContainSubstring("info message"))
			} else {
				Expect(buf.String()).To(BeEmpty())
			}
		},
		Entry("off override blocks info logs", kuma_log.OffLevel, false),
		Entry("info override allows info logs", kuma_log.InfoLevel, true),
		Entry("debug override allows info logs", kuma_log.DebugLevel, true),
	)

	// Error logs: always pass through regardless of component level
	DescribeTable("errors always pass through",
		func(overrideLevel kuma_log.LogLevel) {
			Expect(registry.SetLevel("xds", overrideLevel)).To(Succeed())
			newLogger().WithName("xds").Error(nil, "error message")
			Expect(buf.String()).To(ContainSubstring("error message"))
		},
		Entry("off level does not suppress errors", kuma_log.OffLevel),
		Entry("info level allows errors", kuma_log.InfoLevel),
	)

	It("override on unrelated component does not affect unnamed logger", func() {
		Expect(registry.SetLevel("xds", kuma_log.OffLevel)).To(Succeed())

		newLogger().Info("root message")

		Expect(buf.String()).To(ContainSubstring("root message"))
	})

	It("falls back to base level when no override set", func() {
		logger := newLogger().WithName("xds")

		logger.V(1).Info("debug message")
		Expect(buf.String()).To(BeEmpty())

		logger.Info("info message")
		Expect(buf.String()).To(ContainSubstring("info message"))
	})

	It("returns to base level after override is reset", func() {
		Expect(registry.SetLevel("xds", kuma_log.OffLevel)).To(Succeed())
		logger := newLogger().WithName("xds")

		logger.Info("suppressed")
		Expect(buf.String()).To(BeEmpty())

		registry.ResetLevel("xds")
		logger.Info("visible after reset")
		Expect(buf.String()).To(ContainSubstring("visible after reset"))
	})

	It("child-level override takes precedence over parent", func() {
		Expect(registry.SetLevel("xds", kuma_log.OffLevel)).To(Succeed())
		Expect(registry.SetLevel("xds.server", kuma_log.DebugLevel)).To(Succeed())

		newLogger().WithName("xds").WithName("server").V(1).Info("child debug visible")

		Expect(buf.String()).To(ContainSubstring("child debug visible"))
	})

	It("deep 3-level hierarchy inherits from top ancestor", func() {
		Expect(registry.SetLevel("plugins", kuma_log.DebugLevel)).To(Succeed())

		newLogger().WithName("plugins").WithName("authn").WithName("tokens").V(1).Info("deep debug")

		Expect(buf.String()).To(ContainSubstring("deep debug"))
	})
})
