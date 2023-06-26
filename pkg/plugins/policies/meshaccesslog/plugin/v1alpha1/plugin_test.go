package v1alpha1_test

import (
	"fmt"
	"path/filepath"
	"strings"

	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/xds"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	core_rules "github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	policies_xds "github.com/kumahq/kuma/pkg/plugins/policies/core/xds"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshaccesslog/api/v1alpha1"
	plugin "github.com/kumahq/kuma/pkg/plugins/policies/meshaccesslog/plugin/v1alpha1"
	gateway_plugin "github.com/kumahq/kuma/pkg/plugins/runtime/gateway"
	"github.com/kumahq/kuma/pkg/test/matchers"
	"github.com/kumahq/kuma/pkg/test/resources/builders"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
	test_xds "github.com/kumahq/kuma/pkg/test/xds"
	"github.com/kumahq/kuma/pkg/util/pointer"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	. "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
	"github.com/kumahq/kuma/pkg/xds/generator"
)

var _ = Describe("MeshAccessLog", func() {
	BeforeEach(func() {
		core.TempDir = func() string {
			return "/tmp"
		}
	})

	type sidecarTestCase struct {
		resources         []core_xds.Resource
		outbounds         []*mesh_proto.Dataplane_Networking_Outbound
		toRules           core_rules.ToRules
		fromRules         core_rules.FromRules
		expectedListeners []string
		expectedClusters  []string
	}
	DescribeTable("should generate proper Envoy config",
		func(given sidecarTestCase) {
			resourceSet := core_xds.NewResourceSet()
			for _, res := range given.resources {
				r := res
				resourceSet.Add(&r)
			}

			context := xds_context.Context{
				Mesh: xds_context.MeshContext{
					Resource: &core_mesh.MeshResource{
						Meta: &test_model.ResourceMeta{
							Name: "default",
						},
					},
				},
			}
			proxy := xds.Proxy{
				APIVersion: envoy_common.APIV3,
				Dataplane: &core_mesh.DataplaneResource{
					Meta: &test_model.ResourceMeta{
						Mesh: "default",
						Name: "backend",
					},
					Spec: &mesh_proto.Dataplane{
						Networking: &mesh_proto.Dataplane_Networking{
							Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
								{
									Tags: map[string]string{
										mesh_proto.ServiceTag: "backend",
									},
									Address: "127.0.0.1",
									Port:    17777,
								},
							},
							Outbound: append([]*mesh_proto.Dataplane_Networking_Outbound{
								{
									Address: "127.0.0.1",
									Port:    27777,
									Tags: map[string]string{
										mesh_proto.ServiceTag: "other-service",
									},
								},
							},
								given.outbounds...,
							),
						},
					},
				},
				Policies: xds.MatchedPolicies{
					Dynamic: map[core_model.ResourceType]xds.TypedMatchingPolicies{
						api.MeshAccessLogType: {
							Type:      api.MeshAccessLogType,
							ToRules:   given.toRules,
							FromRules: given.fromRules,
						},
					},
				},
			}
			plugin := plugin.NewPlugin().(core_plugins.PolicyPlugin)

			Expect(plugin.Apply(resourceSet, context, &proxy)).To(Succeed())
			policies_xds.ResourceArrayShouldEqual(resourceSet.ListOf(envoy_resource.ListenerType), given.expectedListeners)
			policies_xds.ResourceArrayShouldEqual(resourceSet.ListOf(envoy_resource.ClusterType), given.expectedClusters)
		},
		Entry("basic outbound route", sidecarTestCase{
			resources: []core_xds.Resource{{
				Name:   "outbound",
				Origin: generator.OriginOutbound,
				Resource: NewListenerBuilder(envoy_common.APIV3).
					Configure(OutboundListener("outbound:127.0.0.1:27777", "127.0.0.1", 27777, core_xds.SocketAddressProtocolTCP)).
					Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3).
						Configure(HttpConnectionManager("127.0.0.1:27777", false)).
						Configure(
							HttpOutboundRoute(
								"backend",
								envoy_common.Routes{{
									Clusters: []envoy_common.Cluster{envoy_common.NewCluster(
										envoy_common.WithService("backend"),
										envoy_common.WithWeight(100),
									)},
								}},
								map[string]map[string]bool{
									"kuma.io/service": {
										"web": true,
									},
								},
							),
						),
					)).MustBuild(),
			}},
			toRules: core_rules.ToRules{
				Rules: []*core_rules.Rule{
					{
						Subset: core_rules.Subset{},
						Conf: api.Conf{
							Backends: &[]api.Backend{{
								File: &api.FileBackend{
									Path: "/tmp/log",
								},
							}},
						},
					},
				},
			},
			expectedListeners: []string{
				`
            address:
              socketAddress:
                address: 127.0.0.1
                portValue: 27777
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
                              inlineString: |
                                [%START_TIME%] default "%REQ(:method)% %REQ(x-envoy-original-path?:path)% %PROTOCOL%" %RESPONSE_CODE% %RESPONSE_FLAGS% %BYTES_RECEIVED% %BYTES_SENT% %DURATION% %RESP(x-envoy-upstream-service-time)% "%REQ(x-forwarded-for)%" "%REQ(user-agent)%" "%REQ(x-b3-traceid?x-datadog-traceid)%" "%REQ(x-request-id)%" "%REQ(:authority)%" "backend" "other-service" "" "%UPSTREAM_HOST%"
                      path: /tmp/log
                  httpFilters:
                  - name: envoy.filters.http.router
                    typedConfig:
                      '@type': type.googleapis.com/envoy.extensions.filters.http.router.v3.Router
                  routeConfig:
                    name: outbound:backend
                    validateClusters: false
                    requestHeadersToAdd:
                    - header:
                        key: x-kuma-tags
                        value: '&kuma.io/service=web&'
                    virtualHosts:
                    - domains:
                      - '*'
                      name: backend
                      routes:
                      - match:
                          prefix: /
                        route:
                          cluster: backend
                          timeout: 0s
                  statPrefix: "127_0_0_1_27777"
            name: outbound:127.0.0.1:27777
            trafficDirection: OUTBOUND`,
			},
		}),
		Entry("outbound tcpproxy with file backend and default format", sidecarTestCase{
			resources: []core_xds.Resource{{
				Name:   "outbound",
				Origin: generator.OriginOutbound,
				Resource: NewListenerBuilder(envoy_common.APIV3).
					Configure(OutboundListener("outbound:127.0.0.1:27777", "127.0.0.1", 27777, core_xds.SocketAddressProtocolTCP)).
					Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3).
						Configure(TcpProxyDeprecated(
							"127.0.0.1:27777",
							envoy_common.NewCluster(
								envoy_common.WithService("backend"),
								envoy_common.WithWeight(100),
							),
						)),
					)).MustBuild(),
			}},
			toRules: core_rules.ToRules{
				Rules: []*core_rules.Rule{
					{
						Subset: core_rules.Subset{},
						Conf: api.Conf{
							Backends: &[]api.Backend{{
								File: &api.FileBackend{
									Path: "/tmp/log",
								},
							}},
						},
					},
				},
			},
			expectedListeners: []string{
				`
            address:
              socketAddress:
                address: 127.0.0.1
                portValue: 27777
            filterChains:
                - filters:
                    - name: envoy.filters.network.tcp_proxy
                      typedConfig:
                        '@type': type.googleapis.com/envoy.extensions.filters.network.tcp_proxy.v3.TcpProxy
                        accessLog:
                            - name: envoy.access_loggers.file
                              typedConfig:
                                '@type': type.googleapis.com/envoy.extensions.access_loggers.file.v3.FileAccessLog
                                logFormat:
                                    textFormatSource:
                                        inlineString: |
                                            [%START_TIME%] %RESPONSE_FLAGS% default (backend)->%UPSTREAM_HOST%(other-service) took %DURATION%ms, sent %BYTES_SENT% bytes, received: %BYTES_RECEIVED% bytes
                                path: /tmp/log
                        cluster: backend
                        statPrefix: "127_0_0_1_27777"
            name: outbound:127.0.0.1:27777
            trafficDirection: OUTBOUND`,
			},
		}),
		Entry("outbound tcpproxy with file backend and plain format", sidecarTestCase{
			resources: []core_xds.Resource{{
				Name:   "outbound",
				Origin: generator.OriginOutbound,
				Resource: NewListenerBuilder(envoy_common.APIV3).
					Configure(OutboundListener("outbound:127.0.0.1:27777", "127.0.0.1", 27777, core_xds.SocketAddressProtocolTCP)).
					Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3).
						Configure(TcpProxyDeprecated(
							"127.0.0.1:27777",
							envoy_common.NewCluster(
								envoy_common.WithService("backend"),
								envoy_common.WithWeight(100),
							),
						)),
					)).MustBuild(),
			}},
			toRules: core_rules.ToRules{
				Rules: []*core_rules.Rule{
					{
						Subset: core_rules.Subset{},
						Conf: api.Conf{
							Backends: &[]api.Backend{{
								File: &api.FileBackend{
									Path: "/tmp/log",
									Format: &api.Format{
										Plain: pointer.To("custom format [%START_TIME%] %RESPONSE_FLAGS%"),
									},
								},
							}},
						},
					},
				},
			},
			expectedListeners: []string{
				`
            address:
              socketAddress:
                address: 127.0.0.1
                portValue: 27777
            filterChains:
                - filters:
                    - name: envoy.filters.network.tcp_proxy
                      typedConfig:
                        '@type': type.googleapis.com/envoy.extensions.filters.network.tcp_proxy.v3.TcpProxy
                        accessLog:
                            - name: envoy.access_loggers.file
                              typedConfig:
                                '@type': type.googleapis.com/envoy.extensions.access_loggers.file.v3.FileAccessLog
                                logFormat:
                                    textFormatSource:
                                        inlineString: |
                                            custom format [%START_TIME%] %RESPONSE_FLAGS%
                                path: /tmp/log
                        cluster: backend
                        statPrefix: "127_0_0_1_27777"
            name: outbound:127.0.0.1:27777
            trafficDirection: OUTBOUND`,
			},
		}),
		Entry("outbound tcpproxy with file backend and json format", sidecarTestCase{
			resources: []core_xds.Resource{{
				Name:   "outbound",
				Origin: generator.OriginOutbound,
				Resource: NewListenerBuilder(envoy_common.APIV3).
					Configure(OutboundListener("outbound:127.0.0.1:27777", "127.0.0.1", 27777, core_xds.SocketAddressProtocolTCP)).
					Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3).
						Configure(TcpProxyDeprecated(
							"127.0.0.1:27777",
							envoy_common.NewCluster(
								envoy_common.WithService("backend"),
								envoy_common.WithWeight(100),
							),
						)),
					)).MustBuild(),
			}},
			toRules: core_rules.ToRules{
				Rules: []*core_rules.Rule{
					{
						Subset: core_rules.Subset{},
						Conf: api.Conf{
							Backends: &[]api.Backend{{
								File: &api.FileBackend{
									Path: "/tmp/log",
									Format: &api.Format{
										Json: pointer.To([]api.JsonValue{
											{Key: "protocol", Value: "%PROTOCOL%"},
											{Key: "duration", Value: "%DURATION%"},
										}),
									},
								},
							}},
						},
					},
				},
			},
			expectedListeners: []string{
				`
            address:
              socketAddress:
                address: 127.0.0.1
                portValue: 27777
            filterChains:
                - filters:
                    - name: envoy.filters.network.tcp_proxy
                      typedConfig:
                        '@type': type.googleapis.com/envoy.extensions.filters.network.tcp_proxy.v3.TcpProxy
                        accessLog:
                            - name: envoy.access_loggers.file
                              typedConfig:
                                '@type': type.googleapis.com/envoy.extensions.access_loggers.file.v3.FileAccessLog
                                logFormat:
                                    jsonFormat:
                                      duration: '%DURATION%'
                                      protocol: '%PROTOCOL%'
                                path: /tmp/log
                        cluster: backend
                        statPrefix: "127_0_0_1_27777"
            name: outbound:127.0.0.1:27777
            trafficDirection: OUTBOUND`,
			},
		}),
		Entry("outbound tcpproxy with tcp backend and default format", sidecarTestCase{
			resources: []core_xds.Resource{{
				Name:   "outbound",
				Origin: generator.OriginOutbound,
				Resource: NewListenerBuilder(envoy_common.APIV3).
					Configure(OutboundListener("outbound:127.0.0.1:27777", "127.0.0.1", 27777, core_xds.SocketAddressProtocolTCP)).
					Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3).
						Configure(TcpProxyDeprecated(
							"127.0.0.1:27777",
							envoy_common.NewCluster(
								envoy_common.WithService("backend"),
								envoy_common.WithWeight(100),
							),
						)),
					)).MustBuild(),
			}},
			toRules: core_rules.ToRules{
				Rules: []*core_rules.Rule{
					{
						Subset: core_rules.Subset{},
						Conf: api.Conf{
							Backends: &[]api.Backend{{
								Tcp: &api.TCPBackend{
									Address: "logging.backend",
								},
							}},
						},
					},
				},
			},
			expectedListeners: []string{
				`
            address:
              socketAddress:
                address: 127.0.0.1
                portValue: 27777
            filterChains:
                - filters:
                    - name: envoy.filters.network.tcp_proxy
                      typedConfig:
                        '@type': type.googleapis.com/envoy.extensions.filters.network.tcp_proxy.v3.TcpProxy
                        accessLog:
                            - name: envoy.access_loggers.file
                              typedConfig:
                                '@type': type.googleapis.com/envoy.extensions.access_loggers.file.v3.FileAccessLog
                                logFormat:
                                    jsonFormat:
                                        address: logging.backend
                                        message: |
                                            [%START_TIME%] %RESPONSE_FLAGS% default (backend)->%UPSTREAM_HOST%(other-service) took %DURATION%ms, sent %BYTES_SENT% bytes, received: %BYTES_RECEIVED% bytes
                                path: /tmp/kuma-al-backend-default.sock
                        cluster: backend
                        statPrefix: "127_0_0_1_27777"
            name: outbound:127.0.0.1:27777
            trafficDirection: OUTBOUND`,
			},
		}),
		Entry("outbound tcpproxy with opentelemetry backend and plain format", sidecarTestCase{
			resources: []core_xds.Resource{{
				Name:   "other-service",
				Origin: generator.OriginOutbound,
				Resource: NewListenerBuilder(envoy_common.APIV3).
					Configure(OutboundListener("outbound:127.0.0.1:27777", "127.0.0.1", 27777, core_xds.SocketAddressProtocolTCP)).
					Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3).
						Configure(TcpProxyDeprecated(
							"127.0.0.1:27777",
							envoy_common.NewCluster(
								envoy_common.WithService("other-service"),
								envoy_common.WithWeight(100),
							),
						)),
					)).MustBuild(),
			}, {
				Name:   "foo",
				Origin: generator.OriginOutbound,
				Resource: NewListenerBuilder(envoy_common.APIV3).
					Configure(OutboundListener("outbound:127.0.0.1:27778", "127.0.0.1", 27778, core_xds.SocketAddressProtocolTCP)).
					Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3).
						Configure(TcpProxyDeprecated(
							"127.0.0.1:27778",
							envoy_common.NewCluster(
								envoy_common.WithService("foo-service"),
								envoy_common.WithWeight(100),
							),
						)),
					)).MustBuild(),
			}, {
				Name:   "bar",
				Origin: generator.OriginOutbound,
				Resource: NewListenerBuilder(envoy_common.APIV3).
					Configure(OutboundListener("outbound:127.0.0.1:27779", "127.0.0.1", 27779, core_xds.SocketAddressProtocolTCP)).
					Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3).
						Configure(TcpProxyDeprecated(
							"127.0.0.1:27779",
							envoy_common.NewCluster(
								envoy_common.WithService("bar-service"),
								envoy_common.WithWeight(100),
							),
						)),
					)).MustBuild(),
			}},
			outbounds: []*mesh_proto.Dataplane_Networking_Outbound{{
				Address: "127.0.0.1",
				Port:    27778,
				Tags: map[string]string{
					mesh_proto.ServiceTag: "foo-service",
				},
			}, {
				Address: "127.0.0.1",
				Port:    27779,
				Tags: map[string]string{
					mesh_proto.ServiceTag: "bar-service",
				},
			}},
			toRules: core_rules.ToRules{
				Rules: []*core_rules.Rule{
					{
						Subset: core_rules.Subset{{
							Key:   mesh_proto.ServiceTag,
							Value: "other-service",
						}},
						Conf: api.Conf{
							Backends: &[]api.Backend{{
								OpenTelemetry: &api.OtelBackend{
									Endpoint: "otel-collector",
								},
							}},
						},
					},
					{
						Subset: core_rules.Subset{{
							Key:   mesh_proto.ServiceTag,
							Value: "foo-service",
						}},
						Conf: api.Conf{
							Backends: &[]api.Backend{{
								OpenTelemetry: &api.OtelBackend{
									Endpoint: "otel-collector",
									Body: &apiextensionsv1.JSON{
										Raw: []byte("%KUMA_MESH%"),
									},
								},
							}},
						},
					},
					{
						Subset: core_rules.Subset{{
							Key:   mesh_proto.ServiceTag,
							Value: "bar-service",
						}},
						Conf: api.Conf{
							Backends: &[]api.Backend{{
								OpenTelemetry: &api.OtelBackend{
									Endpoint: "other-otel-collector:5317",
									Body: &apiextensionsv1.JSON{
										Raw: []byte(`{
										  "kvlistValue": {
											"values": [
											  {"key": "mesh", "value": {"stringValue": "%KUMA_MESH%"}}
											]
										  }
									    }`),
									},
								},
							}},
						},
					},
				},
			},
			expectedClusters: []string{`
            altStatName: meshaccesslog_opentelemetry_0
            connectTimeout: 10s
            dnsLookupFamily: V4_ONLY
            loadAssignment:
                clusterName: meshaccesslog:opentelemetry:0
                endpoints:
                    - lbEndpoints:
                        - endpoint:
                            address:
                                socketAddress:
                                    address: otel-collector
                                    portValue: 4317
            name: meshaccesslog:opentelemetry:0
            type: STRICT_DNS
            typedExtensionProtocolOptions:
                envoy.extensions.upstreams.http.v3.HttpProtocolOptions:
                    '@type': type.googleapis.com/envoy.extensions.upstreams.http.v3.HttpProtocolOptions
                    explicitHttpConfig:
                        http2ProtocolOptions: {}
            `, `
            altStatName: meshaccesslog_opentelemetry_1
            connectTimeout: 10s
            dnsLookupFamily: V4_ONLY
            loadAssignment:
                clusterName: meshaccesslog:opentelemetry:1
                endpoints:
                    - lbEndpoints:
                        - endpoint:
                            address:
                                socketAddress:
                                    address: other-otel-collector
                                    portValue: 5317
            name: meshaccesslog:opentelemetry:1
            type: STRICT_DNS
            typedExtensionProtocolOptions:
                envoy.extensions.upstreams.http.v3.HttpProtocolOptions:
                    '@type': type.googleapis.com/envoy.extensions.upstreams.http.v3.HttpProtocolOptions
                    explicitHttpConfig:
                        http2ProtocolOptions: {}
            `},
			expectedListeners: []string{
				`
            address:
              socketAddress:
                address: 127.0.0.1
                portValue: 27779
            filterChains:
                - filters:
                    - name: envoy.filters.network.tcp_proxy
                      typedConfig:
                        '@type': type.googleapis.com/envoy.extensions.filters.network.tcp_proxy.v3.TcpProxy
                        accessLog:
                            - name: envoy.access_loggers.open_telemetry
                              typedConfig:
                                '@type': type.googleapis.com/envoy.extensions.access_loggers.open_telemetry.v3.OpenTelemetryAccessLogConfig
                                body:
                                    kvlistValue:
                                        values:
                                            - key: mesh
                                              value:
                                                  stringValue: default
                                attributes: {}
                                commonConfig:
                                    grpcService:
                                        envoyGrpc:
                                            clusterName: meshaccesslog:opentelemetry:1
                                    logName: MeshAccessLog
                                    transportApiVersion: V3
                        cluster: bar-service
                        statPrefix: "127_0_0_1_27779"
            name: outbound:127.0.0.1:27779
            trafficDirection: OUTBOUND`, `
            address:
              socketAddress:
                address: 127.0.0.1
                portValue: 27778
            filterChains:
                - filters:
                    - name: envoy.filters.network.tcp_proxy
                      typedConfig:
                        '@type': type.googleapis.com/envoy.extensions.filters.network.tcp_proxy.v3.TcpProxy
                        accessLog:
                            - name: envoy.access_loggers.open_telemetry
                              typedConfig:
                                '@type': type.googleapis.com/envoy.extensions.access_loggers.open_telemetry.v3.OpenTelemetryAccessLogConfig
                                body:
                                    stringValue: default
                                attributes: {}
                                commonConfig:
                                    grpcService:
                                        envoyGrpc:
                                            clusterName: meshaccesslog:opentelemetry:0
                                    logName: MeshAccessLog
                                    transportApiVersion: V3
                        cluster: foo-service
                        statPrefix: "127_0_0_1_27778"
            name: outbound:127.0.0.1:27778
            trafficDirection: OUTBOUND`, `
            address:
              socketAddress:
                address: 127.0.0.1
                portValue: 27777
            filterChains:
                - filters:
                    - name: envoy.filters.network.tcp_proxy
                      typedConfig:
                        '@type': type.googleapis.com/envoy.extensions.filters.network.tcp_proxy.v3.TcpProxy
                        accessLog:
                            - name: envoy.access_loggers.open_telemetry
                              typedConfig:
                                '@type': type.googleapis.com/envoy.extensions.access_loggers.open_telemetry.v3.OpenTelemetryAccessLogConfig
                                body:
                                    stringValue: '[%START_TIME%] %RESPONSE_FLAGS% default (backend)->%UPSTREAM_HOST%(other-service) took %DURATION%ms, sent %BYTES_SENT% bytes, received: %BYTES_RECEIVED% bytes'
                                attributes: {}
                                commonConfig:
                                    grpcService:
                                        envoyGrpc:
                                            clusterName: meshaccesslog:opentelemetry:0
                                    logName: MeshAccessLog
                                    transportApiVersion: V3
                        cluster: other-service
                        statPrefix: "127_0_0_1_27777"
            name: outbound:127.0.0.1:27777
            trafficDirection: OUTBOUND`,
			},
		}),
		Entry("outbound tcpproxy with tcp backend and plain format", sidecarTestCase{
			resources: []core_xds.Resource{{
				Name:   "outbound",
				Origin: generator.OriginOutbound,
				Resource: NewListenerBuilder(envoy_common.APIV3).
					Configure(OutboundListener("outbound:127.0.0.1:27777", "127.0.0.1", 27777, core_xds.SocketAddressProtocolTCP)).
					Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3).
						Configure(TcpProxyDeprecated(
							"127.0.0.1:27777",
							envoy_common.NewCluster(
								envoy_common.WithService("backend"),
								envoy_common.WithWeight(100),
							),
						)),
					)).MustBuild(),
			}},
			toRules: core_rules.ToRules{
				Rules: []*core_rules.Rule{
					{
						Subset: core_rules.Subset{},
						Conf: api.Conf{
							Backends: &[]api.Backend{{
								Tcp: &api.TCPBackend{
									Address: "logging.backend",
									Format: &api.Format{
										Plain: pointer.To("custom format [%START_TIME%] %RESPONSE_FLAGS%"),
									},
								},
							}},
						},
					},
				},
			},
			expectedListeners: []string{
				`
            address:
              socketAddress:
                address: 127.0.0.1
                portValue: 27777
            filterChains:
                - filters:
                    - name: envoy.filters.network.tcp_proxy
                      typedConfig:
                        '@type': type.googleapis.com/envoy.extensions.filters.network.tcp_proxy.v3.TcpProxy
                        accessLog:
                            - name: envoy.access_loggers.file
                              typedConfig:
                                '@type': type.googleapis.com/envoy.extensions.access_loggers.file.v3.FileAccessLog
                                logFormat:
                                    jsonFormat:
                                        address: logging.backend
                                        message: |
                                            custom format [%START_TIME%] %RESPONSE_FLAGS%
                                path: /tmp/kuma-al-backend-default.sock
                        cluster: backend
                        statPrefix: "127_0_0_1_27777"
            name: outbound:127.0.0.1:27777
            trafficDirection: OUTBOUND`,
			},
		}),
		Entry("outbound tcpproxy with tcp backend and json format", sidecarTestCase{
			resources: []core_xds.Resource{{
				Name:   "outbound",
				Origin: generator.OriginOutbound,
				Resource: NewListenerBuilder(envoy_common.APIV3).
					Configure(OutboundListener("outbound:127.0.0.1:27777", "127.0.0.1", 27777, core_xds.SocketAddressProtocolTCP)).
					Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3).
						Configure(TcpProxyDeprecated(
							"127.0.0.1:27777",
							envoy_common.NewCluster(
								envoy_common.WithService("backend"),
								envoy_common.WithWeight(100),
							),
						)),
					)).MustBuild(),
			}},
			toRules: core_rules.ToRules{
				Rules: []*core_rules.Rule{
					{
						Subset: core_rules.Subset{},
						Conf: api.Conf{
							Backends: &[]api.Backend{{
								Tcp: &api.TCPBackend{
									Address: "logging.backend",
									Format: &api.Format{
										Json: pointer.To([]api.JsonValue{
											{Key: "protocol", Value: "%PROTOCOL%"},
											{Key: "duration", Value: "%DURATION%"},
										}),
									},
								},
							}},
						},
					},
				},
			},
			expectedListeners: []string{
				`
            address:
              socketAddress:
                address: 127.0.0.1
                portValue: 27777
            filterChains:
                - filters:
                    - name: envoy.filters.network.tcp_proxy
                      typedConfig:
                        '@type': type.googleapis.com/envoy.extensions.filters.network.tcp_proxy.v3.TcpProxy
                        accessLog:
                            - name: envoy.access_loggers.file
                              typedConfig:
                                '@type': type.googleapis.com/envoy.extensions.access_loggers.file.v3.FileAccessLog
                                logFormat:
                                    jsonFormat:
                                        address: logging.backend
                                        message: 
                                            duration: '%DURATION%'
                                            protocol: '%PROTOCOL%'
                                path: /tmp/kuma-al-backend-default.sock
                        cluster: backend
                        statPrefix: "127_0_0_1_27777"
            name: outbound:127.0.0.1:27777
            trafficDirection: OUTBOUND`,
			},
		}),
		Entry("basic outbound route without match", sidecarTestCase{
			resources: []core_xds.Resource{{
				Name:   "outbound",
				Origin: generator.OriginOutbound,
				Resource: NewListenerBuilder(envoy_common.APIV3).
					Configure(OutboundListener("outbound:127.0.0.1:27777", "127.0.0.1", 27777, core_xds.SocketAddressProtocolTCP)).
					Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3).
						Configure(HttpConnectionManager("127.0.0.1:27777", false)).
						Configure(
							HttpOutboundRoute(
								"backend",
								envoy_common.Routes{
									{
										Clusters: []envoy_common.Cluster{envoy_common.NewCluster(
											envoy_common.WithService("backend"),
											envoy_common.WithWeight(100),
										)},
									},
								},
								map[string]map[string]bool{
									"kuma.io/service": {
										"web": true,
									},
								},
							),
						),
					)).MustBuild(),
			}},
			toRules: core_rules.ToRules{
				Rules: []*core_rules.Rule{
					{
						Subset: core_rules.Subset{{
							Key:   mesh_proto.ServiceTag,
							Value: "other",
						}},
						Conf: api.Conf{
							Backends: &[]api.Backend{{
								File: &api.FileBackend{
									Path: "/tmp/log",
								},
							}},
						},
					},
				},
			},
			expectedListeners: []string{
				`
            address:
              socketAddress:
                address: 127.0.0.1
                portValue: 27777
            filterChains:
            - filters:
              - name: envoy.filters.network.http_connection_manager
                typedConfig:
                  '@type': type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager
                  httpFilters:
                  - name: envoy.filters.http.router
                    typedConfig:
                      '@type': type.googleapis.com/envoy.extensions.filters.http.router.v3.Router
                  routeConfig:
                    name: outbound:backend
                    validateClusters: false
                    requestHeadersToAdd:
                    - header:
                        key: x-kuma-tags
                        value: '&kuma.io/service=web&'
                    virtualHosts:
                    - domains:
                      - '*'
                      name: backend
                      routes:
                      - match:
                          prefix: /
                        route:
                          cluster: backend
                          timeout: 0s
                  statPrefix: "127_0_0_1_27777"
            name: outbound:127.0.0.1:27777
            trafficDirection: OUTBOUND`,
			},
		}),
		Entry("basic inbound route", sidecarTestCase{
			resources: []core_xds.Resource{{
				Name:   "inbound",
				Origin: generator.OriginInbound,
				Resource: NewListenerBuilder(envoy_common.APIV3).
					Configure(InboundListener("inbound:127.0.0.1:17777", "127.0.0.1", 17777, core_xds.SocketAddressProtocolTCP)).
					Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3).
						Configure(HttpConnectionManager("127.0.0.1:17777", false)).
						Configure(
							HttpInboundRoutes(
								"backend",
								envoy_common.Routes{
									{
										Clusters: []envoy_common.Cluster{envoy_common.NewCluster(
											envoy_common.WithService("backend"),
											envoy_common.WithWeight(100),
										)},
									},
								},
							),
						),
					)).MustBuild(),
			}},
			fromRules: core_rules.FromRules{
				Rules: map[core_rules.InboundListener]core_rules.Rules{
					{Address: "127.0.0.1", Port: 17777}: {{
						Subset: core_rules.Subset{},
						Conf: api.Conf{
							Backends: &[]api.Backend{{
								File: &api.FileBackend{
									Path: "/tmp/log",
								},
							}},
						},
					}},
				},
			},
			expectedListeners: []string{
				`
            address:
              socketAddress:
                address: 127.0.0.1
                portValue: 17777
            enableReusePort: false
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
                              inlineString: |
                                [%START_TIME%] default "%REQ(:method)% %REQ(x-envoy-original-path?:path)% %PROTOCOL%" %RESPONSE_CODE% %RESPONSE_FLAGS% %BYTES_RECEIVED% %BYTES_SENT% %DURATION% %RESP(x-envoy-upstream-service-time)% "%REQ(x-forwarded-for)%" "%REQ(user-agent)%" "%REQ(x-b3-traceid?x-datadog-traceid)%" "%REQ(x-request-id)%" "%REQ(:authority)%" "unknown" "backend" "" "%UPSTREAM_HOST%"
                      path: /tmp/log
                  httpFilters:
                  - name: envoy.filters.http.router
                    typedConfig:
                      '@type': type.googleapis.com/envoy.extensions.filters.http.router.v3.Router
                  routeConfig:
                    name: inbound:backend
                    validateClusters: false
                    requestHeadersToRemove:
                    - x-kuma-tags
                    virtualHosts:
                    - domains:
                      - '*'
                      name: backend
                      routes:
                      - match:
                          prefix: /
                        route:
                          cluster: backend
                          timeout: 0s
                  statPrefix: "127_0_0_1_17777"
            name: inbound:127.0.0.1:17777
            trafficDirection: INBOUND`,
			},
		}),
	)
	type gatewayTestCase struct {
		routes  []*core_mesh.MeshGatewayRouteResource
		toRules core_rules.ToRules
	}
	DescribeTable("should generate proper Envoy config for MeshGateway Dataplanes",
		func(given gatewayTestCase) {
			gateways := core_mesh.MeshGatewayResourceList{
				Items: []*core_mesh.MeshGatewayResource{{
					Meta: &test_model.ResourceMeta{Name: "gateway", Mesh: "default"},
					Spec: &mesh_proto.MeshGateway{
						Selectors: []*mesh_proto.Selector{
							{
								Match: map[string]string{
									mesh_proto.ServiceTag: "gateway",
								},
							},
						},
						Conf: &mesh_proto.MeshGateway_Conf{
							Listeners: []*mesh_proto.MeshGateway_Listener{
								{
									Protocol: mesh_proto.MeshGateway_Listener_HTTP,
									Port:     8080,
								},
							},
						},
					},
				}},
			}
			resources := xds_context.NewResources()
			resources.MeshLocalResources[core_mesh.MeshGatewayType] = &gateways
			resources.MeshLocalResources[core_mesh.MeshGatewayRouteType] = &core_mesh.MeshGatewayRouteResourceList{
				Items: given.routes,
			}

			context := test_xds.CreateSampleMeshContextWith(resources)
			proxy := xds.Proxy{
				APIVersion: "v3",
				Dataplane: &core_mesh.DataplaneResource{
					Meta: &test_model.ResourceMeta{
						Mesh: "default",
						Name: "gateway",
					},
					Spec: &mesh_proto.Dataplane{
						Networking: &mesh_proto.Dataplane_Networking{
							Address: "127.0.0.1",
							Gateway: &mesh_proto.Dataplane_Networking_Gateway{
								Tags: map[string]string{
									mesh_proto.ServiceTag: "gateway",
								},
								Type: mesh_proto.Dataplane_Networking_Gateway_BUILTIN,
							},
						},
					},
				},
				Policies: xds.MatchedPolicies{
					Dynamic: map[core_model.ResourceType]xds.TypedMatchingPolicies{
						api.MeshAccessLogType: {
							Type:    api.MeshAccessLogType,
							ToRules: given.toRules,
						},
					},
				},
			}

			gatewayGenerator := gateway_plugin.NewGenerator("test-zone")
			generatedResources, err := gatewayGenerator.Generate(context, &proxy)
			Expect(err).NotTo(HaveOccurred())

			plugin := plugin.NewPlugin().(core_plugins.PolicyPlugin)
			Expect(plugin.Apply(generatedResources, context, &proxy)).To(Succeed())

			nameSplit := strings.Split(GinkgoT().Name(), " ")
			name := nameSplit[len(nameSplit)-1]

			Expect(getResourceYaml(generatedResources.ListOf(envoy_resource.ListenerType))).To(matchers.MatchGoldenYAML(filepath.Join("testdata", fmt.Sprintf("%s.gateway.listener.golden.yaml", name))))
			Expect(getResourceYaml(generatedResources.ListOf(envoy_resource.ClusterType))).To(matchers.MatchGoldenYAML(filepath.Join("testdata", fmt.Sprintf("%s.gateway.cluster.golden.yaml", name))))
			Expect(getResourceYaml(generatedResources.ListOf(envoy_resource.RouteType))).To(matchers.MatchGoldenYAML(filepath.Join("testdata", fmt.Sprintf("%s.gateway.route.golden.yaml", name))))
		},
		Entry("basic-gateway", gatewayTestCase{
			routes: []*core_mesh.MeshGatewayRouteResource{
				builders.GatewayRoute().
					WithName("sample-gateway-route").
					WithGateway("gateway").
					WithExactMatchHttpRoute("/", "backend", "other-service").
					Build(),
			},
			toRules: core_rules.ToRules{
				Rules: []*core_rules.Rule{
					{
						Subset: core_rules.Subset{},
						Conf: api.Conf{
							Backends: &[]api.Backend{{
								File: &api.FileBackend{
									Path: "/tmp/log",
								},
							}},
						},
					},
				},
			},
		}),
	)
})

func getResourceYaml(list core_xds.ResourceList) []byte {
	actualResource, err := util_proto.ToYAML(list[0].Resource)
	Expect(err).ToNot(HaveOccurred())
	return actualResource
}
