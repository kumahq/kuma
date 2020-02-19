package accesslog_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	. "github.com/Kong/kuma/pkg/envoy/accesslog"

	accesslog_config "github.com/envoyproxy/go-control-plane/envoy/config/accesslog/v2"
	accesslog_data "github.com/envoyproxy/go-control-plane/envoy/data/accesslog/v2"
)

var _ = Describe("ResponseHeaderFormatter", func() {

	Describe("FormatHttpLogEntry()", func() {
		example := &accesslog_data.HTTPAccessLogEntry{
			Response: &accesslog_data.HTTPResponseProperties{
				ResponseHeaders: map[string]string{
					"server":       "Tomcat",
					"content-type": "application/json",
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
				formatter := &ResponseHeaderFormatter{HeaderFormatter{
					Header: given.header, AltHeader: given.altHeader, MaxLength: given.maxLength}}
				// when
				actual, err := formatter.FormatHttpLogEntry(given.entry)
				// then
				Expect(err).ToNot(HaveOccurred())
				// and
				Expect(actual).To(Equal(given.expected))
			},
			Entry("CONTENT-TYPE", testCase{
				header:   "CONTENT-TYPE",
				entry:    example,
				expected: `application/json`,
			}),
			Entry("CACHE-CONTROL", testCase{
				header:   "CACHE-CONTROL", // missing header
				entry:    example,
				expected: ``,
			}),
			Entry("CACHE-CONTROL or CONTENT-TYPE", testCase{
				header:    "CACHE-CONTROL", // missing header
				altHeader: "CONTENT-TYPE",
				entry:     example,
				expected:  `application/json`,
			}),
			Entry("CACHE-CONTROL or CONTENT-TYPE w/ MaxLength", testCase{
				header:    "CACHE-CONTROL", // missing header
				altHeader: "CONTENT-TYPE",
				maxLength: 3,
				entry:     example,
				expected:  `app`,
			}),
			Entry("SERVER or CONTENT-TYPE", testCase{
				header:    "SERVER",
				altHeader: "CONTENT-TYPE",
				entry:     example,
				expected:  `Tomcat`,
			}),
			Entry("SERVER or CONTENT-TYPE w/ MaxLength", testCase{
				header:    "SERVER",
				altHeader: "CONTENT-TYPE",
				maxLength: 3,
				entry:     example,
				expected:  `Tom`,
			}),
		)
	})

	Describe("FormatTcpLogEntry()", func() {
		It("should always return an empty string", func() {
			// setup
			formatter := &ResponseHeaderFormatter{HeaderFormatter{
				Header: "CONTENT-TYPE", AltHeader: "SERVER", MaxLength: 123}}
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
				formatter := &ResponseHeaderFormatter{HeaderFormatter{
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
				header:    "CONTENT-TYPE",
				altHeader: "",
				config:    &accesslog_config.HttpGrpcAccessLogConfig{},
				expected: &accesslog_config.HttpGrpcAccessLogConfig{
					AdditionalResponseHeadersToLog: []string{"CONTENT-TYPE"},
				},
			}),
			Entry("altHeader", testCase{
				header:    "",
				altHeader: "SERVER",
				config:    &accesslog_config.HttpGrpcAccessLogConfig{},
				expected: &accesslog_config.HttpGrpcAccessLogConfig{
					AdditionalResponseHeadersToLog: []string{"SERVER"},
				},
			}),
			Entry("header and altHeader", testCase{
				header:    "CONTENT-TYPE",
				altHeader: "SERVER",
				config:    &accesslog_config.HttpGrpcAccessLogConfig{},
				expected: &accesslog_config.HttpGrpcAccessLogConfig{
					AdditionalResponseHeadersToLog: []string{"CONTENT-TYPE", "SERVER"},
				},
			}),
			Entry("none w/ initial config", testCase{
				header:    "",
				altHeader: "",
				config: &accesslog_config.HttpGrpcAccessLogConfig{
					AdditionalResponseHeadersToLog: []string{"CACHE-CONTROL", "EXPIRES"},
				},
				expected: &accesslog_config.HttpGrpcAccessLogConfig{
					AdditionalResponseHeadersToLog: []string{"CACHE-CONTROL", "EXPIRES"},
				},
			}),
			Entry("header w/ initial config", testCase{
				header:    "CONTENT-TYPE",
				altHeader: "",
				config: &accesslog_config.HttpGrpcAccessLogConfig{
					AdditionalResponseHeadersToLog: []string{"CACHE-CONTROL", "EXPIRES"},
				},
				expected: &accesslog_config.HttpGrpcAccessLogConfig{
					AdditionalResponseHeadersToLog: []string{"CACHE-CONTROL", "EXPIRES", "CONTENT-TYPE"},
				},
			}),
			Entry("altHeader w/ initial config", testCase{
				header:    "",
				altHeader: "SERVER",
				config: &accesslog_config.HttpGrpcAccessLogConfig{
					AdditionalResponseHeadersToLog: []string{"CACHE-CONTROL", "EXPIRES"},
				},
				expected: &accesslog_config.HttpGrpcAccessLogConfig{
					AdditionalResponseHeadersToLog: []string{"CACHE-CONTROL", "EXPIRES", "SERVER"},
				},
			}),
			Entry("header and altHeader w/ initial config", testCase{
				header:    "CONTENT-TYPE",
				altHeader: "SERVER",
				config: &accesslog_config.HttpGrpcAccessLogConfig{
					AdditionalResponseHeadersToLog: []string{"CACHE-CONTROL", "EXPIRES"},
				},
				expected: &accesslog_config.HttpGrpcAccessLogConfig{
					AdditionalResponseHeadersToLog: []string{"CACHE-CONTROL", "EXPIRES", "CONTENT-TYPE", "SERVER"},
				},
			}),
		)
	})
})
