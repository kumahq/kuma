package v3_test

import (
	accesslog_data "github.com/envoyproxy/go-control-plane/envoy/data/accesslog/v3"
	accesslog_config "github.com/envoyproxy/go-control-plane/envoy/extensions/access_loggers/grpc/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/pkg/envoy/accesslog/v3"
)

var _ = Describe("ResponseTrailerOperator", func() {

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
				fragment := &ResponseTrailerOperator{HeaderFormatter{
					Header: given.header, AltHeader: given.altHeader, MaxLength: given.maxLength}}
				// when
				actual, err := fragment.FormatHttpLogEntry(given.entry)
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
			fragment := &ResponseTrailerOperator{HeaderFormatter{
				Header: "grpc-status", AltHeader: "grpc-message", MaxLength: 123}}
			// when
			actual, err := fragment.FormatTcpLogEntry(&accesslog_data.TCPAccessLogEntry{})
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
				fragment := &ResponseTrailerOperator{HeaderFormatter{
					Header: given.header, AltHeader: given.altHeader}}
				// when
				err := fragment.ConfigureHttpLog(given.config)
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

	Describe("String()", func() {
		type testCase struct {
			header    string
			altHeader string
			maxLength int
			expected  string
		}

		DescribeTable("should return correct canonical representation",
			func(given testCase) {
				// setup
				fragment := &ResponseTrailerOperator{HeaderFormatter{
					Header: given.header, AltHeader: given.altHeader, MaxLength: given.maxLength}}

				// when
				actual := fragment.String()
				// then
				Expect(actual).To(Equal(given.expected))

			},
			Entry("%TRAILER()%", testCase{
				expected: `%TRAILER()%`,
			}),
			Entry("%TRAILER():10%", testCase{
				maxLength: 10,
				expected:  `%TRAILER():10%`,
			}),
			Entry("%TRAILER(grpc-status)%", testCase{
				header:   `grpc-status`,
				expected: `%TRAILER(grpc-status)%`,
			}),
			Entry("%TRAILER(grpc-status):10%", testCase{
				header:    `grpc-status`,
				maxLength: 10,
				expected:  `%TRAILER(grpc-status):10%`,
			}),
			Entry("%TRAILER(?grpc-message)%", testCase{
				altHeader: `grpc-message`,
				expected:  `%TRAILER(?grpc-message)%`,
			}),
			Entry("%TRAILER(?grpc-message):10%", testCase{
				altHeader: `grpc-message`,
				maxLength: 10,
				expected:  `%TRAILER(?grpc-message):10%`,
			}),
			Entry("%TRAILER(grpc-status?grpc-message)%", testCase{
				header:    "grpc-status",
				altHeader: `grpc-message`,
				expected:  `%TRAILER(grpc-status?grpc-message)%`,
			}),
			Entry("%TRAILER(grpc-status?grpc-message):10%", testCase{
				header:    "grpc-status",
				altHeader: `grpc-message`,
				maxLength: 10,
				expected:  `%TRAILER(grpc-status?grpc-message):10%`,
			}),
		)
	})
})
