package accesslog_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	. "github.com/Kong/kuma/pkg/envoy/accesslog"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"

	accesslog_config "github.com/envoyproxy/go-control-plane/envoy/config/accesslog/v2"
	accesslog_data "github.com/envoyproxy/go-control-plane/envoy/data/accesslog/v2"
)

var _ = Describe("RequestHeaderFormatter", func() {

	Describe("FormatHttpLogEntry()", func() {
		example := &accesslog_data.HTTPAccessLogEntry{
			Request: &accesslog_data.HTTPRequestProperties{
				RequestMethod:       envoy_core.RequestMethod_POST,
				Scheme:              "https",
				Authority:           "backend.internal:8080",
				Path:                "/api/version",
				UserAgent:           "curl",
				Referer:             "www.google.com",
				ForwardedFor:        "10.0.0.1",
				RequestId:           "025169aa-8317-4ebd-b0dd-2f0872ec444a",
				OriginalPath:        "/xyz",
				RequestHeadersBytes: 123,
				RequestBodyBytes:    456,
				RequestHeaders: map[string]string{
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
				formatter := &RequestHeaderFormatter{HeaderFormatter{
					Header: given.header, AltHeader: given.altHeader, MaxLength: given.maxLength}}
				// when
				actual, err := formatter.FormatHttpLogEntry(given.entry)
				// then
				Expect(err).ToNot(HaveOccurred())
				// and
				Expect(actual).To(Equal(given.expected))
			},
			Entry(":METHOD - `POST`", testCase{
				header:   ":METHOD",
				entry:    example,
				expected: `POST`,
			}),
			Entry(":METHOD - ``", testCase{
				header:   ":METHOD",
				entry:    &accesslog_data.HTTPAccessLogEntry{},
				expected: ``,
			}),
			Entry(":SCHEME", testCase{
				header:   ":SCHEME",
				entry:    example,
				expected: `https`,
			}),
			Entry(":AUTHORITY", testCase{
				header:   ":AUTHORITY",
				entry:    example,
				expected: `backend.internal:8080`,
			}),
			Entry(":PATH", testCase{
				header:   ":PATH",
				entry:    example,
				expected: `/api/version`,
			}),
			Entry("USER-AGENT", testCase{
				header:   "USER-AGENT",
				entry:    example,
				expected: `curl`,
			}),
			Entry("REFERER", testCase{
				header:   "REFERER",
				entry:    example,
				expected: `www.google.com`,
			}),
			Entry("X-FORWARDED-FOR", testCase{
				header:   "X-FORWARDED-FOR",
				entry:    example,
				expected: `10.0.0.1`,
			}),
			Entry("X-REQUEST-ID", testCase{
				header:   "X-REQUEST-ID",
				entry:    example,
				expected: `025169aa-8317-4ebd-b0dd-2f0872ec444a`,
			}),
			Entry("X-ENVOY-ORIGINAL-PATH", testCase{
				header:   "X-ENVOY-ORIGINAL-PATH",
				entry:    example,
				expected: `/xyz`,
			}),
			Entry("X-ENVOY-ORIGINAL-PATH", testCase{
				header:   "X-ENVOY-ORIGINAL-PATH",
				entry:    example,
				expected: `/xyz`,
			}),
			Entry("CONTENT-TYPE", testCase{
				header:   "CONTENT-TYPE",
				entry:    example,
				expected: `application/json`,
			}),
			Entry("ORIGIN", testCase{
				header:   "ORIGIN", // missing header
				entry:    example,
				expected: ``,
			}),
			Entry("ORIGIN or :AUTHORITY", testCase{
				header:    "ORIGIN", // missing header
				altHeader: ":AUTHORITY",
				entry:     example,
				expected:  `backend.internal:8080`,
			}),
			Entry("ORIGIN or :AUTHORITY w/ MaxLength", testCase{
				header:    "ORIGIN", // missing header
				altHeader: ":AUTHORITY",
				maxLength: 7,
				entry:     example,
				expected:  `backend`,
			}),
			Entry(":PATH or X-ENVOY-ORIGINAL-PATH", testCase{
				header:    ":PATH",
				altHeader: "X-ENVOY-ORIGINAL-PATH",
				entry:     example,
				expected:  `/api/version`,
			}),
			Entry(":PATH or X-ENVOY-ORIGINAL-PATH w/ MaxLength", testCase{
				header:    ":PATH",
				altHeader: "X-ENVOY-ORIGINAL-PATH",
				maxLength: 5,
				entry:     example,
				expected:  `/api/`,
			}),
		)
	})

	Describe("FormatTcpLogEntry()", func() {
		It("should always return an empty string", func() {
			// setup
			formatter := &RequestHeaderFormatter{HeaderFormatter{
				Header: ":PATH", AltHeader: "X-ENVOY-ORIGINAL-PATH", MaxLength: 123}}
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
				formatter := &RequestHeaderFormatter{HeaderFormatter{
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
				header:    ":PATH",
				altHeader: "",
				config:    &accesslog_config.HttpGrpcAccessLogConfig{},
				expected: &accesslog_config.HttpGrpcAccessLogConfig{
					AdditionalRequestHeadersToLog: []string{":PATH"},
				},
			}),
			Entry("altHeader", testCase{
				header:    "",
				altHeader: "X-ENVOY-ORIGINAL-PATH",
				config:    &accesslog_config.HttpGrpcAccessLogConfig{},
				expected: &accesslog_config.HttpGrpcAccessLogConfig{
					AdditionalRequestHeadersToLog: []string{"X-ENVOY-ORIGINAL-PATH"},
				},
			}),
			Entry("header and altHeader", testCase{
				header:    ":PATH",
				altHeader: "X-ENVOY-ORIGINAL-PATH",
				config:    &accesslog_config.HttpGrpcAccessLogConfig{},
				expected: &accesslog_config.HttpGrpcAccessLogConfig{
					AdditionalRequestHeadersToLog: []string{":PATH", "X-ENVOY-ORIGINAL-PATH"},
				},
			}),
			Entry("none w/ initial config", testCase{
				header:    "",
				altHeader: "",
				config: &accesslog_config.HttpGrpcAccessLogConfig{
					AdditionalRequestHeadersToLog: []string{":SCHEME", ":AUTHORITY"},
				},
				expected: &accesslog_config.HttpGrpcAccessLogConfig{
					AdditionalRequestHeadersToLog: []string{":SCHEME", ":AUTHORITY"},
				},
			}),
			Entry("header w/ initial config", testCase{
				header:    ":PATH",
				altHeader: "",
				config: &accesslog_config.HttpGrpcAccessLogConfig{
					AdditionalRequestHeadersToLog: []string{":SCHEME", ":AUTHORITY"},
				},
				expected: &accesslog_config.HttpGrpcAccessLogConfig{
					AdditionalRequestHeadersToLog: []string{":SCHEME", ":AUTHORITY", ":PATH"},
				},
			}),
			Entry("altHeader w/ initial config", testCase{
				header:    "",
				altHeader: "X-ENVOY-ORIGINAL-PATH",
				config: &accesslog_config.HttpGrpcAccessLogConfig{
					AdditionalRequestHeadersToLog: []string{":SCHEME", ":AUTHORITY"},
				},
				expected: &accesslog_config.HttpGrpcAccessLogConfig{
					AdditionalRequestHeadersToLog: []string{":SCHEME", ":AUTHORITY", "X-ENVOY-ORIGINAL-PATH"},
				},
			}),
			Entry("header and altHeader w/ initial config", testCase{
				header:    ":PATH",
				altHeader: "X-ENVOY-ORIGINAL-PATH",
				config: &accesslog_config.HttpGrpcAccessLogConfig{
					AdditionalRequestHeadersToLog: []string{":SCHEME", ":AUTHORITY"},
				},
				expected: &accesslog_config.HttpGrpcAccessLogConfig{
					AdditionalRequestHeadersToLog: []string{":SCHEME", ":AUTHORITY", ":PATH", "X-ENVOY-ORIGINAL-PATH"},
				},
			}),
		)
	})
})
