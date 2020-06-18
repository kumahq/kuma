package listeners_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	. "github.com/Kong/kuma/pkg/xds/envoy/listeners"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	mesh_core "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	"github.com/Kong/kuma/pkg/core/xds"
	core_xds "github.com/Kong/kuma/pkg/core/xds"
	util_proto "github.com/Kong/kuma/pkg/util/proto"
)

var _ = Describe("HttpAccessLogConfigurer", func() {

	type testCase struct {
		listenerName    string
		listenerAddress string
		listenerPort    uint32
		statsName       string
		routeName       string
		backend         *mesh_proto.LoggingBackend
		expected        string
	}

	DescribeTable("should generate proper Envoy config",
		func(given testCase) {
			// given
			mesh := "demo"
			sourceService := "web"
			destinationService := "backend"
			proxy := &core_xds.Proxy{
				Id: xds.ProxyId{
					Name: "web",
					Mesh: "example",
				},
				Dataplane: &mesh_core.DataplaneResource{
					Spec: mesh_proto.Dataplane{
						Networking: &mesh_proto.Dataplane_Networking{
							Address: "192.168.0.1",
							Inbound: []*mesh_proto.Dataplane_Networking_Inbound{{
								Port:        80,
								ServicePort: 8080,
								Tags: map[string]string{
									"service": "web",
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
			listener, err := NewListenerBuilder().
				Configure(OutboundListener(given.listenerName, given.listenerAddress, given.listenerPort)).
				Configure(FilterChain(NewFilterChainBuilder().
					Configure(HttpConnectionManager(given.statsName)).
					Configure(HttpOutboundRoute(given.routeName)).
					Configure(HttpAccessLog(mesh, TrafficDirectionOutbound, sourceService, destinationService, given.backend, proxy)))).
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
              - name: envoy.http_connection_manager
                typedConfig:
                  '@type': type.googleapis.com/envoy.config.filter.network.http_connection_manager.v2.HttpConnectionManager
                  httpFilters:
                  - name: envoy.router
                  rds:
                    configSource:
                      ads: {}
                    routeConfigName: outbound:backend
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
            name: outbound:127.0.0.1:27070
            address:
              socketAddress:
                address: 127.0.0.1
                portValue: 27070
            filterChains:
            - filters:
              - name: envoy.http_connection_manager
                typedConfig:
                  '@type': type.googleapis.com/envoy.config.filter.network.http_connection_manager.v2.HttpConnectionManager
                  accessLog:
                  - name: envoy.file_access_log
                    typedConfig:
                      '@type': type.googleapis.com/envoy.config.accesslog.v2.FileAccessLog
                      format: |
                        [%START_TIME%] demo "%REQ(:method)% %REQ(x-envoy-original-path?:path)% %PROTOCOL%" %RESPONSE_CODE% %RESPONSE_FLAGS% %BYTES_RECEIVED% %BYTES_SENT% %DURATION% %RESP(x-envoy-upstream-service-time)% "%REQ(x-forwarded-for)%" "%REQ(user-agent)%" "%REQ(x-request-id)%" "%REQ(:authority)%" "web" "backend" "192.168.0.1" "%UPSTREAM_HOST%"
                      path: /tmp/log
                  httpFilters:
                  - name: envoy.router
                  rds:
                    configSource:
                      ads: {}
                    routeConfigName: outbound:backend
                  statPrefix: backend
            trafficDirection: OUTBOUND
`,
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
            name: outbound:127.0.0.1:27070
            address:
              socketAddress:
                address: 127.0.0.1
                portValue: 27070
            filterChains:
            - filters:
              - name: envoy.http_connection_manager
                typedConfig:
                  '@type': type.googleapis.com/envoy.config.filter.network.http_connection_manager.v2.HttpConnectionManager
                  accessLog:
                  - name: envoy.http_grpc_access_log
                    typedConfig:
                      '@type': type.googleapis.com/envoy.config.accesslog.v2.HttpGrpcAccessLogConfig
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
                        logName: |
                          127.0.0.1:1234;[%START_TIME%] "%REQ(x-request-id)%" "%REQ(:authority)%" "%REQ(origin)%" "%REQ(content-type)%" "web" "backend" "192.168.0.1:0" "192.168.0.1" "%UPSTREAM_HOST%"

                          "%RESP(server):5%" "%TRAILER(grpc-message):7%" "DYNAMIC_METADATA(namespace:object:key):9" "FILTER_STATE(filter.state.key):12"
                  httpFilters:
                  - name: envoy.router
                  rds:
                    configSource:
                      ads: {}
                    routeConfigName: outbound:backend
                  statPrefix: backend
            trafficDirection: OUTBOUND
`,
		}),
	)
})
