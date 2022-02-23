package v3_test

import (
	"time"

	accesslog_data "github.com/envoyproxy/go-control-plane/envoy/data/accesslog/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/pkg/envoy/accesslog/v3"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

var _ = Describe("FieldOperator", func() {

	Describe("FormatHttpLogEntry()", func() {

		type testCase struct {
			field    string
			entry    *accesslog_data.HTTPAccessLogEntry
			expected string
		}

		DescribeTable("should format properly",
			func(given testCase) {
				// given
				fragment := FieldOperator(given.field)
				// when
				actual, err := fragment.FormatHttpLogEntry(given.entry)
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
			Entry("BYTES_SENT: 456", testCase{
				field: "BYTES_SENT",
				entry: &accesslog_data.HTTPAccessLogEntry{
					Response: &accesslog_data.HTTPResponseProperties{
						ResponseBodyBytes: 456,
					},
				},
				expected: `456`,
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
						ResponseCode: util_proto.UInt32(200),
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
			Entry("REQUEST_DURATION: ``", testCase{
				field:    "REQUEST_DURATION",
				expected: ``,
			}),
			Entry("REQUEST_DURATION: `57` millis", testCase{
				field: "REQUEST_DURATION",
				entry: &accesslog_data.HTTPAccessLogEntry{
					CommonProperties: &accesslog_data.AccessLogCommon{
						TimeToLastRxByte: util_proto.Duration(57000 * time.Microsecond),
					},
				},
				expected: `57`, // time in millis
			}),
			Entry("RESPONSE_DURATION: ``", testCase{
				field:    "RESPONSE_DURATION",
				expected: ``,
			}),
			Entry("RESPONSE_DURATION: `102` millis", testCase{
				field: "RESPONSE_DURATION",
				entry: &accesslog_data.HTTPAccessLogEntry{
					CommonProperties: &accesslog_data.AccessLogCommon{
						TimeToFirstUpstreamRxByte: util_proto.Duration(102000 * time.Microsecond),
					},
				},
				expected: `102`, // time in millis
			}),
			Entry("RESPONSE_TX_DURATION: ``", testCase{
				field:    "RESPONSE_TX_DURATION",
				expected: ``,
			}),
			Entry("RESPONSE_TX_DURATION: no TimeToFirstUpstreamRxByte", testCase{
				field: "RESPONSE_TX_DURATION",
				entry: &accesslog_data.HTTPAccessLogEntry{
					CommonProperties: &accesslog_data.AccessLogCommon{
						TimeToLastDownstreamTxByte: util_proto.Duration(123000 * time.Microsecond),
					},
				},
				expected: ``,
			}),
			Entry("RESPONSE_TX_DURATION: no TimeToLastDownstreamTxByte", testCase{
				field: "RESPONSE_TX_DURATION",
				entry: &accesslog_data.HTTPAccessLogEntry{
					CommonProperties: &accesslog_data.AccessLogCommon{
						TimeToFirstUpstreamRxByte: util_proto.Duration(102000 * time.Microsecond),
					},
				},
				expected: ``,
			}),
			Entry("RESPONSE_TX_DURATION: `23` millis", testCase{
				field: "RESPONSE_TX_DURATION",
				entry: &accesslog_data.HTTPAccessLogEntry{
					CommonProperties: &accesslog_data.AccessLogCommon{
						TimeToFirstUpstreamRxByte:  util_proto.Duration(102000 * time.Microsecond),
						TimeToLastDownstreamTxByte: util_proto.Duration(123000 * time.Microsecond),
					},
				},
				expected: `21`, // time in millis
			}),
			Entry("GRPC_STATUS: OK", testCase{
				field: "GRPC_STATUS",
				entry: &accesslog_data.HTTPAccessLogEntry{
					Response: &accesslog_data.HTTPResponseProperties{
						ResponseTrailers: map[string]string{
							"grpc-status": "0",
						},
					},
				},
				expected: "OK",
			}),
			Entry("GRPC_STATUS: no status", testCase{
				field:    "GRPC_STATUS",
				expected: "",
			}),
			Entry("GRPC_STATUS: InvalidCode", testCase{
				field: "GRPC_STATUS",
				entry: &accesslog_data.HTTPAccessLogEntry{
					Response: &accesslog_data.HTTPResponseProperties{
						ResponseTrailers: map[string]string{
							"grpc-status": "",
						},
					},
				},
				expected: "InvalidCode",
			}),
		)
	})

	Describe("FormatTcpLogEntry()", func() {

		type testCase struct {
			field    string
			entry    *accesslog_data.TCPAccessLogEntry
			expected string
		}

		DescribeTable("should format properly",
			func(given testCase) {
				// given
				fragment := FieldOperator(given.field)
				// when
				actual, err := fragment.FormatTcpLogEntry(given.entry)
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
			Entry("BYTES_SENT: 456", testCase{
				field: "BYTES_SENT",
				entry: &accesslog_data.TCPAccessLogEntry{
					ConnectionProperties: &accesslog_data.ConnectionProperties{
						SentBytes: 456,
					},
				},
				expected: `456`,
			}),
			Entry("PROTOCOL", testCase{
				field:    "PROTOCOL",
				expected: ``, // replicate Envoy's behavior
			}),
			Entry("RESPONSE_CODE", testCase{
				field:    "RESPONSE_CODE",
				expected: `0`, // replicate Envoy's behavior
			}),
			Entry("RESPONSE_CODE_DETAILS", testCase{
				field:    "RESPONSE_CODE_DETAILS",
				expected: ``, // replicate Envoy's behavior
			}),
			Entry("REQUEST_DURATION: ``", testCase{
				field:    "REQUEST_DURATION",
				expected: ``, // replicate Envoy's behavior
			}),
			Entry("REQUEST_DURATION: `57` millis", testCase{
				field: "REQUEST_DURATION",
				entry: &accesslog_data.TCPAccessLogEntry{
					CommonProperties: &accesslog_data.AccessLogCommon{
						TimeToLastRxByte: util_proto.Duration(57000 * time.Microsecond),
					},
				},
				expected: ``, // replicate Envoy's behavior
			}),
			Entry("RESPONSE_DURATION: ``", testCase{
				field:    "RESPONSE_DURATION",
				expected: ``, // replicate Envoy's behavior
			}),
			Entry("RESPONSE_DURATION: `102` millis", testCase{
				field: "RESPONSE_DURATION",
				entry: &accesslog_data.TCPAccessLogEntry{
					CommonProperties: &accesslog_data.AccessLogCommon{
						TimeToFirstUpstreamRxByte: util_proto.Duration(102000 * time.Microsecond),
					},
				},
				expected: ``, // replicate Envoy's behavior
			}),
			Entry("RESPONSE_TX_DURATION: ``", testCase{
				field:    "RESPONSE_TX_DURATION",
				expected: ``, // replicate Envoy's behavior
			}),
			Entry("RESPONSE_TX_DURATION: no TimeToFirstUpstreamRxByte", testCase{
				field: "RESPONSE_TX_DURATION",
				entry: &accesslog_data.TCPAccessLogEntry{
					CommonProperties: &accesslog_data.AccessLogCommon{
						TimeToLastDownstreamTxByte: util_proto.Duration(123000 * time.Microsecond),
					},
				},
				expected: ``, // replicate Envoy's behavior
			}),
			Entry("RESPONSE_TX_DURATION: no TimeToLastDownstreamTxByte", testCase{
				field: "RESPONSE_TX_DURATION",
				entry: &accesslog_data.TCPAccessLogEntry{
					CommonProperties: &accesslog_data.AccessLogCommon{
						TimeToFirstUpstreamRxByte: util_proto.Duration(102000 * time.Microsecond),
					},
				},
				expected: ``, // replicate Envoy's behavior
			}),
			Entry("RESPONSE_TX_DURATION: `23` millis", testCase{
				field: "RESPONSE_TX_DURATION",
				entry: &accesslog_data.TCPAccessLogEntry{
					CommonProperties: &accesslog_data.AccessLogCommon{
						TimeToFirstUpstreamRxByte:  util_proto.Duration(102000 * time.Microsecond),
						TimeToLastDownstreamTxByte: util_proto.Duration(123000 * time.Microsecond),
					},
				},
				expected: ``, // replicate Envoy's behavior
			}),
		)
	})

	Describe("FormatHttpLogEntry() and FormatTcpLogEntry()", func() {

		type testCase struct {
			field            string
			commonProperties *accesslog_data.AccessLogCommon
			expected         string
		}

		DescribeTable("should format properly",
			func(given testCase) {
				// given
				fragment := FieldOperator(given.field)

				// when
				actual, err := fragment.FormatHttpLogEntry(&accesslog_data.HTTPAccessLogEntry{
					CommonProperties: given.commonProperties,
				})
				// then
				Expect(err).ToNot(HaveOccurred())
				// and
				Expect(actual).To(Equal(given.expected))

				// when
				actual, err = fragment.FormatTcpLogEntry(&accesslog_data.TCPAccessLogEntry{
					CommonProperties: given.commonProperties,
				})
				// then
				Expect(err).ToNot(HaveOccurred())
				// and
				Expect(actual).To(Equal(given.expected))
			},
			Entry("UPSTREAM_TRANSPORT_FAILURE_REASON: ``", testCase{
				field:    "UPSTREAM_TRANSPORT_FAILURE_REASON",
				expected: ``,
			}),
			Entry("UPSTREAM_TRANSPORT_FAILURE_REASON: `mystery`", testCase{
				field: "UPSTREAM_TRANSPORT_FAILURE_REASON",
				commonProperties: &accesslog_data.AccessLogCommon{
					UpstreamTransportFailureReason: "mystery",
				},
				expected: `mystery`,
			}),
			Entry("DURATION: ``", testCase{
				field:    "DURATION",
				expected: ``,
			}),
			Entry("DURATION: `123`", testCase{
				field: "DURATION",
				commonProperties: &accesslog_data.AccessLogCommon{
					TimeToLastDownstreamTxByte: util_proto.Duration(123000 * time.Microsecond),
				},
				expected: `123`,
			}),
			Entry("RESPONSE_FLAGS: ``", testCase{
				field:    "RESPONSE_FLAGS",
				expected: ``,
			}),
			Entry("RESPONSE_FLAGS: `FailedLocalHealthcheck`", testCase{
				field: "RESPONSE_FLAGS",
				commonProperties: &accesslog_data.AccessLogCommon{
					ResponseFlags: &accesslog_data.ResponseFlags{
						FailedLocalHealthcheck: true,
					},
				},
				expected: `LH`,
			}),
			Entry("RESPONSE_FLAGS: `NoHealthyUpstream`", testCase{
				field: "RESPONSE_FLAGS",
				commonProperties: &accesslog_data.AccessLogCommon{
					ResponseFlags: &accesslog_data.ResponseFlags{
						NoHealthyUpstream: true,
					},
				},
				expected: `UH`,
			}),
			Entry("RESPONSE_FLAGS: `UpstreamRequestTimeout`", testCase{
				field: "RESPONSE_FLAGS",
				commonProperties: &accesslog_data.AccessLogCommon{
					ResponseFlags: &accesslog_data.ResponseFlags{
						UpstreamRequestTimeout: true,
					},
				},
				expected: `UT`,
			}),
			Entry("RESPONSE_FLAGS: `LocalReset`", testCase{
				field: "RESPONSE_FLAGS",
				commonProperties: &accesslog_data.AccessLogCommon{
					ResponseFlags: &accesslog_data.ResponseFlags{
						LocalReset: true,
					},
				},
				expected: `LR`,
			}),
			Entry("RESPONSE_FLAGS: `UpstreamRemoteReset`", testCase{
				field: "RESPONSE_FLAGS",
				commonProperties: &accesslog_data.AccessLogCommon{
					ResponseFlags: &accesslog_data.ResponseFlags{
						UpstreamRemoteReset: true,
					},
				},
				expected: `UR`,
			}),
			Entry("RESPONSE_FLAGS: `UpstreamConnectionFailure`", testCase{
				field: "RESPONSE_FLAGS",
				commonProperties: &accesslog_data.AccessLogCommon{
					ResponseFlags: &accesslog_data.ResponseFlags{
						UpstreamConnectionFailure: true,
					},
				},
				expected: `UF`,
			}),
			Entry("RESPONSE_FLAGS: `UpstreamConnectionTermination`", testCase{
				field: "RESPONSE_FLAGS",
				commonProperties: &accesslog_data.AccessLogCommon{
					ResponseFlags: &accesslog_data.ResponseFlags{
						UpstreamConnectionTermination: true,
					},
				},
				expected: `UC`,
			}),
			Entry("RESPONSE_FLAGS: `UpstreamOverflow`", testCase{
				field: "RESPONSE_FLAGS",
				commonProperties: &accesslog_data.AccessLogCommon{
					ResponseFlags: &accesslog_data.ResponseFlags{
						UpstreamOverflow: true,
					},
				},
				expected: `UO`,
			}),
			Entry("RESPONSE_FLAGS: `NoRouteFound`", testCase{
				field: "RESPONSE_FLAGS",
				commonProperties: &accesslog_data.AccessLogCommon{
					ResponseFlags: &accesslog_data.ResponseFlags{
						NoRouteFound: true,
					},
				},
				expected: `NR`,
			}),
			Entry("RESPONSE_FLAGS: `DelayInjected`", testCase{
				field: "RESPONSE_FLAGS",
				commonProperties: &accesslog_data.AccessLogCommon{
					ResponseFlags: &accesslog_data.ResponseFlags{
						DelayInjected: true,
					},
				},
				expected: `DI`,
			}),
			Entry("RESPONSE_FLAGS: `FaultInjected`", testCase{
				field: "RESPONSE_FLAGS",
				commonProperties: &accesslog_data.AccessLogCommon{
					ResponseFlags: &accesslog_data.ResponseFlags{
						FaultInjected: true,
					},
				},
				expected: `FI`,
			}),
			Entry("RESPONSE_FLAGS: `RateLimited`", testCase{
				field: "RESPONSE_FLAGS",
				commonProperties: &accesslog_data.AccessLogCommon{
					ResponseFlags: &accesslog_data.ResponseFlags{
						RateLimited: true,
					},
				},
				expected: `RL`,
			}),
			Entry("RESPONSE_FLAGS: `UnauthorizedDetails`", testCase{
				field: "RESPONSE_FLAGS",
				commonProperties: &accesslog_data.AccessLogCommon{
					ResponseFlags: &accesslog_data.ResponseFlags{
						UnauthorizedDetails: &accesslog_data.ResponseFlags_Unauthorized{
							Reason: accesslog_data.ResponseFlags_Unauthorized_REASON_UNSPECIFIED,
						},
					},
				},
				expected: ``,
			}),
			Entry("RESPONSE_FLAGS: `UnauthorizedExternalService`", testCase{
				field: "RESPONSE_FLAGS",
				commonProperties: &accesslog_data.AccessLogCommon{
					ResponseFlags: &accesslog_data.ResponseFlags{
						UnauthorizedDetails: &accesslog_data.ResponseFlags_Unauthorized{
							Reason: accesslog_data.ResponseFlags_Unauthorized_EXTERNAL_SERVICE,
						},
					},
				},
				expected: `UAEX`,
			}),
			Entry("RESPONSE_FLAGS: `RateLimitServiceError`", testCase{
				field: "RESPONSE_FLAGS",
				commonProperties: &accesslog_data.AccessLogCommon{
					ResponseFlags: &accesslog_data.ResponseFlags{
						RateLimitServiceError: true,
					},
				},
				expected: `RLSE`,
			}),
			Entry("RESPONSE_FLAGS: `DownstreamConnectionTermination`", testCase{
				field: "RESPONSE_FLAGS",
				commonProperties: &accesslog_data.AccessLogCommon{
					ResponseFlags: &accesslog_data.ResponseFlags{
						DownstreamConnectionTermination: true,
					},
				},
				expected: `DC`,
			}),
			Entry("RESPONSE_FLAGS: `UpstreamRetryLimitExceeded`", testCase{
				field: "RESPONSE_FLAGS",
				commonProperties: &accesslog_data.AccessLogCommon{
					ResponseFlags: &accesslog_data.ResponseFlags{
						UpstreamRetryLimitExceeded: true,
					},
				},
				expected: `URX`,
			}),
			Entry("RESPONSE_FLAGS: `StreamIdleTimeout`", testCase{
				field: "RESPONSE_FLAGS",
				commonProperties: &accesslog_data.AccessLogCommon{
					ResponseFlags: &accesslog_data.ResponseFlags{
						StreamIdleTimeout: true,
					},
				},
				expected: `SI`,
			}),
			Entry("RESPONSE_FLAGS: `InvalidEnvoyRequestHeaders`", testCase{
				field: "RESPONSE_FLAGS",
				commonProperties: &accesslog_data.AccessLogCommon{
					ResponseFlags: &accesslog_data.ResponseFlags{
						InvalidEnvoyRequestHeaders: true,
					},
				},
				expected: `IH`,
			}),
			Entry("RESPONSE_FLAGS: `DownstreamProtocolError`", testCase{
				field: "RESPONSE_FLAGS",
				commonProperties: &accesslog_data.AccessLogCommon{
					ResponseFlags: &accesslog_data.ResponseFlags{
						DownstreamProtocolError: true,
					},
				},
				expected: `DPE`,
			}),
			Entry("RESPONSE_FLAGS: all`", testCase{
				field: "RESPONSE_FLAGS",
				commonProperties: &accesslog_data.AccessLogCommon{
					ResponseFlags: &accesslog_data.ResponseFlags{
						FailedLocalHealthcheck:        true,
						NoHealthyUpstream:             true,
						UpstreamRequestTimeout:        true,
						LocalReset:                    true,
						UpstreamRemoteReset:           true,
						UpstreamConnectionFailure:     true,
						UpstreamConnectionTermination: true,
						UpstreamOverflow:              true,
						NoRouteFound:                  true,
						DelayInjected:                 true,
						FaultInjected:                 true,
						RateLimited:                   true,
						UnauthorizedDetails: &accesslog_data.ResponseFlags_Unauthorized{
							Reason: accesslog_data.ResponseFlags_Unauthorized_EXTERNAL_SERVICE,
						},
						RateLimitServiceError:           true,
						DownstreamConnectionTermination: true,
						UpstreamRetryLimitExceeded:      true,
						StreamIdleTimeout:               true,
						InvalidEnvoyRequestHeaders:      true,
						DownstreamProtocolError:         true,
					},
				},
				expected: `LH,UH,UT,LR,UR,UF,UC,UO,NR,DI,FI,RL,UAEX,RLSE,DC,URX,SI,IH,DPE`,
			}),
			Entry("UPSTREAM_HOST: ``", testCase{
				field:    "UPSTREAM_HOST",
				expected: ``,
			}),
			Entry("UPSTREAM_HOST: `outbound:backend`", testCase{
				field: "UPSTREAM_HOST",
				commonProperties: &accesslog_data.AccessLogCommon{
					UpstreamRemoteAddress: EnvoySocketAddress("10.0.0.2", 443),
				},
				expected: `10.0.0.2:443`,
			}),
			Entry("UPSTREAM_CLUSTER: ``", testCase{
				field:    "UPSTREAM_CLUSTER",
				expected: ``,
			}),
			Entry("UPSTREAM_CLUSTER: `outbound:backend`", testCase{
				field: "UPSTREAM_CLUSTER",
				commonProperties: &accesslog_data.AccessLogCommon{
					UpstreamCluster: "outbound:backend",
				},
				expected: `outbound:backend`,
			}),
			Entry("UPSTREAM_LOCAL_ADDRESS: ``", testCase{
				field:    "UPSTREAM_LOCAL_ADDRESS",
				expected: ``,
			}),
			Entry("UPSTREAM_LOCAL_ADDRESS: `127.0.0.2:10001`", testCase{
				field: "UPSTREAM_LOCAL_ADDRESS",
				commonProperties: &accesslog_data.AccessLogCommon{
					UpstreamLocalAddress: EnvoySocketAddress("127.0.0.2", 10001),
				},
				expected: `127.0.0.2:10001`,
			}),
			Entry("DOWNSTREAM_LOCAL_ADDRESS: ``", testCase{
				field:    "DOWNSTREAM_LOCAL_ADDRESS",
				expected: ``,
			}),
			Entry("DOWNSTREAM_LOCAL_ADDRESS: `127.0.0.1:10000`", testCase{
				field: "DOWNSTREAM_LOCAL_ADDRESS",
				commonProperties: &accesslog_data.AccessLogCommon{
					DownstreamLocalAddress: EnvoySocketAddress("127.0.0.1", 10000),
				},
				expected: `127.0.0.1:10000`,
			}),
			Entry("DOWNSTREAM_LOCAL_ADDRESS_WITHOUT_PORT: ``", testCase{
				field:    "DOWNSTREAM_LOCAL_ADDRESS_WITHOUT_PORT",
				expected: ``,
			}),
			Entry("DOWNSTREAM_LOCAL_ADDRESS_WITHOUT_PORT: `127.0.0.1`", testCase{
				field: "DOWNSTREAM_LOCAL_ADDRESS_WITHOUT_PORT",
				commonProperties: &accesslog_data.AccessLogCommon{
					DownstreamLocalAddress: EnvoySocketAddress("127.0.0.1", 10000),
				},
				expected: `127.0.0.1`,
			}),
			Entry("DOWNSTREAM_REMOTE_ADDRESS: ``", testCase{
				field:    "DOWNSTREAM_REMOTE_ADDRESS",
				expected: ``,
			}),
			Entry("DOWNSTREAM_REMOTE_ADDRESS: `127.0.0.3:53165`", testCase{
				field: "DOWNSTREAM_REMOTE_ADDRESS",
				commonProperties: &accesslog_data.AccessLogCommon{
					DownstreamRemoteAddress: EnvoySocketAddress("127.0.0.3", 53165),
				},
				expected: `127.0.0.3:53165`,
			}),
			Entry("DOWNSTREAM_REMOTE_ADDRESS_WITHOUT_PORT: ``", testCase{
				field:    "DOWNSTREAM_REMOTE_ADDRESS_WITHOUT_PORT",
				expected: ``,
			}),
			Entry("DOWNSTREAM_REMOTE_ADDRESS_WITHOUT_PORT: `127.0.0.3`", testCase{
				field: "DOWNSTREAM_REMOTE_ADDRESS_WITHOUT_PORT",
				commonProperties: &accesslog_data.AccessLogCommon{
					DownstreamRemoteAddress: EnvoySocketAddress("127.0.0.3", 53165),
				},
				expected: `127.0.0.3`,
			}),
			Entry("DOWNSTREAM_DIRECT_REMOTE_ADDRESS: ``", testCase{
				field:    "DOWNSTREAM_DIRECT_REMOTE_ADDRESS",
				expected: ``,
			}),
			Entry("DOWNSTREAM_DIRECT_REMOTE_ADDRESS: `127.0.0.1:53166`", testCase{
				field: "DOWNSTREAM_DIRECT_REMOTE_ADDRESS",
				commonProperties: &accesslog_data.AccessLogCommon{
					DownstreamDirectRemoteAddress: EnvoySocketAddress("127.0.0.1", 53166),
				},
				expected: `127.0.0.1:53166`,
			}),
			Entry("DOWNSTREAM_DIRECT_REMOTE_ADDRESS_WITHOUT_PORT: ``", testCase{
				field:    "DOWNSTREAM_DIRECT_REMOTE_ADDRESS_WITHOUT_PORT",
				expected: ``,
			}),
			Entry("DOWNSTREAM_DIRECT_REMOTE_ADDRESS_WITHOUT_PORT: `127.0.0.4`", testCase{
				field: "DOWNSTREAM_DIRECT_REMOTE_ADDRESS_WITHOUT_PORT",
				commonProperties: &accesslog_data.AccessLogCommon{
					DownstreamDirectRemoteAddress: EnvoySocketAddress("127.0.0.4", 53166),
				},
				expected: `127.0.0.4`,
			}),
			Entry("REQUESTED_SERVER_NAME: ``", testCase{
				field:    "REQUESTED_SERVER_NAME",
				expected: ``,
			}),
			Entry("REQUESTED_SERVER_NAME: `backend.internal`", testCase{
				field: "REQUESTED_SERVER_NAME",
				commonProperties: &accesslog_data.AccessLogCommon{
					TlsProperties: &accesslog_data.TLSProperties{
						TlsSniHostname: "backend.internal",
					},
				},
				expected: `backend.internal`,
			}),
			Entry("ROUTE_NAME: ``", testCase{
				field:    "ROUTE_NAME",
				expected: ``,
			}),
			Entry("ROUTE_NAME: `outbound:backend`", testCase{
				field: "ROUTE_NAME",
				commonProperties: &accesslog_data.AccessLogCommon{
					RouteName: "outbound:backend",
				},
				expected: `outbound:backend`,
			}),
			Entry("DOWNSTREAM_PEER_URI_SAN: ``", testCase{
				field:    "DOWNSTREAM_PEER_URI_SAN",
				expected: ``,
			}),
			Entry("DOWNSTREAM_PEER_URI_SAN: `spiffe://default/web,spiffe://default/web-admin`", testCase{
				field: "DOWNSTREAM_PEER_URI_SAN",
				commonProperties: &accesslog_data.AccessLogCommon{
					TlsProperties: &accesslog_data.TLSProperties{
						PeerCertificateProperties: &accesslog_data.TLSProperties_CertificateProperties{
							SubjectAltName: []*accesslog_data.TLSProperties_CertificateProperties_SubjectAltName{
								{
									San: &accesslog_data.TLSProperties_CertificateProperties_SubjectAltName_Uri{
										Uri: "spiffe://default/web",
									},
								},
								{
									San: &accesslog_data.TLSProperties_CertificateProperties_SubjectAltName_Dns{
										Dns: "web.internal",
									},
								},
								{
									San: &accesslog_data.TLSProperties_CertificateProperties_SubjectAltName_Uri{
										Uri: "spiffe://default/web-admin",
									},
								},
							},
						},
					},
				},
				expected: `spiffe://default/web,spiffe://default/web-admin`,
			}),
			Entry("DOWNSTREAM_LOCAL_URI_SAN: ``", testCase{
				field:    "DOWNSTREAM_LOCAL_URI_SAN",
				expected: ``,
			}),
			Entry("DOWNSTREAM_LOCAL_URI_SAN: `spiffe://default/backend,spiffe://default/backend-admin`", testCase{
				field: "DOWNSTREAM_LOCAL_URI_SAN",
				commonProperties: &accesslog_data.AccessLogCommon{
					TlsProperties: &accesslog_data.TLSProperties{
						LocalCertificateProperties: &accesslog_data.TLSProperties_CertificateProperties{
							SubjectAltName: []*accesslog_data.TLSProperties_CertificateProperties_SubjectAltName{
								{
									San: &accesslog_data.TLSProperties_CertificateProperties_SubjectAltName_Uri{
										Uri: "spiffe://default/backend",
									},
								},
								{
									San: &accesslog_data.TLSProperties_CertificateProperties_SubjectAltName_Dns{
										Dns: "backend.internal",
									},
								},
								{
									San: &accesslog_data.TLSProperties_CertificateProperties_SubjectAltName_Uri{
										Uri: "spiffe://default/backend-admin",
									},
								},
							},
						},
					},
				},
				expected: `spiffe://default/backend,spiffe://default/backend-admin`,
			}),
			Entry("DOWNSTREAM_PEER_SUBJECT: ``", testCase{
				field:    "DOWNSTREAM_PEER_SUBJECT",
				expected: ``,
			}),
			Entry("DOWNSTREAM_PEER_SUBJECT: `CN=web,OU=IT,O=Webshop,L=San Francisco,ST=California,C=US`", testCase{
				field: "DOWNSTREAM_PEER_SUBJECT",
				commonProperties: &accesslog_data.AccessLogCommon{
					TlsProperties: &accesslog_data.TLSProperties{
						PeerCertificateProperties: &accesslog_data.TLSProperties_CertificateProperties{
							Subject: `CN=web,OU=IT,O=Webshop,L=San Francisco,ST=California,C=US`,
						},
					},
				},
				expected: `CN=web,OU=IT,O=Webshop,L=San Francisco,ST=California,C=US`,
			}),
			Entry("DOWNSTREAM_LOCAL_SUBJECT: ``", testCase{
				field:    "DOWNSTREAM_LOCAL_SUBJECT",
				expected: ``,
			}),
			Entry("DOWNSTREAM_LOCAL_SUBJECT: `CN=backend,OU=IT,O=Webshop,L=San Francisco,ST=California,C=US`", testCase{
				field: "DOWNSTREAM_LOCAL_SUBJECT",
				commonProperties: &accesslog_data.AccessLogCommon{
					TlsProperties: &accesslog_data.TLSProperties{
						LocalCertificateProperties: &accesslog_data.TLSProperties_CertificateProperties{
							Subject: `CN=backend,OU=IT,O=Webshop,L=San Francisco,ST=California,C=US`,
						},
					},
				},
				expected: `CN=backend,OU=IT,O=Webshop,L=San Francisco,ST=California,C=US`,
			}),
			Entry("DOWNSTREAM_TLS_SESSION_ID: ``", testCase{
				field:    "DOWNSTREAM_TLS_SESSION_ID",
				expected: ``,
			}),
			Entry("DOWNSTREAM_TLS_SESSION_ID: `b10662bf6bd4e6a068f0910d3d60c50f000355840fea4ce6844626b61c973901`", testCase{
				field: "DOWNSTREAM_TLS_SESSION_ID",
				commonProperties: &accesslog_data.AccessLogCommon{
					TlsProperties: &accesslog_data.TLSProperties{
						TlsSessionId: "b10662bf6bd4e6a068f0910d3d60c50f000355840fea4ce6844626b61c973901",
					},
				},
				expected: `b10662bf6bd4e6a068f0910d3d60c50f000355840fea4ce6844626b61c973901`,
			}),
			Entry("DOWNSTREAM_TLS_CIPHER: ``", testCase{
				field:    "DOWNSTREAM_TLS_CIPHER",
				expected: ``,
			}),
			Entry("DOWNSTREAM_TLS_CIPHER: `ECDHE-RSA-CHACHA20-POLY1305`", testCase{
				field: "DOWNSTREAM_TLS_CIPHER",
				commonProperties: &accesslog_data.AccessLogCommon{
					TlsProperties: &accesslog_data.TLSProperties{
						TlsCipherSuite: util_proto.UInt32(
							uint32(TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305)),
					},
				},
				expected: `ECDHE-RSA-CHACHA20-POLY1305`,
			}),
			Entry("DOWNSTREAM_TLS_VERSION: ``", testCase{
				field:    "DOWNSTREAM_TLS_VERSION",
				expected: ``,
			}),
			Entry("DOWNSTREAM_TLS_VERSION: `TLSv1`", testCase{
				field: "DOWNSTREAM_TLS_VERSION",
				commonProperties: &accesslog_data.AccessLogCommon{
					TlsProperties: &accesslog_data.TLSProperties{
						TlsVersion: accesslog_data.TLSProperties_TLSv1,
					},
				},
				expected: `TLSv1`,
			}),
			Entry("DOWNSTREAM_TLS_VERSION: `TLSv1.1`", testCase{
				field: "DOWNSTREAM_TLS_VERSION",
				commonProperties: &accesslog_data.AccessLogCommon{
					TlsProperties: &accesslog_data.TLSProperties{
						TlsVersion: accesslog_data.TLSProperties_TLSv1_1,
					},
				},
				expected: `TLSv1.1`,
			}),
			Entry("DOWNSTREAM_TLS_VERSION: `TLSv1.2`", testCase{
				field: "DOWNSTREAM_TLS_VERSION",
				commonProperties: &accesslog_data.AccessLogCommon{
					TlsProperties: &accesslog_data.TLSProperties{
						TlsVersion: accesslog_data.TLSProperties_TLSv1_2,
					},
				},
				expected: `TLSv1.2`,
			}),
			Entry("DOWNSTREAM_TLS_VERSION: `TLSv1.3`", testCase{
				field: "DOWNSTREAM_TLS_VERSION",
				commonProperties: &accesslog_data.AccessLogCommon{
					TlsProperties: &accesslog_data.TLSProperties{
						TlsVersion: accesslog_data.TLSProperties_TLSv1_3,
					},
				},
				expected: `TLSv1.3`,
			}),
			Entry("DOWNSTREAM_PEER_FINGERPRINT_256", testCase{
				field:    "DOWNSTREAM_PEER_FINGERPRINT_256",
				expected: `UNSUPPORTED_COMMAND(%DOWNSTREAM_PEER_FINGERPRINT_256%)`,
			}),
			Entry("DOWNSTREAM_PEER_SERIAL", testCase{
				field:    "DOWNSTREAM_PEER_SERIAL",
				expected: `UNSUPPORTED_COMMAND(%DOWNSTREAM_PEER_SERIAL%)`,
			}),
			Entry("DOWNSTREAM_PEER_ISSUER", testCase{
				field:    "DOWNSTREAM_PEER_ISSUER",
				expected: `UNSUPPORTED_COMMAND(%DOWNSTREAM_PEER_ISSUER%)`,
			}),
			Entry("DOWNSTREAM_PEER_CERT", testCase{
				field:    "DOWNSTREAM_PEER_CERT",
				expected: `UNSUPPORTED_COMMAND(%DOWNSTREAM_PEER_CERT%)`,
			}),
			Entry("DOWNSTREAM_PEER_CERT_V_START", testCase{
				field:    "DOWNSTREAM_PEER_CERT_V_START",
				expected: `UNSUPPORTED_COMMAND(%DOWNSTREAM_PEER_CERT_V_START%)`,
			}),
			Entry("DOWNSTREAM_PEER_CERT_V_END", testCase{
				field:    "DOWNSTREAM_PEER_CERT_V_END",
				expected: `UNSUPPORTED_COMMAND(%DOWNSTREAM_PEER_CERT_V_END%)`,
			}),
			Entry("HOSTNAME", testCase{
				field:    "HOSTNAME",
				expected: `UNSUPPORTED_COMMAND(%HOSTNAME%)`,
			}),
		)
	})

	Describe("String()", func() {
		type testCase struct {
			field    string
			expected string
		}

		DescribeTable("should return correct canonical representation",
			func(given testCase) {
				// setup
				fragment := FieldOperator(given.field)

				// when
				actual := fragment.String()
				// then
				Expect(actual).To(Equal(given.expected))

			},
			Entry("%BYTES_RECEIVED%", testCase{
				field:    "BYTES_RECEIVED",
				expected: `%BYTES_RECEIVED%`,
			}),
		)
	})
})
