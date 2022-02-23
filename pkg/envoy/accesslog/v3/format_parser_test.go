package v3_test

import (
	"time"

	accesslog_data "github.com/envoyproxy/go-control-plane/envoy/data/accesslog/v3"
	accesslog_config "github.com/envoyproxy/go-control-plane/envoy/extensions/access_loggers/grpc/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/pkg/envoy/accesslog/v3"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

var _ = Describe("ParseFormat()", func() {

	Context("valid format string", func() {

		commonProperties := &accesslog_data.AccessLogCommon{
			StartTime:                  util_proto.MustTimestampProto(time.Unix(1582062737, 987654321)),
			TimeToLastRxByte:           util_proto.Duration(57000 * time.Microsecond),
			TimeToFirstUpstreamRxByte:  util_proto.Duration(102000 * time.Microsecond),
			TimeToLastDownstreamTxByte: util_proto.Duration(123000 * time.Microsecond),
			ResponseFlags: &accesslog_data.ResponseFlags{
				UpstreamConnectionFailure:  true,
				UpstreamRetryLimitExceeded: true,
			},
			DownstreamLocalAddress:         EnvoySocketAddress("127.0.0.1", 10000),
			DownstreamRemoteAddress:        EnvoySocketAddress("127.0.0.3", 53165),
			DownstreamDirectRemoteAddress:  EnvoySocketAddress("127.0.0.4", 53166),
			UpstreamCluster:                "outbound:backend",
			UpstreamLocalAddress:           EnvoySocketAddress("127.0.0.2", 10001),
			UpstreamRemoteAddress:          EnvoySocketAddress("10.0.0.2", 443),
			UpstreamTransportFailureReason: "mystery",
			RouteName:                      "outbound:backend",
			TlsProperties: &accesslog_data.TLSProperties{
				TlsSniHostname: "backend.internal",
				PeerCertificateProperties: &accesslog_data.TLSProperties_CertificateProperties{
					Subject: `CN=web,OU=IT,O=Webshop,L=San Francisco,ST=California,C=US`,
					SubjectAltName: []*accesslog_data.TLSProperties_CertificateProperties_SubjectAltName{
						{
							San: &accesslog_data.TLSProperties_CertificateProperties_SubjectAltName_Uri{
								Uri: "spiffe://default/web",
							},
						},
					},
				},
				LocalCertificateProperties: &accesslog_data.TLSProperties_CertificateProperties{
					Subject: `CN=backend,OU=IT,O=Webshop,L=San Francisco,ST=California,C=US`,
					SubjectAltName: []*accesslog_data.TLSProperties_CertificateProperties_SubjectAltName{
						{
							San: &accesslog_data.TLSProperties_CertificateProperties_SubjectAltName_Uri{
								Uri: "spiffe://default/backend",
							},
						},
					},
				},
				TlsSessionId: "b10662bf6bd4e6a068f0910d3d60c50f000355840fea4ce6844626b61c973901",
				TlsVersion:   accesslog_data.TLSProperties_TLSv1_2,
				TlsCipherSuite: util_proto.UInt32(
					uint32(TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305)),
			},
		}

		httpExample := &accesslog_data.HTTPAccessLogEntry{
			CommonProperties: commonProperties,
			ProtocolVersion:  accesslog_data.HTTPAccessLogEntry_HTTP11,
			Request: &accesslog_data.HTTPRequestProperties{
				Scheme:           "https",
				Authority:        "backend.internal:8080",
				Path:             "/api",
				RequestBodyBytes: 234,
			},
			Response: &accesslog_data.HTTPResponseProperties{
				ResponseCode:        util_proto.UInt32(200),
				ResponseCodeDetails: "response code details",
				ResponseHeaders: map[string]string{
					"server":       "Tomcat",
					"content-type": "application/json",
				},
				ResponseTrailers: map[string]string{
					"grpc-status":  "14",
					"grpc-message": "unavailable",
				},
				ResponseBodyBytes: 567,
			},
		}

		tcpExample := &accesslog_data.TCPAccessLogEntry{
			CommonProperties: commonProperties,
			ConnectionProperties: &accesslog_data.ConnectionProperties{
				ReceivedBytes: 234,
				SentBytes:     567,
			},
		}

		type testCase struct {
			format       string
			expectedHTTP string
			expectedTCP  string
		}

		DescribeTable("should succefully parse valid format string",
			func(given testCase) {
				// when
				format, err := ParseFormat(given.format)
				// then
				Expect(err).ToNot(HaveOccurred())

				// when
				actual, err := format.FormatHttpLogEntry(httpExample)
				// then
				Expect(err).ToNot(HaveOccurred())
				// and
				Expect(actual).To(Equal(given.expectedHTTP))

				// when
				actual, err = format.FormatTcpLogEntry(tcpExample)
				// then
				Expect(err).ToNot(HaveOccurred())
				// and
				Expect(actual).To(Equal(given.expectedTCP))
			},
			Entry("empty string", testCase{
				format:       "",
				expectedHTTP: ``,
				expectedTCP:  ``,
			}),
			Entry("literal string", testCase{
				format:       `text without Envoy command operators`,
				expectedHTTP: `text without Envoy command operators`,
				expectedTCP:  `text without Envoy command operators`,
			}),
			Entry("%START_TIME%", testCase{
				format:       `%START_TIME%`,
				expectedHTTP: `2020-02-18T21:52:17.987Z`,
				expectedTCP:  `2020-02-18T21:52:17.987Z`,
			}),
			Entry("%START_TIME()%", testCase{
				format:       `%START_TIME%`,
				expectedHTTP: `2020-02-18T21:52:17.987Z`,
				expectedTCP:  `2020-02-18T21:52:17.987Z`,
			}),
			Entry("%START_TIME(%Y/%m/%dT%H:%M:%S%z %s)%", testCase{
				format:       `%START_TIME(%Y/%m/%dT%H:%M:%S%z %s)%`, // user-defined format is not supported yet
				expectedHTTP: `2020-02-18T21:52:17.987Z`,
				expectedTCP:  `2020-02-18T21:52:17.987Z`,
			}),
			Entry("%START_TIME(%s.%3f)%", testCase{
				format:       `%START_TIME(%s.%3f)%`, // user-defined format is not supported yet
				expectedHTTP: `2020-02-18T21:52:17.987Z`,
				expectedTCP:  `2020-02-18T21:52:17.987Z`,
			}),
			Entry("%BYTES_RECEIVED%", testCase{
				format:       `%BYTES_RECEIVED%`,
				expectedHTTP: `234`,
				expectedTCP:  `234`,
			}),
			Entry("%PROTOCOL%", testCase{
				format:       `%PROTOCOL%`,
				expectedHTTP: `HTTP/1.1`,
				expectedTCP:  `-`, // replicate Envoy's behavior
			}),
			Entry("%RESPONSE_CODE%", testCase{
				format:       `%RESPONSE_CODE%`,
				expectedHTTP: `200`,
				expectedTCP:  `0`, // replicate Envoy's behavior
			}),
			Entry("%RESPONSE_CODE_DETAILS%", testCase{
				format:       `%RESPONSE_CODE_DETAILS%`,
				expectedHTTP: `response code details`,
				expectedTCP:  `-`, // replicate Envoy's behavior
			}),
			Entry("%BYTES_SENT%", testCase{
				format:       `%BYTES_SENT%`,
				expectedHTTP: `567`,
				expectedTCP:  `567`,
			}),
			Entry("%REQUEST_DURATION%", testCase{
				format:       `%REQUEST_DURATION%`,
				expectedHTTP: `57`, // time in millis
				expectedTCP:  `-`,  // replicate Envoy's behavior
			}),
			Entry("%RESPONSE_DURATION%", testCase{
				format:       `%RESPONSE_DURATION%`,
				expectedHTTP: `102`, // time in millis
				expectedTCP:  `-`,   // replicate Envoy's behavior
			}),
			Entry("%RESPONSE_TX_DURATION%", testCase{
				format:       `%RESPONSE_TX_DURATION%`,
				expectedHTTP: `21`, // time in millis
				expectedTCP:  `-`,  // replicate Envoy's behavior
			}),
			Entry("%UPSTREAM_TRANSPORT_FAILURE_REASON%", testCase{
				format:       `%UPSTREAM_TRANSPORT_FAILURE_REASON%`,
				expectedHTTP: `mystery`,
				expectedTCP:  `mystery`,
			}),
			Entry("%DURATION%", testCase{
				format:       `%DURATION%`,
				expectedHTTP: `123`,
				expectedTCP:  `123`,
			}),
			Entry("%RESPONSE_FLAGS%", testCase{
				format:       `%RESPONSE_FLAGS%`,
				expectedHTTP: `UF,URX`,
				expectedTCP:  `UF,URX`,
			}),
			Entry("%REQ()%", testCase{ // apparently, Envoy allows both `Header` and `AltHeader` to be empty
				format:       `%REQ()%`,
				expectedHTTP: `-`, // replicate Envoy's behavior
				expectedTCP:  `-`, // replicate Envoy's behavior
			}),
			Entry("%REQ():10%", testCase{ // apparently, Envoy allows both `Header` and `AltHeader` to be empty
				format:       `%REQ():10%`,
				expectedHTTP: `-`, // replicate Envoy's behavior
				expectedTCP:  `-`, // replicate Envoy's behavior
			}),
			Entry("%REQ(:authority)%", testCase{ // pseudo header
				format:       `%REQ(:authority)%`,
				expectedHTTP: `backend.internal:8080`,
				expectedTCP:  `-`, // replicate Envoy's behavior
			}),
			Entry("%REQ(:authority):7%", testCase{ // max length
				format:       `%REQ(:authority):7%`,
				expectedHTTP: `backend`,
				expectedTCP:  `-`, // replicate Envoy's behavior
			}),
			Entry("%REQ(x-missing-header?:authority)%", testCase{ // altHeader
				format:       `%REQ(x-missing-header?:authority)%`,
				expectedHTTP: `backend.internal:8080`,
				expectedTCP:  `-`, // replicate Envoy's behavior
			}),
			Entry("%REQ(x-missing-header?:authority):7%", testCase{ // altHeader w/ maxLen
				format:       `%REQ(x-missing-header?:authority):7%`,
				expectedHTTP: `backend`,
				expectedTCP:  `-`, // replicate Envoy's behavior
			}),
			Entry("%REQ(x-missing-header?:AUTHORITY):7%", testCase{ // uppercase altHeader w/ maxLen
				format:       `%REQ(x-missing-header?:AUTHORITY):7%`,
				expectedHTTP: `backend`,
				expectedTCP:  `-`, // replicate Envoy's behavior
			}),
			Entry("%REQ(:authority?:path)%", testCase{ // header over altHeader
				format:       `%REQ(:authority?:path)%`,
				expectedHTTP: `backend.internal:8080`,
				expectedTCP:  `-`, // replicate Envoy's behavior
			}),
			Entry("%REQ(:authority?:path):7%", testCase{ // header over altHeader w/ maxLen
				format:       `%REQ(:authority?:path):7%`,
				expectedHTTP: `backend`,
				expectedTCP:  `-`, // replicate Envoy's behavior
			}),
			Entry("%REQ(:AUTHORITY?:path):7%", testCase{ // uppercase header over altHeader w/ maxLen
				format:       `%REQ(:AUTHORITY?:path):7%`,
				expectedHTTP: `backend`,
				expectedTCP:  `-`, // replicate Envoy's behavior
			}),
			Entry("%RESP()%", testCase{ // apparently, Envoy allows both `Header` and `AltHeader` to be empty
				format:       `%RESP()%`,
				expectedHTTP: `-`, // replicate Envoy's behavior
				expectedTCP:  `-`, // replicate Envoy's behavior
			}),
			Entry("%RESP():10%", testCase{ // apparently, Envoy allows both `Header` and `AltHeader` to be empty
				format:       `%RESP():10%`,
				expectedHTTP: `-`, // replicate Envoy's behavior
				expectedTCP:  `-`, // replicate Envoy's behavior
			}),
			Entry("%RESP(server)%", testCase{ // pseudo header
				format:       `%RESP(server)%`,
				expectedHTTP: `Tomcat`,
				expectedTCP:  `-`, // replicate Envoy's behavior
			}),
			Entry("%RESP(server):3%", testCase{ // max length
				format:       `%RESP(server):3%`,
				expectedHTTP: `Tom`,
				expectedTCP:  `-`, // replicate Envoy's behavior
			}),
			Entry("%RESP(x-missing-header?server)%", testCase{ // altHeader
				format:       `%RESP(x-missing-header?server)%`,
				expectedHTTP: `Tomcat`,
				expectedTCP:  `-`, // replicate Envoy's behavior
			}),
			Entry("%RESP(x-missing-header?server):3%", testCase{ // altHeader w/ maxLen
				format:       `%RESP(x-missing-header?server):3%`,
				expectedHTTP: `Tom`,
				expectedTCP:  `-`, // replicate Envoy's behavior
			}),
			Entry("%RESP(x-missing-header?SERVER):3%", testCase{ // uppercase altHeader w/ maxLen
				format:       `%RESP(x-missing-header?SERVER):3%`,
				expectedHTTP: `Tom`,
				expectedTCP:  `-`, // replicate Envoy's behavior
			}),
			Entry("%RESP(server?content-type)%", testCase{ // header over altHeader
				format:       `%RESP(server?:content-type)%`,
				expectedHTTP: `Tomcat`,
				expectedTCP:  `-`, // replicate Envoy's behavior
			}),
			Entry("%RESP(server?content-type):3%", testCase{ // header over altHeader w/ maxLen
				format:       `%RESP(server?content-type):3%`,
				expectedHTTP: `Tom`,
				expectedTCP:  `-`, // replicate Envoy's behavior
			}),
			Entry("%RESP(SERVER?content-type):3%", testCase{ // uppercase header over altHeader w/ maxLen
				format:       `%RESP(SERVER?content-type):3%`,
				expectedHTTP: `Tom`,
				expectedTCP:  `-`, // replicate Envoy's behavior
			}),
			Entry("%TRAILER()%", testCase{ // apparently, Envoy allows both `Header` and `AltHeader` to be empty
				format:       `%TRAILER()%`,
				expectedHTTP: `-`, // replicate Envoy's behavior
				expectedTCP:  `-`, // replicate Envoy's behavior
			}),
			Entry("%TRAILER():10%", testCase{ // apparently, Envoy allows both `Header` and `AltHeader` to be empty
				format:       `%TRAILER():10%`,
				expectedHTTP: `-`, // replicate Envoy's behavior
				expectedTCP:  `-`, // replicate Envoy's behavior
			}),
			Entry("%TRAILER(grpc-status)%", testCase{ // pseudo header
				format:       `%TRAILER(grpc-status)%`,
				expectedHTTP: `14`,
				expectedTCP:  `-`, // replicate Envoy's behavior
			}),
			Entry("%TRAILER(grpc-status):1%", testCase{ // max length
				format:       `%TRAILER(grpc-status):1%`,
				expectedHTTP: `1`,
				expectedTCP:  `-`, // replicate Envoy's behavior
			}),
			Entry("%TRAILER(x-missing-header?grpc-status)%", testCase{ // altHeader
				format:       `%TRAILER(x-missing-header?grpc-status)%`,
				expectedHTTP: `14`,
				expectedTCP:  `-`, // replicate Envoy's behavior
			}),
			Entry("%TRAILER(x-missing-header?grpc-status):1%", testCase{ // altHeader w/ maxLen
				format:       `%TRAILER(x-missing-header?grpc-status):1%`,
				expectedHTTP: `1`,
				expectedTCP:  `-`, // replicate Envoy's behavior
			}),
			Entry("%TRAILER(x-missing-header?GRPC-STATUS):1%", testCase{ // uppercase altHeader w/ maxLen
				format:       `%TRAILER(x-missing-header?GRPC-STATUS):1%`,
				expectedHTTP: `1`,
				expectedTCP:  `-`, // replicate Envoy's behavior
			}),
			Entry("%TRAILER(grpc-status?grpc-message)%", testCase{ // header over altHeader
				format:       `%TRAILER(grpc-status?:grpc-message)%`,
				expectedHTTP: `14`,
				expectedTCP:  `-`, // replicate Envoy's behavior
			}),
			Entry("%TRAILER(grpc-status?grpc-message):1%", testCase{ // header over altHeader w/ maxLen
				format:       `%TRAILER(grpc-status?grpc-message):1%`,
				expectedHTTP: `1`,
				expectedTCP:  `-`, // replicate Envoy's behavior
			}),
			Entry("%TRAILER(GRPC-STATUS?grpc-message):1%", testCase{ // uppercase header over altHeader w/ maxLen
				format:       `%TRAILER(GRPC-STATUS?grpc-message):1%`,
				expectedHTTP: `1`,
				expectedTCP:  `-`, // replicate Envoy's behavior
			}),
			Entry("%DYNAMIC_METADATA()%", testCase{ // apparently, Envoy allows both `FilterNamespace` and `Path` to be empty
				format:       `%DYNAMIC_METADATA()%`,
				expectedHTTP: `UNSUPPORTED_COMMAND(%DYNAMIC_METADATA(NAMESPACE:KEY*):Z%)`, // not supported yet
				expectedTCP:  `UNSUPPORTED_COMMAND(%DYNAMIC_METADATA(NAMESPACE:KEY*):Z%)`, // not supported yet
			}),
			Entry("%DYNAMIC_METADATA():10%", testCase{ // apparently, Envoy allows both `FilterNamespace` and `Path` to be empty
				format:       `%DYNAMIC_METADATA():10%`,
				expectedHTTP: `UNSUPPORTED_COMMAND(%DYNAMIC_METADATA(NAMESPACE:KEY*):Z%)`, // not supported yet
				expectedTCP:  `UNSUPPORTED_COMMAND(%DYNAMIC_METADATA(NAMESPACE:KEY*):Z%)`, // not supported yet
			}),
			Entry("%DYNAMIC_METADATA(com.test.my_filter)%", testCase{
				format:       `%DYNAMIC_METADATA(com.test.my_filter)%`,
				expectedHTTP: `UNSUPPORTED_COMMAND(%DYNAMIC_METADATA(NAMESPACE:KEY*):Z%)`, // not supported yet
				expectedTCP:  `UNSUPPORTED_COMMAND(%DYNAMIC_METADATA(NAMESPACE:KEY*):Z%)`, // not supported yet
			}),
			Entry("%DYNAMIC_METADATA(com.test.my_filter):10%", testCase{
				format:       `%DYNAMIC_METADATA(com.test.my_filter):10%`,
				expectedHTTP: `UNSUPPORTED_COMMAND(%DYNAMIC_METADATA(NAMESPACE:KEY*):Z%)`, // not supported yet
				expectedTCP:  `UNSUPPORTED_COMMAND(%DYNAMIC_METADATA(NAMESPACE:KEY*):Z%)`, // not supported yet
			}),
			Entry("%DYNAMIC_METADATA(com.test.my_filter:test_key)%", testCase{
				format:       `%DYNAMIC_METADATA(com.test.my_filter:test_key)%`,
				expectedHTTP: `UNSUPPORTED_COMMAND(%DYNAMIC_METADATA(NAMESPACE:KEY*):Z%)`, // not supported yet
				expectedTCP:  `UNSUPPORTED_COMMAND(%DYNAMIC_METADATA(NAMESPACE:KEY*):Z%)`, // not supported yet
			}),
			Entry("%DYNAMIC_METADATA(com.test.my_filter:test_key):10%", testCase{
				format:       `%DYNAMIC_METADATA(com.test.my_filter:test_key):10%`,
				expectedHTTP: `UNSUPPORTED_COMMAND(%DYNAMIC_METADATA(NAMESPACE:KEY*):Z%)`, // not supported yet
				expectedTCP:  `UNSUPPORTED_COMMAND(%DYNAMIC_METADATA(NAMESPACE:KEY*):Z%)`, // not supported yet
			}),
			Entry("%DYNAMIC_METADATA(com.test.my_filter:test_object:inner_key)%", testCase{
				format:       `%DYNAMIC_METADATA(com.test.my_filter:test_object:inner_key)%`,
				expectedHTTP: `UNSUPPORTED_COMMAND(%DYNAMIC_METADATA(NAMESPACE:KEY*):Z%)`, // not supported yet
				expectedTCP:  `UNSUPPORTED_COMMAND(%DYNAMIC_METADATA(NAMESPACE:KEY*):Z%)`, // not supported yet
			}),
			Entry("%DYNAMIC_METADATA(com.test.my_filter:test_object:inner_key):10%", testCase{
				format:       `%DYNAMIC_METADATA(com.test.my_filter:test_object:inner_key):10%`,
				expectedHTTP: `UNSUPPORTED_COMMAND(%DYNAMIC_METADATA(NAMESPACE:KEY*):Z%)`, // not supported yet
				expectedTCP:  `UNSUPPORTED_COMMAND(%DYNAMIC_METADATA(NAMESPACE:KEY*):Z%)`, // not supported yet
			}),
			Entry("%FILTER_STATE(key)%", testCase{
				format:       `%FILTER_STATE(key)%`,
				expectedHTTP: `UNSUPPORTED_COMMAND(%FILTER_STATE(KEY):Z%)`, // not supported yet
				expectedTCP:  `UNSUPPORTED_COMMAND(%FILTER_STATE(KEY):Z%)`, // not supported yet
			}),
			Entry("%FILTER_STATE(key):10%", testCase{
				format:       `%FILTER_STATE(key):10%`,
				expectedHTTP: `UNSUPPORTED_COMMAND(%FILTER_STATE(KEY):Z%)`, // not supported yet
				expectedTCP:  `UNSUPPORTED_COMMAND(%FILTER_STATE(KEY):Z%)`, // not supported yet
			}),
			Entry("%UPSTREAM_HOST%", testCase{
				format:       `%UPSTREAM_HOST%`,
				expectedHTTP: `10.0.0.2:443`,
				expectedTCP:  `10.0.0.2:443`,
			}),
			Entry("%UPSTREAM_CLUSTER%", testCase{
				format:       `%UPSTREAM_CLUSTER%`,
				expectedHTTP: `outbound:backend`,
				expectedTCP:  `outbound:backend`,
			}),
			Entry("%UPSTREAM_LOCAL_ADDRESS%", testCase{
				format:       `%UPSTREAM_LOCAL_ADDRESS%`,
				expectedHTTP: `127.0.0.2:10001`,
				expectedTCP:  `127.0.0.2:10001`,
			}),
			Entry("%DOWNSTREAM_LOCAL_ADDRESS%", testCase{
				format:       `%DOWNSTREAM_LOCAL_ADDRESS%`,
				expectedHTTP: `127.0.0.1:10000`,
				expectedTCP:  `127.0.0.1:10000`,
			}),
			Entry("%DOWNSTREAM_LOCAL_ADDRESS_WITHOUT_PORT%", testCase{
				format:       `%DOWNSTREAM_LOCAL_ADDRESS_WITHOUT_PORT%`,
				expectedHTTP: `127.0.0.1`,
				expectedTCP:  `127.0.0.1`,
			}),
			Entry("%DOWNSTREAM_REMOTE_ADDRESS%", testCase{
				format:       `%DOWNSTREAM_REMOTE_ADDRESS%`,
				expectedHTTP: `127.0.0.3:53165`,
				expectedTCP:  `127.0.0.3:53165`,
			}),
			Entry("%DOWNSTREAM_REMOTE_ADDRESS_WITHOUT_PORT%", testCase{
				format:       `%DOWNSTREAM_REMOTE_ADDRESS_WITHOUT_PORT%`,
				expectedHTTP: `127.0.0.3`,
				expectedTCP:  `127.0.0.3`,
			}),
			Entry("%DOWNSTREAM_DIRECT_REMOTE_ADDRESS%", testCase{
				format:       `%DOWNSTREAM_DIRECT_REMOTE_ADDRESS%`,
				expectedHTTP: `127.0.0.4:53166`,
				expectedTCP:  `127.0.0.4:53166`,
			}),
			Entry("%DOWNSTREAM_DIRECT_REMOTE_ADDRESS_WITHOUT_PORT%", testCase{
				format:       `%DOWNSTREAM_DIRECT_REMOTE_ADDRESS_WITHOUT_PORT%`,
				expectedHTTP: `127.0.0.4`,
				expectedTCP:  `127.0.0.4`,
			}),
			Entry("%REQUESTED_SERVER_NAME%", testCase{
				format:       `%REQUESTED_SERVER_NAME%`,
				expectedHTTP: `backend.internal`,
				expectedTCP:  `backend.internal`,
			}),
			Entry("%ROUTE_NAME%", testCase{
				format:       `%ROUTE_NAME%`,
				expectedHTTP: `outbound:backend`,
				expectedTCP:  `outbound:backend`,
			}),
			Entry("%DOWNSTREAM_PEER_URI_SAN%", testCase{
				format:       `%DOWNSTREAM_PEER_URI_SAN%`,
				expectedHTTP: `spiffe://default/web`,
				expectedTCP:  `spiffe://default/web`,
			}),
			Entry("%DOWNSTREAM_LOCAL_URI_SAN%", testCase{
				format:       `%DOWNSTREAM_LOCAL_URI_SAN%`,
				expectedHTTP: `spiffe://default/backend`,
				expectedTCP:  `spiffe://default/backend`,
			}),
			Entry("%DOWNSTREAM_PEER_SUBJECT%", testCase{
				format:       `%DOWNSTREAM_PEER_SUBJECT%`,
				expectedHTTP: `CN=web,OU=IT,O=Webshop,L=San Francisco,ST=California,C=US`,
				expectedTCP:  `CN=web,OU=IT,O=Webshop,L=San Francisco,ST=California,C=US`,
			}),
			Entry("%DOWNSTREAM_LOCAL_SUBJECT%", testCase{
				format:       `%DOWNSTREAM_LOCAL_SUBJECT%`,
				expectedHTTP: `CN=backend,OU=IT,O=Webshop,L=San Francisco,ST=California,C=US`,
				expectedTCP:  `CN=backend,OU=IT,O=Webshop,L=San Francisco,ST=California,C=US`,
			}),
			Entry("%DOWNSTREAM_TLS_SESSION_ID%", testCase{
				format:       `%DOWNSTREAM_TLS_SESSION_ID%`,
				expectedHTTP: `b10662bf6bd4e6a068f0910d3d60c50f000355840fea4ce6844626b61c973901`,
				expectedTCP:  `b10662bf6bd4e6a068f0910d3d60c50f000355840fea4ce6844626b61c973901`,
			}),
			Entry("%DOWNSTREAM_TLS_CIPHER%", testCase{
				format:       `%DOWNSTREAM_TLS_CIPHER%`,
				expectedHTTP: `ECDHE-RSA-CHACHA20-POLY1305`,
				expectedTCP:  `ECDHE-RSA-CHACHA20-POLY1305`,
			}),
			Entry("%DOWNSTREAM_TLS_VERSION%", testCase{
				format:       `%DOWNSTREAM_TLS_VERSION%`,
				expectedHTTP: `TLSv1.2`,
				expectedTCP:  `TLSv1.2`,
			}),
			Entry("%DOWNSTREAM_PEER_FINGERPRINT_256%", testCase{
				format:       `%DOWNSTREAM_PEER_FINGERPRINT_256%`,
				expectedHTTP: `UNSUPPORTED_COMMAND(%DOWNSTREAM_PEER_FINGERPRINT_256%)`,
				expectedTCP:  `UNSUPPORTED_COMMAND(%DOWNSTREAM_PEER_FINGERPRINT_256%)`,
			}),
			Entry("%DOWNSTREAM_PEER_SERIAL%", testCase{
				format:       `%DOWNSTREAM_PEER_SERIAL%`,
				expectedHTTP: `UNSUPPORTED_COMMAND(%DOWNSTREAM_PEER_SERIAL%)`,
				expectedTCP:  `UNSUPPORTED_COMMAND(%DOWNSTREAM_PEER_SERIAL%)`,
			}),
			Entry("%DOWNSTREAM_PEER_ISSUER%", testCase{
				format:       `%DOWNSTREAM_PEER_ISSUER%`,
				expectedHTTP: `UNSUPPORTED_COMMAND(%DOWNSTREAM_PEER_ISSUER%)`,
				expectedTCP:  `UNSUPPORTED_COMMAND(%DOWNSTREAM_PEER_ISSUER%)`,
			}),
			Entry("%DOWNSTREAM_PEER_CERT%", testCase{
				format:       `%DOWNSTREAM_PEER_CERT%`,
				expectedHTTP: `UNSUPPORTED_COMMAND(%DOWNSTREAM_PEER_CERT%)`,
				expectedTCP:  `UNSUPPORTED_COMMAND(%DOWNSTREAM_PEER_CERT%)`,
			}),
			Entry("%DOWNSTREAM_PEER_CERT_V_START%", testCase{
				format:       `%DOWNSTREAM_PEER_CERT_V_START%`,
				expectedHTTP: `UNSUPPORTED_COMMAND(%DOWNSTREAM_PEER_CERT_V_START%)`,
				expectedTCP:  `UNSUPPORTED_COMMAND(%DOWNSTREAM_PEER_CERT_V_START%)`,
			}),
			Entry("%DOWNSTREAM_PEER_CERT_V_END%", testCase{
				format:       `%DOWNSTREAM_PEER_CERT_V_END%`,
				expectedHTTP: `UNSUPPORTED_COMMAND(%DOWNSTREAM_PEER_CERT_V_END%)`,
				expectedTCP:  `UNSUPPORTED_COMMAND(%DOWNSTREAM_PEER_CERT_V_END%)`,
			}),
			Entry("%HOSTNAME%", testCase{
				format:       `%HOSTNAME%`,
				expectedHTTP: `UNSUPPORTED_COMMAND(%HOSTNAME%)`,
				expectedTCP:  `UNSUPPORTED_COMMAND(%HOSTNAME%)`,
			}),
			Entry("%KUMA_SOURCE_ADDRESS%", testCase{
				format:       `%KUMA_SOURCE_ADDRESS%`,
				expectedHTTP: `%KUMA_SOURCE_ADDRESS%`, // placeholder must be rendered "as is"
				expectedTCP:  `%KUMA_SOURCE_ADDRESS%`,
			}),
			Entry("%KUMA_SOURCE_ADDRESS_WITHOUT_PORT%", testCase{
				format:       `%KUMA_SOURCE_ADDRESS_WITHOUT_PORT%`,
				expectedHTTP: `%KUMA_SOURCE_ADDRESS_WITHOUT_PORT%`, // placeholder must be rendered "as is"
				expectedTCP:  `%KUMA_SOURCE_ADDRESS_WITHOUT_PORT%`,
			}),
			Entry("%KUMA_SOURCE_SERVICE%", testCase{
				format:       `%KUMA_SOURCE_SERVICE%`,
				expectedHTTP: `%KUMA_SOURCE_SERVICE%`, // placeholder must be rendered "as is"
				expectedTCP:  `%KUMA_SOURCE_SERVICE%`,
			}),
			Entry("%KUMA_DESTINATION_SERVICE%", testCase{
				format:       `%KUMA_DESTINATION_SERVICE%`,
				expectedHTTP: `%KUMA_DESTINATION_SERVICE%`, // placeholder must be rendered "as is"
				expectedTCP:  `%KUMA_DESTINATION_SERVICE%`,
			}),
			Entry("composite", testCase{
				format:       `[%START_TIME%] "%REQ(:METHOD)% %REQ(X-ENVOY-ORIGINAL-PATH?:PATH)% %PROTOCOL%" %RESPONSE_CODE% %RESPONSE_FLAGS% %BYTES_RECEIVED% %BYTES_SENT% %DURATION% %RESP(X-ENVOY-UPSTREAM-SERVICE-TIME)% "%REQ(X-FORWARDED-FOR)%" "%REQ(USER-AGENT)%" "%REQ(X-REQUEST-ID)%" "%REQ(:AUTHORITY)%"`,
				expectedHTTP: `[2020-02-18T21:52:17.987Z] "- /api HTTP/1.1" 200 UF,URX 234 567 123 - "-" "-" "-" "backend.internal:8080"`,
				expectedTCP:  `[2020-02-18T21:52:17.987Z] "- - -" 0 UF,URX 234 567 123 - "-" "-" "-" "-"`,
			}),
			Entry("multi-line", testCase{
				format: `
[%START_TIME%]
"%REQ(:METHOD)% %REQ(X-ENVOY-ORIGINAL-PATH?:PATH)% %PROTOCOL%"
%RESPONSE_CODE%
%RESPONSE_FLAGS%
%BYTES_RECEIVED%
%BYTES_SENT%
%DURATION%
%RESP(X-ENVOY-UPSTREAM-SERVICE-TIME)%
"%REQ(X-FORWARDED-FOR)%"
"%REQ(USER-AGENT)%"
"%REQ(X-REQUEST-ID)%"
"%REQ(:AUTHORITY)%"
`,
				expectedHTTP: `
[2020-02-18T21:52:17.987Z]
"- /api HTTP/1.1"
200
UF,URX
234
567
123
-
"-"
"-"
"-"
"backend.internal:8080"
`,
				expectedTCP: `
[2020-02-18T21:52:17.987Z]
"- - -"
0
UF,URX
234
567
123
-
"-"
"-"
"-"
"-"
`,
			}),
		)
	})

	Context("invalid format string", func() {

		type testCase struct {
			format      string
			expectedErr string
		}

		DescribeTable("should reject an invalid format string",
			func(given testCase) {
				// when
				format, err := ParseFormat(given.format)
				// then
				Expect(format).To(BeNil())
				// and
				Expect(err).To(HaveOccurred())
				// and
				Expect(err.Error()).To(Equal(given.expectedErr))
			},
			Entry("unbalanced %", testCase{
				format:      `text with % character`,
				expectedErr: `format string is not valid: expected a command operator to start at position 11, instead got: "% character"`,
			}),
			Entry("%START_TIME(%", testCase{
				format:      `%START_TIME(%`,
				expectedErr: `format string is not valid: expected a command operator to start at position 1, instead got: "%START_TIME(%"`,
			}),
			Entry("%BYTES_RECEIVED()%", testCase{
				format:      `%BYTES_RECEIVED()%`,
				expectedErr: `format string is not valid: command "%BYTES_RECEIVED%" doesn't support arguments or max length constraint, instead got "%BYTES_RECEIVED()%"`,
			}),
			Entry("%REQ%", testCase{
				format:      `%REQ%`,
				expectedErr: `format string is not valid: command "%REQ(X?Y):Z%" requires a header and optional alternative header names as its arguments, instead got "%REQ%"`,
			}),
			Entry("%REQ:10%", testCase{
				format:      `%REQ:10%`,
				expectedErr: `format string is not valid: command "%REQ(X?Y):Z%" requires a header and optional alternative header names as its arguments, instead got "%REQ:10%"`,
			}),
			Entry("%REQ(header-1?header-2?header-3)%", testCase{
				format:      `%REQ(header-1?header-2?header-3)%`,
				expectedErr: `format string is not valid: more than 1 alternative header specified in "%REQ(header-1?header-2?header-3)%"`,
			}),
			Entry("%REQ(header-1?header-2?header-3):10%", testCase{
				format:      `%REQ(header-1?header-2?header-3):10%`,
				expectedErr: `format string is not valid: more than 1 alternative header specified in "%REQ(header-1?header-2?header-3):10%"`,
			}),
			Entry("%REQ(header-1\n?header-2)%", testCase{
				format:      "%REQ(header-1\n?header-2)%",
				expectedErr: "format string is not valid: header name contains a newline in \"%REQ(header-1\\n?header-2)%\"",
			}),
			Entry("%REQ(header-1?\rheader-2)%", testCase{
				format:      "%REQ(header-1?\rheader-2)%",
				expectedErr: "format string is not valid: header name contains a newline in \"%REQ(header-1?\\rheader-2)%\"",
			}),
			Entry("%REQ(header-1\x00)%", testCase{
				format:      "%REQ(header-1\x00)%",
				expectedErr: "format string is not valid: header name contains a newline in \"%REQ(header-1\\x00)%\"",
			}),
			Entry("%RESP%", testCase{
				format:      `%RESP%`,
				expectedErr: `format string is not valid: command "%RESP(X?Y):Z%" requires a header and optional alternative header names as its arguments, instead got "%RESP%"`,
			}),
			Entry("%RESP:10%", testCase{
				format:      `%RESP:10%`,
				expectedErr: `format string is not valid: command "%RESP(X?Y):Z%" requires a header and optional alternative header names as its arguments, instead got "%RESP:10%"`,
			}),
			Entry("%RESP(header-1?header-2?header-3)%", testCase{
				format:      `%RESP(header-1?header-2?header-3)%`,
				expectedErr: `format string is not valid: more than 1 alternative header specified in "%RESP(header-1?header-2?header-3)%"`,
			}),
			Entry("%RESP(header-1?header-2?header-3):10%", testCase{
				format:      `%RESP(header-1?header-2?header-3):10%`,
				expectedErr: `format string is not valid: more than 1 alternative header specified in "%RESP(header-1?header-2?header-3):10%"`,
			}),
			Entry("%RESP(header-1\n?header-2)%", testCase{
				format:      "%RESP(header-1\n?header-2)%",
				expectedErr: "format string is not valid: header name contains a newline in \"%RESP(header-1\\n?header-2)%\"",
			}),
			Entry("%RESP(header-1?\rheader-2)%", testCase{
				format:      "%RESP(header-1?\rheader-2)%",
				expectedErr: "format string is not valid: header name contains a newline in \"%RESP(header-1?\\rheader-2)%\"",
			}),
			Entry("%RESP(header-1\x00)%", testCase{
				format:      "%RESP(header-1\x00)%",
				expectedErr: "format string is not valid: header name contains a newline in \"%RESP(header-1\\x00)%\"",
			}),
			Entry("%TRAILER%", testCase{
				format:      `%TRAILER%`,
				expectedErr: `format string is not valid: command "%TRAILER(X?Y):Z%" requires a header and optional alternative header names as its arguments, instead got "%TRAILER%"`,
			}),
			Entry("%TRAILER:10%", testCase{
				format:      `%TRAILER:10%`,
				expectedErr: `format string is not valid: command "%TRAILER(X?Y):Z%" requires a header and optional alternative header names as its arguments, instead got "%TRAILER:10%"`,
			}),
			Entry("%TRAILER(header-1?header-2?header-3)%", testCase{
				format:      `%TRAILER(header-1?header-2?header-3)%`,
				expectedErr: `format string is not valid: more than 1 alternative header specified in "%TRAILER(header-1?header-2?header-3)%"`,
			}),
			Entry("%TRAILER(header-1?header-2?header-3):10%", testCase{
				format:      `%TRAILER(header-1?header-2?header-3):10%`,
				expectedErr: `format string is not valid: more than 1 alternative header specified in "%TRAILER(header-1?header-2?header-3):10%"`,
			}),
			Entry("%TRAILER(header-1\n?header-2)%", testCase{
				format:      "%TRAILER(header-1\n?header-2)%",
				expectedErr: "format string is not valid: header name contains a newline in \"%TRAILER(header-1\\n?header-2)%\"",
			}),
			Entry("%TRAILER(header-1?\rheader-2)%", testCase{
				format:      "%TRAILER(header-1?\rheader-2)%",
				expectedErr: "format string is not valid: header name contains a newline in \"%TRAILER(header-1?\\rheader-2)%\"",
			}),
			Entry("%TRAILER(header-1\x00)%", testCase{
				format:      "%TRAILER(header-1\x00)%",
				expectedErr: "format string is not valid: header name contains a newline in \"%TRAILER(header-1\\x00)%\"",
			}),
			Entry("%DYNAMIC_METADATA%", testCase{
				format:      `%DYNAMIC_METADATA%`,
				expectedErr: `format string is not valid: command "%DYNAMIC_METADATA(NAMESPACE:KEY*):Z%" requires a filter namespace and optional path as its arguments, instead got "%DYNAMIC_METADATA%"`,
			}),
			Entry("%DYNAMIC_METADATA:10%", testCase{
				format:      `%DYNAMIC_METADATA:10%`,
				expectedErr: `format string is not valid: command "%DYNAMIC_METADATA(NAMESPACE:KEY*):Z%" requires a filter namespace and optional path as its arguments, instead got "%DYNAMIC_METADATA:10%"`,
			}),
			Entry("%FILTER_STATE%", testCase{
				format:      `%FILTER_STATE%`,
				expectedErr: `format string is not valid: command "%FILTER_STATE(KEY):Z%" requires a key as its argument, instead got "%FILTER_STATE%"`,
			}),
			Entry("%FILTER_STATE:10%", testCase{
				format:      `%FILTER_STATE:10%`,
				expectedErr: `format string is not valid: command "%FILTER_STATE(KEY):Z%" requires a key as its argument, instead got "%FILTER_STATE:10%"`,
			}),
			Entry("%FILTER_STATE()%", testCase{
				format:      `%FILTER_STATE()%`,
				expectedErr: `format string is not valid: command "%FILTER_STATE(KEY):Z%" requires a key as its argument, instead got "%FILTER_STATE()%"`,
			}),
			Entry("%FILTER_STATE():10%", testCase{
				format:      `%FILTER_STATE():10%`,
				expectedErr: `format string is not valid: command "%FILTER_STATE(KEY):Z%" requires a key as its argument, instead got "%FILTER_STATE():10%"`,
			}),
		)
	})

	Context("support ConfigureHttpLog() and ConfigureTcpLog()", func() {

		type testCase struct {
			format       string
			expectedHTTP *accesslog_config.HttpGrpcAccessLogConfig // verify the entire config to make sure there are no unexpected changes
			expectedTCP  *accesslog_config.TcpGrpcAccessLogConfig  // verify the entire config to make sure there are no unexpected changes
		}

		DescribeTable("should configure properly",
			func(given testCase) {
				// when
				format, err := ParseFormat(given.format)
				// then
				Expect(err).ToNot(HaveOccurred())

				// given
				actualHTTP := &accesslog_config.HttpGrpcAccessLogConfig{}
				// when
				err = format.ConfigureHttpLog(actualHTTP)
				// then
				Expect(err).ToNot(HaveOccurred())
				// and
				Expect(actualHTTP).To(Equal(given.expectedHTTP))

				// given
				actualTCP := &accesslog_config.TcpGrpcAccessLogConfig{}
				// when
				err = format.ConfigureTcpLog(actualTCP)
				// then
				Expect(err).ToNot(HaveOccurred())
				// and
				Expect(actualTCP).To(Equal(given.expectedTCP))
			},
			Entry("commands without arguments should not have effect on the config", testCase{
				format:       `%START_TIME% %PROTOCOL%`,
				expectedHTTP: &accesslog_config.HttpGrpcAccessLogConfig{},
				expectedTCP:  &accesslog_config.TcpGrpcAccessLogConfig{},
			}),
			Entry("commands with arguments should add them to the config", testCase{
				format: `"%REQ(x-missing-header?:AUTHORITY):1%" "%RESP(DATE?server):2%" "%TRAILER(grpc-status?GRPC-MESSAGE):3%" "%DYNAMIC_METADATA(com.test.my_filter:test_object:inner_key):4%" "%FILTER_STATE(filter.state.key):5%"`,
				expectedHTTP: &accesslog_config.HttpGrpcAccessLogConfig{
					CommonConfig: &accesslog_config.CommonGrpcAccessLogConfig{
						FilterStateObjectsToLog: []string{"filter.state.key"},
					},
					AdditionalRequestHeadersToLog:   []string{"x-missing-header"}, // `:authority` header is captured by default and should not be added as an additional request header to log
					AdditionalResponseHeadersToLog:  []string{"date", "server"},
					AdditionalResponseTrailersToLog: []string{"grpc-status", "grpc-message"},
				},
				expectedTCP: &accesslog_config.TcpGrpcAccessLogConfig{
					CommonConfig: &accesslog_config.CommonGrpcAccessLogConfig{
						FilterStateObjectsToLog: []string{"filter.state.key"},
					},
				},
			}),
			Entry("config should not contain duplicate values", testCase{
				format: `
"%REQ(:AUTHORITY):1%" "%REQ(:path):2%" "%REQ(:authority):3%" 
"%REQ(CONTENT-TYPE):1%" "%REQ(origin):2%" "%REQ(content-type):3%"
"%RESP(SERVER):1%" "%RESP(content-type):2%" "%RESP(server):3%"
"%TRAILER(GRPC-STATUS):1%" "%TRAILER(grpc-message):2%" "%TRAILER(grpc-status):3%"
"%DYNAMIC_METADATA(com.test.my_filter:test_object:inner_key_1):1%" "%DYNAMIC_METADATA(com.test.my_filter:test_object:inner_key_2):2%" "%DYNAMIC_METADATA(com.test.my_filter:test_object:inner_key_1):3%"
"%FILTER_STATE(filter.state.key1):1%" "%FILTER_STATE(filter.state.key2):2%" "%FILTER_STATE(filter.state.key1):3%"
%BYTES_SENT%
%KUMA_SOURCE_SERVICE%
`,
				expectedHTTP: &accesslog_config.HttpGrpcAccessLogConfig{
					CommonConfig: &accesslog_config.CommonGrpcAccessLogConfig{
						FilterStateObjectsToLog: []string{"filter.state.key1", "filter.state.key2"},
					},
					AdditionalRequestHeadersToLog:   []string{"content-type", "origin"}, // only those headers that are not captured by default should be added as additional request headers to log
					AdditionalResponseHeadersToLog:  []string{"server", "content-type"},
					AdditionalResponseTrailersToLog: []string{"grpc-status", "grpc-message"},
				},
				expectedTCP: &accesslog_config.TcpGrpcAccessLogConfig{
					CommonConfig: &accesslog_config.CommonGrpcAccessLogConfig{
						FilterStateObjectsToLog: []string{"filter.state.key1", "filter.state.key2"},
					},
				},
			}),
		)
	})

	Describe("support String()", func() {
		type testCase struct {
			format   string
			expected string
		}

		DescribeTable("should return correct canonical representation",
			func(given testCase) {
				// when
				format, err := ParseFormat(given.format)
				// then
				Expect(err).ToNot(HaveOccurred())

				// when
				actual := format.String()
				// then
				Expect(actual).To(Equal(given.expected))
			},
			Entry("composite", testCase{
				format:   `[%START_TIME%] "%REQ(:METHOD)% %REQ(X-ENVOY-ORIGINAL-PATH?:PATH)% %PROTOCOL%" %RESPONSE_CODE% %RESPONSE_FLAGS% %BYTES_RECEIVED% %BYTES_SENT% %DURATION% %RESP(X-ENVOY-UPSTREAM-SERVICE-TIME)% "%REQ(X-FORWARDED-FOR)%" "%REQ(USER-AGENT)%" "%REQ(X-REQUEST-ID)%" "%REQ(:AUTHORITY)%"`,
				expected: `[%START_TIME%] "%REQ(:method)% %REQ(x-envoy-original-path?:path)% %PROTOCOL%" %RESPONSE_CODE% %RESPONSE_FLAGS% %BYTES_RECEIVED% %BYTES_SENT% %DURATION% %RESP(x-envoy-upstream-service-time)% "%REQ(x-forwarded-for)%" "%REQ(user-agent)%" "%REQ(x-request-id)%" "%REQ(:authority)%"`,
			}),
			Entry("multi-line", testCase{
				format: `
%START_TIME%
%START_TIME(%Y/%m/%dT%H:%M:%S%z %s)%
%REQ(:METHOD)%
%RESP(content-type?SERVER):10%
%TRAILER(PROTOCOL)%
%DYNAMIC_METADATA(com.test.my_filter:test_object:inner_key):10%
%FILTER_STATE(filter.state.key):10%
%BYTES_RECEIVED%
%BYTES_RECEIVED%
%BYTES_SENT%
%PROTOCOL%
%RESPONSE_CODE%
%RESPONSE_CODE_DETAILS%
%REQUEST_DURATION%
%RESPONSE_DURATION%
%RESPONSE_TX_DURATION%
%DURATION%
%RESPONSE_FLAGS%
%UPSTREAM_HOST%
%UPSTREAM_CLUSTER%
%UPSTREAM_LOCAL_ADDRESS%
%DOWNSTREAM_LOCAL_ADDRESS%
%DOWNSTREAM_LOCAL_ADDRESS_WITHOUT_PORT%
%DOWNSTREAM_REMOTE_ADDRESS%
%DOWNSTREAM_REMOTE_ADDRESS_WITHOUT_PORT%
%DOWNSTREAM_DIRECT_REMOTE_ADDRESS%
%DOWNSTREAM_DIRECT_REMOTE_ADDRESS_WITHOUT_PORT%
%REQUESTED_SERVER_NAME%
%ROUTE_NAME%
%DOWNSTREAM_PEER_URI_SAN%
%DOWNSTREAM_LOCAL_URI_SAN%
%DOWNSTREAM_PEER_SUBJECT%
%DOWNSTREAM_LOCAL_SUBJECT%
%DOWNSTREAM_TLS_SESSION_ID%
%DOWNSTREAM_TLS_CIPHER%
%DOWNSTREAM_TLS_VERSION%
%UPSTREAM_TRANSPORT_FAILURE_REASON%
%DOWNSTREAM_PEER_FINGERPRINT_256%
%DOWNSTREAM_PEER_SERIAL%
%DOWNSTREAM_PEER_ISSUER%
%DOWNSTREAM_PEER_CERT%
%DOWNSTREAM_PEER_CERT_V_START%
%DOWNSTREAM_PEER_CERT_V_END%
%HOSTNAME%
%KUMA_SOURCE_ADDRESS%
%KUMA_SOURCE_ADDRESS_WITHOUT_PORT%
%KUMA_SOURCE_SERVICE%
%KUMA_DESTINATION_SERVICE%
`,
				expected: `
%START_TIME%
%START_TIME(%Y/%m/%dT%H:%M:%S%z %s)%
%REQ(:method)%
%RESP(content-type?server):10%
%TRAILER(protocol)%
%DYNAMIC_METADATA(com.test.my_filter:test_object:inner_key):10%
%FILTER_STATE(filter.state.key):10%
%BYTES_RECEIVED%
%BYTES_RECEIVED%
%BYTES_SENT%
%PROTOCOL%
%RESPONSE_CODE%
%RESPONSE_CODE_DETAILS%
%REQUEST_DURATION%
%RESPONSE_DURATION%
%RESPONSE_TX_DURATION%
%DURATION%
%RESPONSE_FLAGS%
%UPSTREAM_HOST%
%UPSTREAM_CLUSTER%
%UPSTREAM_LOCAL_ADDRESS%
%DOWNSTREAM_LOCAL_ADDRESS%
%DOWNSTREAM_LOCAL_ADDRESS_WITHOUT_PORT%
%DOWNSTREAM_REMOTE_ADDRESS%
%DOWNSTREAM_REMOTE_ADDRESS_WITHOUT_PORT%
%DOWNSTREAM_DIRECT_REMOTE_ADDRESS%
%DOWNSTREAM_DIRECT_REMOTE_ADDRESS_WITHOUT_PORT%
%REQUESTED_SERVER_NAME%
%ROUTE_NAME%
%DOWNSTREAM_PEER_URI_SAN%
%DOWNSTREAM_LOCAL_URI_SAN%
%DOWNSTREAM_PEER_SUBJECT%
%DOWNSTREAM_LOCAL_SUBJECT%
%DOWNSTREAM_TLS_SESSION_ID%
%DOWNSTREAM_TLS_CIPHER%
%DOWNSTREAM_TLS_VERSION%
%UPSTREAM_TRANSPORT_FAILURE_REASON%
%DOWNSTREAM_PEER_FINGERPRINT_256%
%DOWNSTREAM_PEER_SERIAL%
%DOWNSTREAM_PEER_ISSUER%
%DOWNSTREAM_PEER_CERT%
%DOWNSTREAM_PEER_CERT_V_START%
%DOWNSTREAM_PEER_CERT_V_END%
%HOSTNAME%
%KUMA_SOURCE_ADDRESS%
%KUMA_SOURCE_ADDRESS_WITHOUT_PORT%
%KUMA_SOURCE_SERVICE%
%KUMA_DESTINATION_SERVICE%
`,
			}),
		)
	})

	Describe("support Interpolate()", func() {
		type testCase struct {
			format   string
			context  map[string]string
			expected string
		}

		DescribeTable("should bind to a given context",
			func(given testCase) {
				// when
				format, err := ParseFormat(given.format)
				// then
				Expect(err).ToNot(HaveOccurred())

				// when
				interpolatedFormat, err := format.Interpolate(InterpolationVariables(given.context))
				// then
				Expect(err).ToNot(HaveOccurred())

				// when
				actual := interpolatedFormat.String()
				// then
				Expect(actual).To(Equal(given.expected))
			},
			Entry("multi-line w/ empty context", testCase{
				format: `
%START_TIME%
%KUMA_SOURCE_ADDRESS%
%DURATION%
%KUMA_SOURCE_ADDRESS_WITHOUT_PORT%
%BYTES_RECEIVED%
%KUMA_SOURCE_SERVICE%
%BYTES_SENT%
%KUMA_DESTINATION_SERVICE%
%PROTOCOL%
`,
				context: nil,
				expected: `
%START_TIME%

%DURATION%

%BYTES_RECEIVED%

%BYTES_SENT%

%PROTOCOL%
`,
			}),
			Entry("multi-line w/ full context", testCase{
				format: `
%START_TIME%
%KUMA_SOURCE_ADDRESS%
%DURATION%
%KUMA_SOURCE_ADDRESS_WITHOUT_PORT%
%BYTES_RECEIVED%
%KUMA_SOURCE_SERVICE%
%BYTES_SENT%
%KUMA_DESTINATION_SERVICE%
%PROTOCOL%
`,
				context: map[string]string{
					"KUMA_SOURCE_ADDRESS":              "10.0.0.3:0",
					"KUMA_SOURCE_ADDRESS_WITHOUT_PORT": "10.0.0.3",
					"KUMA_SOURCE_SERVICE":              "web",
					"KUMA_DESTINATION_SERVICE":         "backend",
				},
				expected: `
%START_TIME%
10.0.0.3:0
%DURATION%
10.0.0.3
%BYTES_RECEIVED%
web
%BYTES_SENT%
backend
%PROTOCOL%
`,
			}),
		)
	})
})
