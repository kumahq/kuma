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
			Entry("no header or altHeader", testCase{
				header:    "",
				altHeader: "",
				entry:     example,
				expected:  ``, // apparently, Envoy allows both `Header` and `AltHeader` to be empty
			}),
			Entry("grpc-message", testCase{
				header:   "grpc-message",
				entry:    example,
				expected: `unavailable`,
			}),
			Entry("x-trailer", testCase{
				header:   "x-trailer", // missing header
				entry:    example,
				expected: ``,
			}),
			Entry("x-trailer or grpc-message", testCase{
				header:    "x-trailer", // missing header
				altHeader: "grpc-message",
				entry:     example,
				expected:  `unavailable`,
			}),
			Entry("x-trailer or grpc-message w/ MaxLength", testCase{
				header:    "x-trailer", // missing header
				altHeader: "grpc-message",
				maxLength: 2,
				entry:     example,
				expected:  `un`,
			}),
			Entry("grpc-status or grpc-message", testCase{
				header:    "grpc-status",
				altHeader: "grpc-message",
				entry:     example,
				expected:  `14`,
			}),
			Entry("grpc-status or grpc-message", testCase{
				header:    "grpc-status",
				altHeader: "grpc-message",
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
				Header: "grpc-status", AltHeader: "grpc-message", MaxLength: 123}}
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
				header:    "grpc-status",
				altHeader: "",
				config:    &accesslog_config.HttpGrpcAccessLogConfig{},
				expected: &accesslog_config.HttpGrpcAccessLogConfig{
					AdditionalResponseTrailersToLog: []string{"grpc-status"},
				},
			}),
			Entry("altHeader", testCase{
				header:    "",
				altHeader: "grpc-message",
				config:    &accesslog_config.HttpGrpcAccessLogConfig{},
				expected: &accesslog_config.HttpGrpcAccessLogConfig{
					AdditionalResponseTrailersToLog: []string{"grpc-message"},
				},
			}),
			Entry("header and altHeader", testCase{
				header:    "grpc-status",
				altHeader: "grpc-message",
				config:    &accesslog_config.HttpGrpcAccessLogConfig{},
				expected: &accesslog_config.HttpGrpcAccessLogConfig{
					AdditionalResponseTrailersToLog: []string{"grpc-status", "grpc-message"},
				},
			}),
			Entry("none w/ initial config", testCase{
				header:    "",
				altHeader: "",
				config: &accesslog_config.HttpGrpcAccessLogConfig{
					AdditionalResponseTrailersToLog: []string{"x-trailer", "x-bye"},
				},
				expected: &accesslog_config.HttpGrpcAccessLogConfig{
					AdditionalResponseTrailersToLog: []string{"x-trailer", "x-bye"},
				},
			}),
			Entry("header w/ initial config", testCase{
				header:    "grpc-status",
				altHeader: "",
				config: &accesslog_config.HttpGrpcAccessLogConfig{
					AdditionalResponseTrailersToLog: []string{"x-trailer", "x-bye"},
				},
				expected: &accesslog_config.HttpGrpcAccessLogConfig{
					AdditionalResponseTrailersToLog: []string{"x-trailer", "x-bye", "grpc-status"},
				},
			}),
			Entry("altHeader w/ initial config", testCase{
				header:    "",
				altHeader: "grpc-message",
				config: &accesslog_config.HttpGrpcAccessLogConfig{
					AdditionalResponseTrailersToLog: []string{"x-trailer", "x-bye"},
				},
				expected: &accesslog_config.HttpGrpcAccessLogConfig{
					AdditionalResponseTrailersToLog: []string{"x-trailer", "x-bye", "grpc-message"},
				},
			}),
			Entry("header and altHeader w/ initial config", testCase{
				header:    "grpc-status",
				altHeader: "grpc-message",
				config: &accesslog_config.HttpGrpcAccessLogConfig{
					AdditionalResponseTrailersToLog: []string{"x-trailer", "x-bye"},
				},
				expected: &accesslog_config.HttpGrpcAccessLogConfig{
					AdditionalResponseTrailersToLog: []string{"x-trailer", "x-bye", "grpc-status", "grpc-message"},
				},
			}),
		)
	})
})
