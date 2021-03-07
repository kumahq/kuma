package v3_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/pkg/xds/envoy/listeners"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	mesh_core "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/xds"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
)

var _ = Describe("NetworkAccessLogConfigurer", func() {

	type testCase struct {
		listenerName    string
		listenerAddress string
		listenerPort    uint32
		statsName       string
		clusters        []envoy_common.ClusterSubset
		backend         *mesh_proto.LoggingBackend
		expected        string
	}

	DescribeTable("should generate proper Envoy config",
		func(given testCase) {
			// given
			meshName := "demo"
			sourceService := "backend"
			destinationService := "db"
			proxy := &core_xds.Proxy{
				Id: xds.ProxyId{
					Name: "backend",
					Mesh: "example",
				},
				Dataplane: &mesh_core.DataplaneResource{
					Spec: &mesh_proto.Dataplane{
						Networking: &mesh_proto.Dataplane_Networking{
							Address: "192.168.0.1",
							Inbound: []*mesh_proto.Dataplane_Networking_Inbound{{
								Port:        1234,
								ServicePort: 8765,
								Tags: map[string]string{
									"kuma.io/service": "backend",
								},
							}},
							Outbound: []*mesh_proto.Dataplane_Networking_Outbound{{
								Port:    15432,
								Service: "db",
							}},
						},
					},
				},
			}

			// when
			listener, err := NewListenerBuilder(envoy_common.APIV3).
				Configure(OutboundListener(given.listenerName, given.listenerAddress, given.listenerPort)).
				Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3).
					Configure(TcpProxy(given.statsName, given.clusters...)).
					Configure(NetworkAccessLog(meshName, envoy_common.TrafficDirectionUnspecified, sourceService, destinationService, given.backend, proxy)))).
				Build()
			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			actual, err := util_proto.ToYAML(listener)
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(actual).To(MatchYAML(given.expected))
		},
		Entry("basic tcp_proxy without access log", testCase{
			listenerName:    "outbound:127.0.0.1:5432",
			listenerAddress: "127.0.0.1",
			listenerPort:    5432,
			statsName:       "db",
			clusters:        []envoy_common.ClusterSubset{{ClusterName: "db", Weight: 200}},
			backend:         nil,
			expected: `
            name: outbound:127.0.0.1:5432
            trafficDirection: OUTBOUND
            address:
              socketAddress:
                address: 127.0.0.1
                portValue: 5432
            filterChains:
            - filters:
              - name: envoy.filters.network.tcp_proxy
                typedConfig:
                  '@type': type.googleapis.com/envoy.extensions.filters.network.tcp_proxy.v3.TcpProxy
                  cluster: db
                  statPrefix: db
`,
		}),
		Entry("basic tcp_proxy with file access log", testCase{
			listenerName:    "outbound:127.0.0.1:5432",
			listenerAddress: "127.0.0.1",
			listenerPort:    5432,
			statsName:       "db",
			clusters:        []envoy_common.ClusterSubset{{ClusterName: "db", Weight: 200}},
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
                portValue: 5432
            filterChains:
            - filters:
              - name: envoy.filters.network.tcp_proxy
                typedConfig:
                  '@type': type.googleapis.com/envoy.extensions.filters.network.tcp_proxy.v3.TcpProxy
                  accessLog:
                  - name: envoy.access_loggers.file
                    typedConfig:
                      '@type': type.googleapis.com/envoy.extensions.access_loggers.file.v3.FileAccessLog
                      format: |+
                        [%START_TIME%] %RESPONSE_FLAGS% demo 192.168.0.1(backend)->%UPSTREAM_HOST%(db) took %DURATION%ms, sent %BYTES_SENT% bytes, received: %BYTES_RECEIVED% bytes

                      path: /tmp/log
                  cluster: db
                  statPrefix: db
            name: outbound:127.0.0.1:5432
            trafficDirection: OUTBOUND
`,
		}),
		Entry("basic tcp_proxy with tcp access log", testCase{
			listenerName:    "outbound:127.0.0.1:5432",
			listenerAddress: "127.0.0.1",
			listenerPort:    5432,
			statsName:       "db",
			clusters:        []envoy_common.ClusterSubset{{ClusterName: "db", Weight: 200}},
			backend: &mesh_proto.LoggingBackend{
				Name: "tcp",
				Format: `[%START_TIME%] "%REQ(X-REQUEST-ID)%" "%REQ(:AUTHORITY)%" "%REQ(ORIGIN)%" "%REQ(CONTENT-TYPE)%" "%KUMA_SOURCE_SERVICE%" "%KUMA_DESTINATION_SERVICE%" "%KUMA_SOURCE_ADDRESS%" "%KUMA_SOURCE_ADDRESS_WITHOUT_PORT%" "%UPSTREAM_HOST%

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
                portValue: 5432
            filterChains:
            - filters:
              - name: envoy.filters.network.tcp_proxy
                typedConfig:
                  '@type': type.googleapis.com/envoy.extensions.filters.network.tcp_proxy.v3.TcpProxy
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
                          127.0.0.1:1234;[%START_TIME%] "%REQ(x-request-id)%" "%REQ(:authority)%" "%REQ(origin)%" "%REQ(content-type)%" "backend" "db" "192.168.0.1:0" "192.168.0.1" "%UPSTREAM_HOST%
            
                          "%RESP(server):5%" "%TRAILER(grpc-message):7%" "DYNAMIC_METADATA(namespace:object:key):9" "FILTER_STATE(filter.state.key):12"
            
                        transportApiVersion: V3
                  cluster: db
                  statPrefix: db
            name: outbound:127.0.0.1:5432
            trafficDirection: OUTBOUND`,
		}),
	)
})
