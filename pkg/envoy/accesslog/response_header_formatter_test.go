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
			Entry("no header or altHeader", testCase{
				header:    "",
				altHeader: "",
				entry:     example,
				expected:  ``, // apparently, Envoy allows both `Header` and `AltHeader` to be empty
			}),
			Entry("content-type", testCase{
				header:   "content-type",
				entry:    example,
				expected: `application/json`,
			}),
			Entry("cache-control", testCase{
				header:   "cache-control", // missing header
				entry:    example,
				expected: ``,
			}),
			Entry("cache-control or content-type", testCase{
				header:    "cache-control", // missing header
				altHeader: "content-type",
				entry:     example,
				expected:  `application/json`,
			}),
			Entry("cache-control or content-type w/ MaxLength", testCase{
				header:    "cache-control", // missing header
				altHeader: "content-type",
				maxLength: 3,
				entry:     example,
				expected:  `app`,
			}),
			Entry("server or content-type", testCase{
				header:    "server",
				altHeader: "content-type",
				entry:     example,
				expected:  `Tomcat`,
			}),
			Entry("server or content-type w/ MaxLength", testCase{
				header:    "server",
				altHeader: "content-type",
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
				Header: "content-type", AltHeader: "server", MaxLength: 123}}
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
				header:    "content-type",
				altHeader: "",
				config:    &accesslog_config.HttpGrpcAccessLogConfig{},
				expected: &accesslog_config.HttpGrpcAccessLogConfig{
					AdditionalResponseHeadersToLog: []string{"content-type"},
				},
			}),
			Entry("altHeader", testCase{
				header:    "",
				altHeader: "server",
				config:    &accesslog_config.HttpGrpcAccessLogConfig{},
				expected: &accesslog_config.HttpGrpcAccessLogConfig{
					AdditionalResponseHeadersToLog: []string{"server"},
				},
			}),
			Entry("header and altHeader", testCase{
				header:    "content-type",
				altHeader: "server",
				config:    &accesslog_config.HttpGrpcAccessLogConfig{},
				expected: &accesslog_config.HttpGrpcAccessLogConfig{
					AdditionalResponseHeadersToLog: []string{"content-type", "server"},
				},
			}),
			Entry("none w/ initial config", testCase{
				header:    "",
				altHeader: "",
				config: &accesslog_config.HttpGrpcAccessLogConfig{
					AdditionalResponseHeadersToLog: []string{"cache-control", "expires"},
				},
				expected: &accesslog_config.HttpGrpcAccessLogConfig{
					AdditionalResponseHeadersToLog: []string{"cache-control", "expires"},
				},
			}),
			Entry("header w/ initial config", testCase{
				header:    "content-type",
				altHeader: "",
				config: &accesslog_config.HttpGrpcAccessLogConfig{
					AdditionalResponseHeadersToLog: []string{"cache-control", "expires"},
				},
				expected: &accesslog_config.HttpGrpcAccessLogConfig{
					AdditionalResponseHeadersToLog: []string{"cache-control", "expires", "content-type"},
				},
			}),
			Entry("altHeader w/ initial config", testCase{
				header:    "",
				altHeader: "server",
				config: &accesslog_config.HttpGrpcAccessLogConfig{
					AdditionalResponseHeadersToLog: []string{"cache-control", "expires"},
				},
				expected: &accesslog_config.HttpGrpcAccessLogConfig{
					AdditionalResponseHeadersToLog: []string{"cache-control", "expires", "server"},
				},
			}),
			Entry("header and altHeader w/ initial config", testCase{
				header:    "content-type",
				altHeader: "server",
				config: &accesslog_config.HttpGrpcAccessLogConfig{
					AdditionalResponseHeadersToLog: []string{"cache-control", "expires"},
				},
				expected: &accesslog_config.HttpGrpcAccessLogConfig{
					AdditionalResponseHeadersToLog: []string{"cache-control", "expires", "content-type", "server"},
				},
			}),
		)
	})
})
