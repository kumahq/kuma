package v3_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	xds_types "github.com/kumahq/kuma/pkg/core/xds/types"
	"github.com/kumahq/kuma/pkg/test/resources/model"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	. "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
)

var _ = Describe("HttpAccessLogConfigurer", func() {
	type testCase struct {
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
			metaData := &core_xds.DataplaneMetadata{WorkDir: "/tmp"}
			if !given.legacyTcpAccessLog {
				metaData.Features = xds_types.Features{xds_types.FeatureTCPAccessLogViaNamedPipe: true}
			}
			proxy := &core_xds.Proxy{
				Id:       *core_xds.BuildProxyId(mesh, "dataplane0"),
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
								Port: 27070,
								Tags: map[string]string{
									mesh_proto.ServiceTag: "backend",
								},
							}},
						},
					},
				},
			}

			// when
			listener, err := NewOutboundListenerBuilder(envoy.APIV3, given.listenerAddress, given.listenerPort, given.listenerProtocol).
				Configure(FilterChain(NewFilterChainBuilder(envoy.APIV3, envoy.AnonymousResource).
					Configure(HttpConnectionManager(given.statsName, false, nil)).
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
                  internalAddressConfig:
                    cidrRanges:
                      - addressPrefix: 127.0.0.1
                        prefixLen: 32
                      - addressPrefix: ::1
                        prefixLen: 128
            trafficDirection: OUTBOUND
`,
		}),
		Entry("basic http_connection_manager with file access log", testCase{
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
                            [%START_TIME%] demo "%REQ(:METHOD)% %REQ(X-ENVOY-ORIGINAL-PATH?:PATH)% %PROTOCOL%" %RESPONSE_CODE% %RESPONSE_FLAGS% %BYTES_RECEIVED% %BYTES_SENT% %DURATION% %RESP(X-ENVOY-UPSTREAM-SERVICE-TIME)% "%REQ(X-FORWARDED-FOR)%" "%REQ(USER-AGENT)%" "%REQ(X-B3-TRACEID?X-DATADOG-TRACEID)%" "%REQ(X-REQUEST-ID)%" "%REQ(:AUTHORITY)%" "web" "backend" "192.168.0.1" "%UPSTREAM_HOST%"
                      path: /tmp/log
                  httpFilters:
                  - name: envoy.filters.http.router
                    typedConfig:
                      '@type': type.googleapis.com/envoy.extensions.filters.http.router.v3.Router
                  statPrefix: backend
                  internalAddressConfig:
                    cidrRanges:
                      - addressPrefix: 127.0.0.1
                        prefixLen: 32
                      - addressPrefix: ::1
                        prefixLen: 128
            name: outbound:127.0.0.1:27070
            trafficDirection: OUTBOUND`,
		}),
		Entry("basic http_connection_manager with tcp access log", testCase{
			listenerAddress: "127.0.0.1",
			listenerPort:    27070,
			statsName:       "backend",
			routeName:       "outbound:backend",
			backend: &mesh_proto.LoggingBackend{
				Name: "tcp",
				Format: `[%START_TIME%] "%REQ(x-request-id)%" "%REQ(:authority)%" "%REQ(origin)%" "%REQ(content-type)%" "%KUMA_SOURCE_SERVICE%" "%KUMA_DESTINATION_SERVICE%" "%KUMA_SOURCE_ADDRESS%" "%KUMA_SOURCE_ADDRESS_WITHOUT_PORT%" "%UPSTREAM_HOST%"
"%RESP(server):5%" "%TRAILER(grpc-message):7%" "DYNAMIC_METADATA(namespace:object:key):9" "FILTER_STATE(filter.state.key):12"
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
                  internalAddressConfig:
                    cidrRanges:
                      - addressPrefix: 127.0.0.1
                        prefixLen: 32
                      - addressPrefix: ::1
                        prefixLen: 128
            name: outbound:127.0.0.1:27070
            trafficDirection: OUTBOUND`,
		}),
	)
})
