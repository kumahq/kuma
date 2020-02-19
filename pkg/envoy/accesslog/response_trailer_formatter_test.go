package accesslog_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	. "github.com/Kong/kuma/pkg/envoy/accesslog"

	accesslog_config "github.com/envoyproxy/go-control-plane/envoy/config/accesslog/v2"
	accesslog_data "github.com/envoyproxy/go-control-plane/envoy/data/accesslog/v2"
)

var _ = Describe("ResponseTrailerFormatter", func() {

	Describe("FormatHttpLogEntry()", func() {
		example := &accesslog_data.HTTPAccessLogEntry{
			Response: &accesslog_data.HTTPResponseProperties{
				ResponseTrailers: map[string]string{
					"grpc-status":  "14",
					"grpc-message": "unavailable",
				},
			},
		}

		type testCase struct {
			header    string
			altHeader string
			maxLength int
			entry     *accesslog_data.HTTPAccessLogEntry
			expected  string
		}

		DescribeTable("should format properly",
			func(given testCase) {
				// setup
				formatter := &ResponseTrailerFormatter{HeaderFormatter{
					Header: given.header, AltHeader: given.altHeader, MaxLength: given.maxLength}}
				// when
				actual, err := formatter.FormatHttpLogEntry(given.entry)
				// then
				Expect(err).ToNot(HaveOccurred())
				// and
				Expect(actual).To(Equal(given.expected))
			},
			Entry("GRPC-MESSAGE", testCase{
				header:   "GRPC-MESSAGE",
				entry:    example,
				expected: `unavailable`,
			}),
			Entry("X-TRAILER", testCase{
				header:   "X-TRAILER", // missing header
				entry:    example,
				expected: ``,
			}),
			Entry("X-TRAILER or GRPC-MESSAGE", testCase{
				header:    "X-TRAILER", // missing header
				altHeader: "GRPC-MESSAGE",
				entry:     example,
				expected:  `unavailable`,
			}),
			Entry("X-TRAILER or GRPC-MESSAGE w/ MaxLength", testCase{
				header:    "X-TRAILER", // missing header
				altHeader: "GRPC-MESSAGE",
				maxLength: 2,
				entry:     example,
				expected:  `un`,
			}),
			Entry("GRPC-STATUS or GRPC-MESSAGE", testCase{
				header:    "GRPC-STATUS",
				altHeader: "GRPC-MESSAGE",
				entry:     example,
				expected:  `14`,
			}),
			Entry("GRPC-STATUS or GRPC-MESSAGE", testCase{
				header:    "GRPC-STATUS",
				altHeader: "GRPC-MESSAGE",
				maxLength: 1,
				entry:     example,
				expected:  `1`,
			}),
		)
	})

	Describe("FormatTcpLogEntry()", func() {
		It("should always return an empty string", func() {
			// setup
			formatter := &ResponseTrailerFormatter{HeaderFormatter{
				Header: "GRPC-STATUS", AltHeader: "GRPC-MESSAGE", MaxLength: 123}}
			// when
			actual, err := formatter.FormatTcpLogEntry(&accesslog_data.TCPAccessLogEntry{})
			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(actual).To(Equal(``))
		})
	})

	Describe("ConfigureHttpLog()", func() {

		type testCase struct {
			header    string
			altHeader string
			config    *accesslog_config.HttpGrpcAccessLogConfig
			expected  *accesslog_config.HttpGrpcAccessLogConfig // verify the entire config to make sure there are no unexpected changes
		}

		DescribeTable("should configure properly",
			func(given testCase) {
				// setup
				formatter := &ResponseTrailerFormatter{HeaderFormatter{
					Header: given.header, AltHeader: given.altHeader}}
				// when
				err := formatter.ConfigureHttpLog(given.config)
				// then
				Expect(err).ToNot(HaveOccurred())
				// and
				Expect(given.config).To(Equal(given.expected))
			},
			Entry("none", testCase{
				header:    "",
				altHeader: "",
				config:    &accesslog_config.HttpGrpcAccessLogConfig{},
				expected:  &accesslog_config.HttpGrpcAccessLogConfig{},
			}),
			Entry("header", testCase{
				header:    "GRPC-STATUS",
				altHeader: "",
				config:    &accesslog_config.HttpGrpcAccessLogConfig{},
				expected: &accesslog_config.HttpGrpcAccessLogConfig{
					AdditionalResponseTrailersToLog: []string{"GRPC-STATUS"},
				},
			}),
			Entry("altHeader", testCase{
				header:    "",
				altHeader: "GRPC-MESSAGE",
				config:    &accesslog_config.HttpGrpcAccessLogConfig{},
				expected: &accesslog_config.HttpGrpcAccessLogConfig{
					AdditionalResponseTrailersToLog: []string{"GRPC-MESSAGE"},
				},
			}),
			Entry("header and altHeader", testCase{
				header:    "GRPC-STATUS",
				altHeader: "GRPC-MESSAGE",
				config:    &accesslog_config.HttpGrpcAccessLogConfig{},
				expected: &accesslog_config.HttpGrpcAccessLogConfig{
					AdditionalResponseTrailersToLog: []string{"GRPC-STATUS", "GRPC-MESSAGE"},
				},
			}),
			Entry("none w/ initial config", testCase{
				header:    "",
				altHeader: "",
				config: &accesslog_config.HttpGrpcAccessLogConfig{
					AdditionalResponseTrailersToLog: []string{"X-TRAILER", "X-BYE"},
				},
				expected: &accesslog_config.HttpGrpcAccessLogConfig{
					AdditionalResponseTrailersToLog: []string{"X-TRAILER", "X-BYE"},
				},
			}),
			Entry("header w/ initial config", testCase{
				header:    "GRPC-STATUS",
				altHeader: "",
				config: &accesslog_config.HttpGrpcAccessLogConfig{
					AdditionalResponseTrailersToLog: []string{"X-TRAILER", "X-BYE"},
				},
				expected: &accesslog_config.HttpGrpcAccessLogConfig{
					AdditionalResponseTrailersToLog: []string{"X-TRAILER", "X-BYE", "GRPC-STATUS"},
				},
			}),
			Entry("altHeader w/ initial config", testCase{
				header:    "",
				altHeader: "GRPC-MESSAGE",
				config: &accesslog_config.HttpGrpcAccessLogConfig{
					AdditionalResponseTrailersToLog: []string{"X-TRAILER", "X-BYE"},
				},
				expected: &accesslog_config.HttpGrpcAccessLogConfig{
					AdditionalResponseTrailersToLog: []string{"X-TRAILER", "X-BYE", "GRPC-MESSAGE"},
				},
			}),
			Entry("header and altHeader w/ initial config", testCase{
				header:    "GRPC-STATUS",
				altHeader: "GRPC-MESSAGE",
				config: &accesslog_config.HttpGrpcAccessLogConfig{
					AdditionalResponseTrailersToLog: []string{"X-TRAILER", "X-BYE"},
				},
				expected: &accesslog_config.HttpGrpcAccessLogConfig{
					AdditionalResponseTrailersToLog: []string{"X-TRAILER", "X-BYE", "GRPC-STATUS", "GRPC-MESSAGE"},
				},
			}),
		)
	})
})
