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

	type logAction struct {
		vLevel  int
		message string
		isError bool
	}

	DescribeTable("per-component overrides",
		func(overrideComponent string, overrideLevel kuma_log.LogLevel, names []string, action logAction, expectOutput bool) {
			if overrideComponent != "" {
				Expect(registry.SetLevel(overrideComponent, overrideLevel)).To(Succeed())
			}
			logger := newLogger()
			for _, n := range names {
				logger = logger.WithName(n)
			}
			if action.isError {
				logger.Error(nil, action.message)
			} else if action.vLevel > 0 {
				logger.V(action.vLevel).Info(action.message)
			} else {
				logger.Info(action.message)
			}
			if expectOutput {
				Expect(buf.String()).To(ContainSubstring(action.message))
			} else {
				Expect(buf.String()).To(BeEmpty())
			}
		},
		Entry("debug override allows V(1) logs",
			"xds", kuma_log.DebugLevel, []string{"xds"},
			logAction{vLevel: 1, message: "debug message"}, true),
		Entry("info override blocks V(1) logs",
			"xds", kuma_log.InfoLevel, []string{"xds"},
			logAction{vLevel: 1, message: "debug message"}, false),
		Entry("off override blocks info logs",
			"xds", kuma_log.OffLevel, []string{"xds"},
			logAction{message: "info message"}, false),
		Entry("parent override applies to child",
			"xds", kuma_log.DebugLevel, []string{"xds", "server"},
			logAction{vLevel: 1, message: "debug from child"}, true),
		Entry("off override does not suppress errors",
			"xds", kuma_log.OffLevel, []string{"xds"},
			logAction{isError: true, message: "error message"}, true),
		Entry("info override allows errors",
			"xds", kuma_log.InfoLevel, []string{"xds"},
			logAction{isError: true, message: "error message"}, true),
		Entry("override on other component does not affect unnamed logger",
			"xds", kuma_log.OffLevel, nil,
			logAction{message: "root message"}, true),
	)

	It("should accumulate dot-separated names", func() {
		named := newLogger().WithName("xds").WithName("server")
		named.Info("test")
		Expect(buf.String()).To(ContainSubstring("test"))
	})

	It("should fall back to base level when no override", func() {
		logger := newLogger().WithName("xds")

		logger.V(1).Info("debug message")
		Expect(buf.String()).To(BeEmpty())

		logger.Info("info message")
		Expect(buf.String()).To(ContainSubstring("info message"))
	})

	It("should return to base level after override is reset", func() {
		Expect(registry.SetLevel("xds", kuma_log.OffLevel)).To(Succeed())
		logger := newLogger().WithName("xds")

		logger.Info("suppressed")
		Expect(buf.String()).To(BeEmpty())

		registry.ResetLevel("xds")
		logger.Info("visible after reset")
		Expect(buf.String()).To(ContainSubstring("visible after reset"))
	})

	It("child-level override overrides parent", func() {
		Expect(registry.SetLevel("xds", kuma_log.OffLevel)).To(Succeed())
		Expect(registry.SetLevel("xds.server", kuma_log.DebugLevel)).To(Succeed())

		child := newLogger().WithName("xds").WithName("server")
		child.V(1).Info("child debug visible")
		Expect(buf.String()).To(ContainSubstring("child debug visible"))
	})

	It("deep 3-level hierarchy inherits from top ancestor", func() {
		Expect(registry.SetLevel("plugins", kuma_log.DebugLevel)).To(Succeed())

		logger := newLogger().WithName("plugins").WithName("authn").WithName("tokens")
		logger.V(1).Info("deep debug")
		Expect(buf.String()).To(ContainSubstring("deep debug"))
	})
})
