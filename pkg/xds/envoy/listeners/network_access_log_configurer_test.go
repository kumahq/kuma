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
	envoy_common "github.com/Kong/kuma/pkg/xds/envoy"
)

var _ = Describe("NetworkAccessLogConfigurer", func() {

	type testCase struct {
		listenerName    string
		listenerAddress string
		listenerPort    uint32
		statsName       string
		clusters        []envoy_common.ClusterInfo
		backend         *mesh_proto.LoggingBackend
		expected        string
	}

	DescribeTable("should generate proper Envoy config",
		func(given testCase) {
			// given
			sourceService := "backend"
			destinationService := "db"
			proxy := &core_xds.Proxy{
				Id: xds.ProxyId{
					Name: "backend",
					Mesh: "example",
				},
				Dataplane: &mesh_core.DataplaneResource{
					Spec: mesh_proto.Dataplane{
						Networking: &mesh_proto.Dataplane_Networking{
							Address: "192.168.0.1",
							Inbound: []*mesh_proto.Dataplane_Networking_Inbound{{
								Port:        1234,
								ServicePort: 8765,
								Tags: map[string]string{
									"service": "backend",
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
			listener, err := NewListenerBuilder().
				Configure(OutboundListener(given.listenerName, given.listenerAddress, given.listenerPort)).
				Configure(FilterChain(NewFilterChainBuilder().
					Configure(TcpProxy(given.statsName, given.clusters...)).
					Configure(NetworkAccessLog(sourceService, destinationService, given.backend, proxy)))).
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
			clusters:        []envoy_common.ClusterInfo{{Name: "db", Weight: 200}},
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
              - name: envoy.tcp_proxy
                typedConfig:
                  '@type': type.googleapis.com/envoy.config.filter.network.tcp_proxy.v2.TcpProxy
                  cluster: db
                  statPrefix: db
`,
		}),
		Entry("basic tcp_proxy with file access log", testCase{
			listenerName:    "outbound:127.0.0.1:5432",
			listenerAddress: "127.0.0.1",
			listenerPort:    5432,
			statsName:       "db",
			clusters:        []envoy_common.ClusterInfo{{Name: "db", Weight: 200}},
			backend: &mesh_proto.LoggingBackend{
				Name: "file",
				Type: &mesh_proto.LoggingBackend_File_{
					File: &mesh_proto.LoggingBackend_File{
						Path: "/tmp/log",
					},
				},
			},
			expected: `
            name: outbound:127.0.0.1:5432
            trafficDirection: OUTBOUND
            address:
              socketAddress:
                address: 127.0.0.1
                portValue: 5432
            filterChains:
            - filters:
              - name: envoy.tcp_proxy
                typedConfig:
                  '@type': type.googleapis.com/envoy.config.filter.network.tcp_proxy.v2.TcpProxy
                  accessLog:
                  - name: envoy.file_access_log
                    typedConfig:
                      '@type': type.googleapis.com/envoy.config.accesslog.v2.FileAccessLog
                      format: |
                        [%START_TIME%] 192.168.0.1:0(backend)->%UPSTREAM_HOST%(db) took %DURATION%ms, sent %BYTES_SENT% bytes, received: %BYTES_RECEIVED% bytes
                      path: /tmp/log
                  cluster: db
                  statPrefix: db
`,
		}),
		Entry("basic tcp_proxy with tcp access log", testCase{
			listenerName:    "outbound:127.0.0.1:5432",
			listenerAddress: "127.0.0.1",
			listenerPort:    5432,
			statsName:       "db",
			clusters:        []envoy_common.ClusterInfo{{Name: "db", Weight: 200}},
			backend: &mesh_proto.LoggingBackend{
				Name:   "tcp",
				Format: "custom format",
				Type: &mesh_proto.LoggingBackend_Tcp_{
					Tcp: &mesh_proto.LoggingBackend_Tcp{
						Address: "127.0.0.1:1234",
					},
				},
			},
			expected: `
            name: outbound:127.0.0.1:5432
            trafficDirection: OUTBOUND
            address:
              socketAddress:
                address: 127.0.0.1
                portValue: 5432
            filterChains:
            - filters:
              - name: envoy.tcp_proxy
                typedConfig:
                  '@type': type.googleapis.com/envoy.config.filter.network.tcp_proxy.v2.TcpProxy
                  accessLog:
                  - name: envoy.http_grpc_access_log
                    typedConfig:
                      '@type': type.googleapis.com/envoy.config.accesslog.v2.HttpGrpcAccessLogConfig
                      commonConfig:
                        grpcService:
                          envoyGrpc:
                            clusterName: access_log_sink
                        logName: 127.0.0.1:1234;custom format
                  cluster: db
                  statPrefix: db
`,
		}),
	)
})
