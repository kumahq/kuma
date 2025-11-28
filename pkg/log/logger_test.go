package log_test

import (
	"bytes"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	kuma_log "github.com/kumahq/kuma/v2/pkg/log"
)

var _ = Describe("Logger", func() {
	Describe("NewLoggerWithGlobalLevel and SetGlobalLogLevel", func() {
		It("should create a logger that respects global level changes", func() {
			// Given a logger with global level
			logger := kuma_log.NewLoggerWithGlobalLevel()
			Expect(logger).NotTo(BeNil())

			// When we set global level to debug
			kuma_log.SetGlobalLogLevel(kuma_log.DebugLevel)

			// Then the logger should be enabled at debug level
			// V(1) corresponds to debug verbosity
			Expect(logger.V(1).Enabled()).To(BeTrue())
		})

		It("should disable logging when level is set to off", func() {
			// Given a logger with global level
			logger := kuma_log.NewLoggerWithGlobalLevel()

			// When we set global level to off
			kuma_log.SetGlobalLogLevel(kuma_log.OffLevel)

			// Then the logger should not be enabled at any level
			Expect(logger.Enabled()).To(BeFalse())
		})

		It("should allow info level logging when set to info", func() {
			// Given a logger with global level
			logger := kuma_log.NewLoggerWithGlobalLevel()

			// When we set global level to info
			kuma_log.SetGlobalLogLevel(kuma_log.InfoLevel)

			// Then info level should be enabled
			Expect(logger.Enabled()).To(BeTrue())

			// But debug level (V(1)) should not be enabled
			Expect(logger.V(1).Enabled()).To(BeFalse())
		})
	})

	Describe("NewLoggerTo", func() {
		It("should create a logger that writes to the specified writer", func() {
			// Given a buffer
			buf := &bytes.Buffer{}

			// When we create a logger to that buffer
			logger := kuma_log.NewLoggerTo(buf, kuma_log.InfoLevel)
			logger.Info("test message")

			// Then the buffer should contain the message
			Expect(buf.String()).To(ContainSubstring("test message"))
		})

		It("should respect the log level", func() {
			// Given a buffer and an info-level logger
			buf := &bytes.Buffer{}
			logger := kuma_log.NewLoggerTo(buf, kuma_log.InfoLevel)

			// When we log at debug level (V(1))
			logger.V(1).Info("debug message")

			// Then the buffer should be empty (debug not enabled at info level)
			Expect(buf.String()).To(BeEmpty())
		})

		It("should return a nop logger when level is off", func() {
			// Given an off-level logger
			buf := &bytes.Buffer{}
			logger := kuma_log.NewLoggerTo(buf, kuma_log.OffLevel)

			// When we try to log
			logger.Info("should not appear")

			// Then the buffer should be empty
			Expect(buf.String()).To(BeEmpty())
		})
	})

	Describe("LogLevel", func() {
		DescribeTable("String() should return correct string",
			func(level kuma_log.LogLevel, expected string) {
				Expect(level.String()).To(Equal(expected))
			},
			Entry("OffLevel", kuma_log.OffLevel, "off"),
			Entry("InfoLevel", kuma_log.InfoLevel, "info"),
			Entry("DebugLevel", kuma_log.DebugLevel, "debug"),
		)

		DescribeTable("ParseLogLevel should parse correctly",
			func(input string, expected kuma_log.LogLevel, shouldError bool) {
				level, err := kuma_log.ParseLogLevel(input)
				if shouldError {
					Expect(err).To(HaveOccurred())
				} else {
					Expect(err).NotTo(HaveOccurred())
					Expect(level).To(Equal(expected))
				}
			},
			Entry("off", "off", kuma_log.OffLevel, false),
			Entry("info", "info", kuma_log.InfoLevel, false),
			Entry("debug", "debug", kuma_log.DebugLevel, false),
			Entry("invalid", "invalid", kuma_log.OffLevel, true),
		)
	})
})
