package v3_test

import (
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	accesslog_data "github.com/envoyproxy/go-control-plane/envoy/data/accesslog/v3"
	accesslog_config "github.com/envoyproxy/go-control-plane/envoy/extensions/access_loggers/grpc/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/pkg/envoy/accesslog/v3"
)

var _ = Describe("RequestHeaderOperator", func() {

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
				fragment := &RequestHeaderOperator{HeaderFormatter{
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
			fragment := &RequestHeaderOperator{HeaderFormatter{
				Header: ":path", AltHeader: "x-envoy-original-path", MaxLength: 123}}
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
				fragment := &RequestHeaderOperator{HeaderFormatter{
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
			Entry("header that is not captured by default", testCase{
				header:    "content-type",
				altHeader: "",
				config:    &accesslog_config.HttpGrpcAccessLogConfig{},
				expected: &accesslog_config.HttpGrpcAccessLogConfig{
					AdditionalRequestHeadersToLog: []string{"content-type"},
				},
			}),
			Entry("header that is captured by default: `:method`", testCase{
				header:    ":method",
				altHeader: "",
				config:    &accesslog_config.HttpGrpcAccessLogConfig{},
				expected:  &accesslog_config.HttpGrpcAccessLogConfig{}, // should not be added as an additional request header to log
			}),
			Entry("header that is captured by default: `:scheme`", testCase{
				header:    ":scheme",
				altHeader: "",
				config:    &accesslog_config.HttpGrpcAccessLogConfig{},
				expected:  &accesslog_config.HttpGrpcAccessLogConfig{}, // should not be added as an additional request header to log
			}),
			Entry("header that is captured by default: `:authority`", testCase{
				header:    ":authority",
				altHeader: "",
				config:    &accesslog_config.HttpGrpcAccessLogConfig{},
				expected:  &accesslog_config.HttpGrpcAccessLogConfig{}, // should not be added as an additional request header to log
			}),
			Entry("header that is captured by default: `:path`", testCase{
				header:    ":path",
				altHeader: "",
				config:    &accesslog_config.HttpGrpcAccessLogConfig{},
				expected:  &accesslog_config.HttpGrpcAccessLogConfig{}, // should not be added as an additional request header to log
			}),
			Entry("header that is captured by default: `user-agent`", testCase{
				header:    "user-agent",
				altHeader: "",
				config:    &accesslog_config.HttpGrpcAccessLogConfig{},
				expected:  &accesslog_config.HttpGrpcAccessLogConfig{}, // should not be added as an additional request header to log
			}),
			Entry("header that is captured by default: `referer`", testCase{
				header:    "referer",
				altHeader: "",
				config:    &accesslog_config.HttpGrpcAccessLogConfig{},
				expected:  &accesslog_config.HttpGrpcAccessLogConfig{}, // should not be added as an additional request header to log
			}),
			Entry("header that is captured by default: `x-forwarded-for`", testCase{
				header:    "x-forwarded-for",
				altHeader: "",
				config:    &accesslog_config.HttpGrpcAccessLogConfig{},
				expected:  &accesslog_config.HttpGrpcAccessLogConfig{}, // should not be added as an additional request header to log
			}),
			Entry("header that is captured by default: `x-request-id`", testCase{
				header:    "x-request-id",
				altHeader: "",
				config:    &accesslog_config.HttpGrpcAccessLogConfig{},
				expected:  &accesslog_config.HttpGrpcAccessLogConfig{}, // should not be added as an additional request header to log
			}),
			Entry("header that is captured by default: `x-envoy-original-path`", testCase{
				header:    "x-envoy-original-path",
				altHeader: "",
				config:    &accesslog_config.HttpGrpcAccessLogConfig{},
				expected:  &accesslog_config.HttpGrpcAccessLogConfig{}, // should not be added as an additional request header to log
			}),
			Entry("altHeader that is not captured by default", testCase{
				header:    "",
				altHeader: "origin",
				config:    &accesslog_config.HttpGrpcAccessLogConfig{},
				expected: &accesslog_config.HttpGrpcAccessLogConfig{
					AdditionalRequestHeadersToLog: []string{"origin"},
				},
			}),
			Entry("altHeader that is captured by default", testCase{
				header:    "",
				altHeader: ":authority",
				config:    &accesslog_config.HttpGrpcAccessLogConfig{},
				expected:  &accesslog_config.HttpGrpcAccessLogConfig{}, // should not be added as an additional request header to log
			}),
			Entry("header and altHeader that are not captured by default", testCase{
				header:    "content-type",
				altHeader: "origin",
				config:    &accesslog_config.HttpGrpcAccessLogConfig{},
				expected: &accesslog_config.HttpGrpcAccessLogConfig{
					AdditionalRequestHeadersToLog: []string{"content-type", "origin"},
				},
			}),
			Entry("header and altHeader that are captured by default", testCase{
				header:    ":authority",
				altHeader: ":path",
				config:    &accesslog_config.HttpGrpcAccessLogConfig{},
				expected:  &accesslog_config.HttpGrpcAccessLogConfig{}, // should not be added as an additional request header to log
			}),
			Entry("none w/ initial config", testCase{
				header:    "",
				altHeader: "",
				config: &accesslog_config.HttpGrpcAccessLogConfig{
					AdditionalRequestHeadersToLog: []string{"x-custom-header-1", "x-custom-header-2"},
				},
				expected: &accesslog_config.HttpGrpcAccessLogConfig{
					AdditionalRequestHeadersToLog: []string{"x-custom-header-1", "x-custom-header-2"},
				},
			}),
			Entry("header w/ initial config", testCase{
				header:    "content-type",
				altHeader: "",
				config: &accesslog_config.HttpGrpcAccessLogConfig{
					AdditionalRequestHeadersToLog: []string{"x-custom-header-1", "x-custom-header-2"},
				},
				expected: &accesslog_config.HttpGrpcAccessLogConfig{
					AdditionalRequestHeadersToLog: []string{"x-custom-header-1", "x-custom-header-2", "content-type"},
				},
			}),
			Entry("altHeader w/ initial config", testCase{
				header:    "",
				altHeader: "origin",
				config: &accesslog_config.HttpGrpcAccessLogConfig{
					AdditionalRequestHeadersToLog: []string{"x-custom-header-1", "x-custom-header-2"},
				},
				expected: &accesslog_config.HttpGrpcAccessLogConfig{
					AdditionalRequestHeadersToLog: []string{"x-custom-header-1", "x-custom-header-2", "origin"},
				},
			}),
			Entry("header and altHeader w/ initial config", testCase{
				header:    "content-type",
				altHeader: "origin",
				config: &accesslog_config.HttpGrpcAccessLogConfig{
					AdditionalRequestHeadersToLog: []string{"x-custom-header-1", "x-custom-header-2"},
				},
				expected: &accesslog_config.HttpGrpcAccessLogConfig{
					AdditionalRequestHeadersToLog: []string{"x-custom-header-1", "x-custom-header-2", "content-type", "origin"},
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
				fragment := &RequestHeaderOperator{HeaderFormatter{
					Header: given.header, AltHeader: given.altHeader, MaxLength: given.maxLength}}

				// when
				actual := fragment.String()
				// then
				Expect(actual).To(Equal(given.expected))

			},
			Entry("%REQ()%", testCase{
				expected: `%REQ()%`,
			}),
			Entry("%REQ():10%", testCase{
				maxLength: 10,
				expected:  `%REQ():10%`,
			}),
			Entry("%REQ(:authority)%", testCase{
				header:   `:authority`,
				expected: `%REQ(:authority)%`,
			}),
			Entry("%REQ(:authority):10%", testCase{
				header:    `:authority`,
				maxLength: 10,
				expected:  `%REQ(:authority):10%`,
			}),
			Entry("%REQ(?origin)%", testCase{
				altHeader: `origin`,
				expected:  `%REQ(?origin)%`,
			}),
			Entry("%REQ(?origin):10%", testCase{
				altHeader: `origin`,
				maxLength: 10,
				expected:  `%REQ(?origin):10%`,
			}),
			Entry("%REQ(:authority?origin)%", testCase{
				header:    ":authority",
				altHeader: `origin`,
				expected:  `%REQ(:authority?origin)%`,
			}),
			Entry("%REQ(:authority?origin):10%", testCase{
				header:    ":authority",
				altHeader: `origin`,
				maxLength: 10,
				expected:  `%REQ(:authority?origin):10%`,
			}),
		)
	})
})
