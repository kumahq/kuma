package accesslog_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	. "github.com/Kong/kuma/pkg/envoy/accesslog"

	"github.com/golang/protobuf/ptypes/wrappers"

	accesslog_data "github.com/envoyproxy/go-control-plane/envoy/data/accesslog/v2"
)

var _ = Describe("FieldFormatter", func() {

	Describe("HTTPAccessLogEntry", func() {

		type testCase struct {
			field    string
			entry    *accesslog_data.HTTPAccessLogEntry
			expected string
		}

		DescribeTable("should format properly",
			func(given testCase) {
				// given
				formatter := FieldFormatter(given.field)
				// when
				actual, err := formatter.FormatHttpLogEntry(given.entry)
				// then
				Expect(err).ToNot(HaveOccurred())
				// and
				Expect(actual).To(Equal(given.expected))
			},
			Entry("BYTES_RECEIVED: 0", testCase{
				field:    "BYTES_RECEIVED",
				expected: `0`,
			}),
			Entry("BYTES_RECEIVED: 123", testCase{
				field: "BYTES_RECEIVED",
				entry: &accesslog_data.HTTPAccessLogEntry{
					Request: &accesslog_data.HTTPRequestProperties{
						RequestBodyBytes: 123,
					},
				},
				expected: `123`,
			}),
			Entry("BYTES_SENT: 0", testCase{
				field:    "BYTES_SENT",
				expected: `0`,
			}),
			Entry("BYTES_SENT: 123", testCase{
				field: "BYTES_SENT",
				entry: &accesslog_data.HTTPAccessLogEntry{
					Response: &accesslog_data.HTTPResponseProperties{
						ResponseBodyBytes: 123,
					},
				},
				expected: `123`,
			}),
			Entry("PROTOCOL: UNSPECIFIED", testCase{
				field:    "PROTOCOL",
				expected: ``,
			}),
			Entry("PROTOCOL: HTTP10", testCase{
				field: "PROTOCOL",
				entry: &accesslog_data.HTTPAccessLogEntry{
					ProtocolVersion: accesslog_data.HTTPAccessLogEntry_HTTP10,
				},
				expected: `HTTP/1.0`,
			}),
			Entry("PROTOCOL: HTTP11", testCase{
				field: "PROTOCOL",
				entry: &accesslog_data.HTTPAccessLogEntry{
					ProtocolVersion: accesslog_data.HTTPAccessLogEntry_HTTP11,
				},
				expected: `HTTP/1.1`,
			}),
			Entry("PROTOCOL: HTTP2", testCase{
				field: "PROTOCOL",
				entry: &accesslog_data.HTTPAccessLogEntry{
					ProtocolVersion: accesslog_data.HTTPAccessLogEntry_HTTP2,
				},
				expected: `HTTP/2`,
			}),
			Entry("PROTOCOL: HTTP3", testCase{
				field: "PROTOCOL",
				entry: &accesslog_data.HTTPAccessLogEntry{
					ProtocolVersion: accesslog_data.HTTPAccessLogEntry_HTTP3,
				},
				expected: `HTTP/3`,
			}),
			Entry("RESPONSE_CODE: 0", testCase{
				field:    "RESPONSE_CODE",
				expected: `0`,
			}),
			Entry("RESPONSE_CODE: 200", testCase{
				field: "RESPONSE_CODE",
				entry: &accesslog_data.HTTPAccessLogEntry{
					Response: &accesslog_data.HTTPResponseProperties{
						ResponseCode: &wrappers.UInt32Value{
							Value: 200,
						},
					},
				},
				expected: `200`,
			}),
			Entry("RESPONSE_CODE_DETAILS: ``", testCase{
				field:    "RESPONSE_CODE_DETAILS",
				expected: ``,
			}),
			Entry("RESPONSE_CODE_DETAILS: `response code details`", testCase{
				field: "RESPONSE_CODE_DETAILS",
				entry: &accesslog_data.HTTPAccessLogEntry{
					Response: &accesslog_data.HTTPResponseProperties{
						ResponseCodeDetails: "response code details",
					},
				},
				expected: `response code details`,
			}),
			Entry("DOWNSTREAM_PEER_FINGERPRINT_256", testCase{
				field:    "DOWNSTREAM_PEER_FINGERPRINT_256",
				expected: `UNSUPPORTED_FIELD(DOWNSTREAM_PEER_FINGERPRINT_256)`,
			}),
			Entry("DOWNSTREAM_PEER_SERIAL", testCase{
				field:    "DOWNSTREAM_PEER_SERIAL",
				expected: `UNSUPPORTED_FIELD(DOWNSTREAM_PEER_SERIAL)`,
			}),
			Entry("DOWNSTREAM_PEER_ISSUER", testCase{
				field:    "DOWNSTREAM_PEER_ISSUER",
				expected: `UNSUPPORTED_FIELD(DOWNSTREAM_PEER_ISSUER)`,
			}),
			Entry("DOWNSTREAM_PEER_CERT", testCase{
				field:    "DOWNSTREAM_PEER_CERT",
				expected: `UNSUPPORTED_FIELD(DOWNSTREAM_PEER_CERT)`,
			}),
			Entry("DOWNSTREAM_PEER_CERT_V_START", testCase{
				field:    "DOWNSTREAM_PEER_CERT_V_START",
				expected: `UNSUPPORTED_FIELD(DOWNSTREAM_PEER_CERT_V_START)`,
			}),
			Entry("DOWNSTREAM_PEER_CERT_V_END", testCase{
				field:    "DOWNSTREAM_PEER_CERT_V_END",
				expected: `UNSUPPORTED_FIELD(DOWNSTREAM_PEER_CERT_V_END)`,
			}),
			Entry("HOSTNAME", testCase{
				field:    "HOSTNAME",
				expected: `UNSUPPORTED_FIELD(HOSTNAME)`,
			}),
		)
	})

	Describe("TCPAccessLogEntry", func() {

		type testCase struct {
			field    string
			entry    *accesslog_data.TCPAccessLogEntry
			expected string
		}

		DescribeTable("should format properly",
			func(given testCase) {
				// given
				formatter := FieldFormatter(given.field)
				// when
				actual, err := formatter.FormatTcpLogEntry(given.entry)
				// then
				Expect(err).ToNot(HaveOccurred())
				// and
				Expect(actual).To(Equal(given.expected))
			},
			Entry("BYTES_RECEIVED: 0", testCase{
				field:    "BYTES_RECEIVED",
				expected: `0`,
			}),
			Entry("BYTES_RECEIVED: 123", testCase{
				field: "BYTES_RECEIVED",
				entry: &accesslog_data.TCPAccessLogEntry{
					ConnectionProperties: &accesslog_data.ConnectionProperties{
						ReceivedBytes: 123,
					},
				},
				expected: `123`,
			}),
			Entry("BYTES_SENT: 0", testCase{
				field:    "BYTES_SENT",
				expected: `0`,
			}),
			Entry("BYTES_SENT: 123", testCase{
				field: "BYTES_SENT",
				entry: &accesslog_data.TCPAccessLogEntry{
					ConnectionProperties: &accesslog_data.ConnectionProperties{
						SentBytes: 123,
					},
				},
				expected: `123`,
			}),
			Entry("PROTOCOL", testCase{
				field:    "PROTOCOL",
				expected: ``,
			}),
			Entry("RESPONSE_CODE", testCase{
				field:    "RESPONSE_CODE",
				expected: `0`, // TODO: is it consistent with file access log?
			}),
			Entry("RESPONSE_CODE_DETAILS", testCase{
				field:    "RESPONSE_CODE_DETAILS",
				expected: ``,
			}),
			Entry("DOWNSTREAM_PEER_FINGERPRINT_256", testCase{
				field:    "DOWNSTREAM_PEER_FINGERPRINT_256",
				expected: `UNSUPPORTED_FIELD(DOWNSTREAM_PEER_FINGERPRINT_256)`,
			}),
			Entry("DOWNSTREAM_PEER_SERIAL", testCase{
				field:    "DOWNSTREAM_PEER_SERIAL",
				expected: `UNSUPPORTED_FIELD(DOWNSTREAM_PEER_SERIAL)`,
			}),
			Entry("DOWNSTREAM_PEER_ISSUER", testCase{
				field:    "DOWNSTREAM_PEER_ISSUER",
				expected: `UNSUPPORTED_FIELD(DOWNSTREAM_PEER_ISSUER)`,
			}),
			Entry("DOWNSTREAM_PEER_CERT", testCase{
				field:    "DOWNSTREAM_PEER_CERT",
				expected: `UNSUPPORTED_FIELD(DOWNSTREAM_PEER_CERT)`,
			}),
			Entry("DOWNSTREAM_PEER_CERT_V_START", testCase{
				field:    "DOWNSTREAM_PEER_CERT_V_START",
				expected: `UNSUPPORTED_FIELD(DOWNSTREAM_PEER_CERT_V_START)`,
			}),
			Entry("DOWNSTREAM_PEER_CERT_V_END", testCase{
				field:    "DOWNSTREAM_PEER_CERT_V_END",
				expected: `UNSUPPORTED_FIELD(DOWNSTREAM_PEER_CERT_V_END)`,
			}),
			Entry("HOSTNAME", testCase{
				field:    "HOSTNAME",
				expected: `UNSUPPORTED_FIELD(HOSTNAME)`,
			}),
		)
	})
})
