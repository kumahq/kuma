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
			Entry(":method - `POST`", testCase{
				header:   ":method",
				entry:    example,
				expected: `POST`,
			}),
			Entry(":method - ``", testCase{
				header:   ":method",
				entry:    &accesslog_data.HTTPAccessLogEntry{},
				expected: ``,
			}),
			Entry(":scheme", testCase{
				header:   ":scheme",
				entry:    example,
				expected: `https`,
			}),
			Entry(":authority", testCase{
				header:   ":authority",
				entry:    example,
				expected: `backend.internal:8080`,
			}),
			Entry(":path", testCase{
				header:   ":path",
				entry:    example,
				expected: `/api/version`,
			}),
			Entry("user-agent", testCase{
				header:   "user-agent",
				entry:    example,
				expected: `curl`,
			}),
			Entry("referer", testCase{
				header:   "referer",
				entry:    example,
				expected: `www.google.com`,
			}),
			Entry("x-forwarded-for", testCase{
				header:   "x-forwarded-for",
				entry:    example,
				expected: `10.0.0.1`,
			}),
			Entry("x-request-id", testCase{
				header:   "x-request-id",
				entry:    example,
				expected: `025169aa-8317-4ebd-b0dd-2f0872ec444a`,
			}),
			Entry("x-envoy-original-path", testCase{
				header:   "x-envoy-original-path",
				entry:    example,
				expected: `/xyz`,
			}),
			Entry("x-envoy-original-path", testCase{
				header:   "x-envoy-original-path",
				entry:    example,
				expected: `/xyz`,
			}),
			Entry("content-type", testCase{
				header:   "content-type",
				entry:    example,
				expected: `application/json`,
			}),
			Entry("origin", testCase{
				header:   "origin", // missing header
				entry:    example,
				expected: ``,
			}),
			Entry("origin or :authority", testCase{
				header:    "origin", // missing header
				altHeader: ":authority",
				entry:     example,
				expected:  `backend.internal:8080`,
			}),
			Entry("origin or :authority w/ MaxLength", testCase{
				header:    "origin", // missing header
				altHeader: ":authority",
				maxLength: 7,
				entry:     example,
				expected:  `backend`,
			}),
			Entry(":path or x-envoy-original-path", testCase{
				header:    ":path",
				altHeader: "x-envoy-original-path",
				entry:     example,
				expected:  `/api/version`,
			}),
			Entry(":path or x-envoy-original-path w/ MaxLength", testCase{
				header:    ":path",
				altHeader: "x-envoy-original-path",
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
				Header: ":path", AltHeader: "x-envoy-original-path", MaxLength: 123}}
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
				header:    ":path",
				altHeader: "",
				config:    &accesslog_config.HttpGrpcAccessLogConfig{},
				expected: &accesslog_config.HttpGrpcAccessLogConfig{
					AdditionalRequestHeadersToLog: []string{":path"},
				},
			}),
			Entry("altHeader", testCase{
				header:    "",
				altHeader: "x-envoy-original-path",
				config:    &accesslog_config.HttpGrpcAccessLogConfig{},
				expected: &accesslog_config.HttpGrpcAccessLogConfig{
					AdditionalRequestHeadersToLog: []string{"x-envoy-original-path"},
				},
			}),
			Entry("header and altHeader", testCase{
				header:    ":path",
				altHeader: "x-envoy-original-path",
				config:    &accesslog_config.HttpGrpcAccessLogConfig{},
				expected: &accesslog_config.HttpGrpcAccessLogConfig{
					AdditionalRequestHeadersToLog: []string{":path", "x-envoy-original-path"},
				},
			}),
			Entry("none w/ initial config", testCase{
				header:    "",
				altHeader: "",
				config: &accesslog_config.HttpGrpcAccessLogConfig{
					AdditionalRequestHeadersToLog: []string{":scheme", ":authority"},
				},
				expected: &accesslog_config.HttpGrpcAccessLogConfig{
					AdditionalRequestHeadersToLog: []string{":scheme", ":authority"},
				},
			}),
			Entry("header w/ initial config", testCase{
				header:    ":path",
				altHeader: "",
				config: &accesslog_config.HttpGrpcAccessLogConfig{
					AdditionalRequestHeadersToLog: []string{":scheme", ":authority"},
				},
				expected: &accesslog_config.HttpGrpcAccessLogConfig{
					AdditionalRequestHeadersToLog: []string{":scheme", ":authority", ":path"},
				},
			}),
			Entry("altHeader w/ initial config", testCase{
				header:    "",
				altHeader: "x-envoy-original-path",
				config: &accesslog_config.HttpGrpcAccessLogConfig{
					AdditionalRequestHeadersToLog: []string{":scheme", ":authority"},
				},
				expected: &accesslog_config.HttpGrpcAccessLogConfig{
					AdditionalRequestHeadersToLog: []string{":scheme", ":authority", "x-envoy-original-path"},
				},
			}),
			Entry("header and altHeader w/ initial config", testCase{
				header:    ":path",
				altHeader: "x-envoy-original-path",
				config: &accesslog_config.HttpGrpcAccessLogConfig{
					AdditionalRequestHeadersToLog: []string{":scheme", ":authority"},
				},
				expected: &accesslog_config.HttpGrpcAccessLogConfig{
					AdditionalRequestHeadersToLog: []string{":scheme", ":authority", ":path", "x-envoy-original-path"},
				},
			}),
		)
	})
})
