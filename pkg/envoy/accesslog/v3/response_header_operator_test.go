package v3_test

import (
	accesslog_data "github.com/envoyproxy/go-control-plane/envoy/data/accesslog/v3"
	accesslog_config "github.com/envoyproxy/go-control-plane/envoy/extensions/access_loggers/grpc/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/pkg/envoy/accesslog/v3"
)

var _ = Describe("ResponseHeaderOperator", func() {

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
				fragment := &ResponseHeaderOperator{HeaderFormatter{
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
			fragment := &ResponseHeaderOperator{HeaderFormatter{
				Header: "content-type", AltHeader: "server", MaxLength: 123}}
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
				fragment := &ResponseHeaderOperator{HeaderFormatter{
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
				fragment := &ResponseHeaderOperator{HeaderFormatter{
					Header: given.header, AltHeader: given.altHeader, MaxLength: given.maxLength}}

				// when
				actual := fragment.String()
				// then
				Expect(actual).To(Equal(given.expected))

			},
			Entry("%RESP()%", testCase{
				expected: `%RESP()%`,
			}),
			Entry("%RESP():10%", testCase{
				maxLength: 10,
				expected:  `%RESP():10%`,
			}),
			Entry("%RESP(content-type)%", testCase{
				header:   `content-type`,
				expected: `%RESP(content-type)%`,
			}),
			Entry("%RESP(content-type):10%", testCase{
				header:    `content-type`,
				maxLength: 10,
				expected:  `%RESP(content-type):10%`,
			}),
			Entry("%RESP(?server)%", testCase{
				altHeader: `server`,
				expected:  `%RESP(?server)%`,
			}),
			Entry("%RESP(?server):10%", testCase{
				altHeader: `server`,
				maxLength: 10,
				expected:  `%RESP(?server):10%`,
			}),
			Entry("%RESP(content-type?server)%", testCase{
				header:    "content-type",
				altHeader: `server`,
				expected:  `%RESP(content-type?server)%`,
			}),
			Entry("%RESP(content-type?server):10%", testCase{
				header:    "content-type",
				altHeader: `server`,
				maxLength: 10,
				expected:  `%RESP(content-type?server):10%`,
			}),
		)
	})
})
