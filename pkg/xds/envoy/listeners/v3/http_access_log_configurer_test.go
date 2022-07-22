package v3_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/test/resources/model"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	. "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
)

var _ = Describe("HttpAccessLogConfigurer", func() {

	type testCase struct {
		listenerName       string
		listenerAddress    string
		listenerPort       uint32
		listenerProtocol   core_xds.SocketAddressProtocol
		statsName          string
		routeName          string
		backend            *mesh_proto.LoggingBackend
		legacyTcpAccessLog bool
		expected           string
	}

	DescribeTable("should generate proper Envoy config",
		func(given testCase) {
			// given
			mesh := "demo"
			sourceService := "web"
			destinationService := "backend"
			metaData := &core_xds.DataplaneMetadata{}
			if !given.legacyTcpAccessLog {
				metaData.Features = core_xds.Features{core_xds.FeatureTCPAccessLogViaNamedPipe: true}
			}
			proxy := &core_xds.Proxy{
				Id:       *core_xds.BuildProxyId("web", "example"),
				Metadata: metaData,
				Dataplane: &core_mesh.DataplaneResource{
					Meta: &model.ResourceMeta{
						Name: "dataplane0",
					},
					Spec: &mesh_proto.Dataplane{
						Networking: &mesh_proto.Dataplane_Networking{
							Address: "192.168.0.1",
							Inbound: []*mesh_proto.Dataplane_Networking_Inbound{{
								Port:        80,
								ServicePort: 8080,
								Tags: map[string]string{
									"kuma.io/service": "web",
								},
							}},
							Outbound: []*mesh_proto.Dataplane_Networking_Outbound{{
								Port:    27070,
								Service: "backend",
							}},
						},
					},
				},
			}

			// when
			listener, err := NewListenerBuilder(envoy.APIV3).
				Configure(OutboundListener(given.listenerName, given.listenerAddress, given.listenerPort, given.listenerProtocol)).
				Configure(FilterChain(NewFilterChainBuilder(envoy.APIV3).
					Configure(HttpConnectionManager(given.statsName, false)).
					Configure(HttpAccessLog(mesh, envoy.TrafficDirectionOutbound, sourceService, destinationService, given.backend, proxy)))).
				Build()
			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			actual, err := util_proto.ToYAML(listener)
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(actual).To(MatchYAML(given.expected))
		},
		Entry("basic http_connection_manager without access log", testCase{
			listenerName:    "outbound:127.0.0.1:27070",
			listenerAddress: "127.0.0.1",
			listenerPort:    27070,
			statsName:       "backend",
			routeName:       "outbound:backend",
			backend:         nil,
			expected: `
            name: outbound:127.0.0.1:27070
            address:
              socketAddress:
                address: 127.0.0.1
                portValue: 27070
            filterChains:
            - filters:
              - name: envoy.filters.network.http_connection_manager
                typedConfig:
                  '@type': type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager
                  httpFilters:
                  - name: envoy.filters.http.router
                    typedConfig:
                      '@type': type.googleapis.com/envoy.extensions.filters.http.router.v3.Router
                  statPrefix: backend
            trafficDirection: OUTBOUND
`,
		}),
		Entry("basic http_connection_manager with file access log", testCase{
			listenerName:    "outbound:127.0.0.1:27070",
			listenerAddress: "127.0.0.1",
			listenerPort:    27070,
			statsName:       "backend",
			routeName:       "outbound:backend",
			backend: &mesh_proto.LoggingBackend{
				Name: "file",
				Type: mesh_proto.LoggingFileType,
				Conf: util_proto.MustToStruct(&mesh_proto.FileLoggingBackendConfig{
					Path: "/tmp/log",
				}),
			},
			expected: `
            address:
              socketAddress:
                address: 127.0.0.1
                portValue: 27070
            filterChains:
            - filters:
              - name: envoy.filters.network.http_connection_manager
                typedConfig:
                  '@type': type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager
                  accessLog:
                  - name: envoy.access_loggers.file
                    typedConfig:
                      '@type': type.googleapis.com/envoy.extensions.access_loggers.file.v3.FileAccessLog
                      logFormat:
                        textFormatSource:
                          inlineString: |+
                            [%START_TIME%] demo "%REQ(:method)% %REQ(x-envoy-original-path?:path)% %PROTOCOL%" %RESPONSE_CODE% %RESPONSE_FLAGS% %BYTES_RECEIVED% %BYTES_SENT% %DURATION% %RESP(x-envoy-upstream-service-time)% "%REQ(x-forwarded-for)%" "%REQ(user-agent)%" "%REQ(x-b3-traceid?x-datadog-traceid)%" "%REQ(x-request-id)%" "%REQ(:authority)%" "web" "backend" "192.168.0.1" "%UPSTREAM_HOST%"
                      path: /tmp/log
                  httpFilters:
                  - name: envoy.filters.http.router
                    typedConfig:
                      '@type': type.googleapis.com/envoy.extensions.filters.http.router.v3.Router
                  statPrefix: backend
            name: outbound:127.0.0.1:27070
            trafficDirection: OUTBOUND`,
		}),
		Entry("basic http_connection_manager with tcp access log", testCase{
			listenerName:    "outbound:127.0.0.1:27070",
			listenerAddress: "127.0.0.1",
			listenerPort:    27070,
			statsName:       "backend",
			routeName:       "outbound:backend",
			backend: &mesh_proto.LoggingBackend{
				Name: "tcp",
				Format: `[%START_TIME%] "%REQ(X-REQUEST-ID)%" "%REQ(:AUTHORITY)%" "%REQ(ORIGIN)%" "%REQ(CONTENT-TYPE)%" "%KUMA_SOURCE_SERVICE%" "%KUMA_DESTINATION_SERVICE%" "%KUMA_SOURCE_ADDRESS%" "%KUMA_SOURCE_ADDRESS_WITHOUT_PORT%" "%UPSTREAM_HOST%"
"%RESP(SERVER):5%" "%TRAILER(GRPC-MESSAGE):7%" "DYNAMIC_METADATA(namespace:object:key):9" "FILTER_STATE(filter.state.key):12"
`, // intentional newline at the end
				Type: mesh_proto.LoggingTcpType,
				Conf: util_proto.MustToStruct(&mesh_proto.TcpLoggingBackendConfig{
					Address: "127.0.0.1:1234",
				}),
			},
			expected: `
            address:
              socketAddress:
                address: 127.0.0.1
                portValue: 27070
            filterChains:
            - filters:
              - name: envoy.filters.network.http_connection_manager
                typedConfig:
                  '@type': type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager
                  accessLog:
                  - name: envoy.access_loggers.file
                    typedConfig:
                      '@type': type.googleapis.com/envoy.extensions.access_loggers.file.v3.FileAccessLog
                      logFormat:
                        textFormatSource:
                          inlineString: |+
                            127.0.0.1:1234;[%START_TIME%] "%REQ(x-request-id)%" "%REQ(:authority)%" "%REQ(origin)%" "%REQ(content-type)%" "web" "backend" "192.168.0.1:0" "192.168.0.1" "%UPSTREAM_HOST%"
                            "%RESP(server):5%" "%TRAILER(grpc-message):7%" "DYNAMIC_METADATA(namespace:object:key):9" "FILTER_STATE(filter.state.key):12"

                      path: /tmp/kuma-al-dataplane0-demo.sock
                  httpFilters:
                  - name: envoy.filters.http.router
                    typedConfig:
                      '@type': type.googleapis.com/envoy.extensions.filters.http.router.v3.Router
                  statPrefix: backend
            name: outbound:127.0.0.1:27070
            trafficDirection: OUTBOUND`,
		}),
		Entry("basic http_connection_manager with legacy tcp access log", testCase{
			listenerName:    "outbound:127.0.0.1:27070",
			listenerAddress: "127.0.0.1",
			listenerPort:    27070,
			statsName:       "backend",
			routeName:       "outbound:backend",
			backend: &mesh_proto.LoggingBackend{
				Name: "tcp",
				Format: `[%START_TIME%] "%REQ(X-REQUEST-ID)%" "%REQ(:AUTHORITY)%" "%REQ(ORIGIN)%" "%REQ(CONTENT-TYPE)%" "%KUMA_SOURCE_SERVICE%" "%KUMA_DESTINATION_SERVICE%" "%KUMA_SOURCE_ADDRESS%" "%KUMA_SOURCE_ADDRESS_WITHOUT_PORT%" "%UPSTREAM_HOST%"
"%RESP(SERVER):5%" "%TRAILER(GRPC-MESSAGE):7%" "DYNAMIC_METADATA(namespace:object:key):9" "FILTER_STATE(filter.state.key):12"
`, // intentional newline at the end
				Type: mesh_proto.LoggingTcpType,
				Conf: util_proto.MustToStruct(&mesh_proto.TcpLoggingBackendConfig{
					Address: "127.0.0.1:1234",
				}),
			},
			legacyTcpAccessLog: true,
			expected: `
            address:
              socketAddress:
                address: 127.0.0.1
                portValue: 27070
            filterChains:
            - filters:
              - name: envoy.filters.network.http_connection_manager
                typedConfig:
                  '@type': type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager
                  accessLog:
                  - name: envoy.access_loggers.http_grpc
                    typedConfig:
                      '@type': type.googleapis.com/envoy.extensions.access_loggers.grpc.v3.HttpGrpcAccessLogConfig
                      additionalRequestHeadersToLog:
                      - origin
                      - content-type
                      additionalResponseHeadersToLog:
                      - server
                      additionalResponseTrailersToLog:
                      - grpc-message
                      commonConfig:
                        grpcService:
                          envoyGrpc:
                            clusterName: access_log_sink
                        logName: |+
                          127.0.0.1:1234;[%START_TIME%] "%REQ(x-request-id)%" "%REQ(:authority)%" "%REQ(origin)%" "%REQ(content-type)%" "web" "backend" "192.168.0.1:0" "192.168.0.1" "%UPSTREAM_HOST%"
                          "%RESP(server):5%" "%TRAILER(grpc-message):7%" "DYNAMIC_METADATA(namespace:object:key):9" "FILTER_STATE(filter.state.key):12"

                        transportApiVersion: V3
                  httpFilters:
                  - name: envoy.filters.http.router
                    typedConfig:
                      '@type': type.googleapis.com/envoy.extensions.filters.http.router.v3.Router
                  statPrefix: backend
            name: outbound:127.0.0.1:27070
            trafficDirection: OUTBOUND`,
		}),
	)
})
