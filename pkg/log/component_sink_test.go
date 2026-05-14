package log_test

import (
	"bytes"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"

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

var _ = Describe("log volume counter", Serial, func() {
	var (
		cv  *prometheus.CounterVec
		buf *bytes.Buffer
		clr *kuma_log.ComponentLevelRegistry
	)

	BeforeEach(func() {
		reg := prometheus.NewRegistry()
		var err error
		cv, err = kuma_log.SetupMetrics(reg)
		Expect(err).NotTo(HaveOccurred())
		buf = &bytes.Buffer{}
		clr = kuma_log.NewComponentLevelRegistry()
	})

	newLogger := func() logr.Logger {
		return kuma_log.NewLoggerToWithRegistry(buf, kuma_log.InfoLevel, clr)
	}

	count := func(logger, level string) float64 {
		return testutil.ToFloat64(cv.WithLabelValues(logger, level))
	}

	It("increments info counter per logger", func() {
		newLogger().WithName("xds").Info("hello")
		Expect(count("xds", "info")).To(Equal(1.0))
	})

	It("increments error counter per logger", func() {
		newLogger().WithName("kds").Error(nil, "oops")
		Expect(count("kds", "error")).To(Equal(1.0))
	})

	It("increments debug counter for V(1) calls", func() {
		Expect(clr.SetLevel("xds", kuma_log.DebugLevel)).To(Succeed())
		newLogger().WithName("xds").V(1).Info("verbose")
		Expect(count("xds", "debug")).To(Equal(1.0))
	})

	It("counts multiple calls", func() {
		l := newLogger().WithName("api")
		l.Info("one")
		l.Info("two")
		l.Info("three")
		Expect(count("api", "info")).To(Equal(3.0))
	})

	It("tracks separate loggers independently", func() {
		newLogger().WithName("xds").Info("from xds")
		newLogger().WithName("kds").Info("from kds")
		Expect(count("xds", "info")).To(Equal(1.0))
		Expect(count("kds", "info")).To(Equal(1.0))
	})
})
