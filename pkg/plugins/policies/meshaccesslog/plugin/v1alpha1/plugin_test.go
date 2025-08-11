package v1alpha1_test

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
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
	"github.com/kumahq/kuma/pkg/test/resources/samples"
	xds_builders "github.com/kumahq/kuma/pkg/test/xds/builders"
	xds_samples "github.com/kumahq/kuma/pkg/test/xds/samples"
	"github.com/kumahq/kuma/pkg/util/pointer"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	. "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
	"github.com/kumahq/kuma/pkg/xds/generator"
)

var _ = Describe("MeshAccessLog", func() {
	type sidecarTestCase struct {
		resources         []core_xds.Resource
		outbounds         []*builders.OutboundBuilder
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

			xdsCtx := xds_samples.SampleContext()
			proxy := xds_builders.Proxy().
				WithMetadata(&xds.DataplaneMetadata{
					AccessLogSocketPath: "/tmp/kuma-al-backend-default.sock",
				}).
				WithDataplane(
					builders.Dataplane().
						WithName("backend").
						WithMesh("default").
						AddInbound(builders.Inbound().
							WithService("backend").
							WithAddress("127.0.0.1").
							WithPort(17777),
						).
						AddOutbound(builders.Outbound().
							WithService("other-service").
							WithAddress("127.0.0.1").
							WithPort(27777),
						).
						AddOutbounds(given.outbounds),
				).
				WithPolicies(
					xds_builders.MatchedPolicies().WithPolicy(api.MeshAccessLogType, given.toRules, given.fromRules),
				).
				WithInternalAddresses(core_xds.InternalAddress{AddressPrefix: "172.16.0.0", PrefixLen: 12}, core_xds.InternalAddress{AddressPrefix: "fc00::", PrefixLen: 7}).
				Build()
			plugin := plugin.NewPlugin().(core_plugins.PolicyPlugin)

			Expect(plugin.Apply(resourceSet, xdsCtx, proxy)).To(Succeed())
			policies_xds.ResourceArrayShouldEqual(resourceSet.ListOf(envoy_resource.ListenerType), given.expectedListeners)
			policies_xds.ResourceArrayShouldEqual(resourceSet.ListOf(envoy_resource.ClusterType), given.expectedClusters)
		},
		Entry("basic outbound route", sidecarTestCase{
			resources: []core_xds.Resource{{
				Name:   "outbound",
				Origin: generator.OriginOutbound,
				Resource: NewOutboundListenerBuilder(envoy_common.APIV3, "127.0.0.1", 27777, core_xds.SocketAddressProtocolTCP).
					Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
						Configure(HttpConnectionManager("127.0.0.1:27777", false, nil)).
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
                                [%START_TIME%] default "%REQ(:METHOD)% %REQ(X-ENVOY-ORIGINAL-PATH?:PATH)% %PROTOCOL%" %RESPONSE_CODE% %RESPONSE_FLAGS% %BYTES_RECEIVED% %BYTES_SENT% %DURATION% %RESP(X-ENVOY-UPSTREAM-SERVICE-TIME)% "%REQ(X-FORWARDED-FOR)%" "%REQ(USER-AGENT)%" "%REQ(X-B3-TRACEID?X-DATADOG-TRACEID)%" "%REQ(X-REQUEST-ID)%" "%REQ(:AUTHORITY)%" "backend" "other-service" "127.0.0.1" "%UPSTREAM_HOST%"
                      path: /tmp/log
                  httpFilters:
                  - name: envoy.filters.http.router
                    typedConfig:
                      '@type': type.googleapis.com/envoy.extensions.filters.http.router.v3.Router
                  internalAddressConfig:
                      cidrRanges:
                          - addressPrefix: 127.0.0.1
                            prefixLen: 32
                          - addressPrefix: ::1
                            prefixLen: 128
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
				Resource: NewOutboundListenerBuilder(envoy_common.APIV3, "127.0.0.1", 27777, core_xds.SocketAddressProtocolTCP).
					Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
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
                                            [%START_TIME%] %RESPONSE_FLAGS% default 127.0.0.1(backend)->%UPSTREAM_HOST%(other-service) took %DURATION%ms, sent %BYTES_SENT% bytes, received: %BYTES_RECEIVED% bytes
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
				Resource: NewOutboundListenerBuilder(envoy_common.APIV3, "127.0.0.1", 27777, core_xds.SocketAddressProtocolTCP).
					Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
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
				Resource: NewOutboundListenerBuilder(envoy_common.APIV3, "127.0.0.1", 27777, core_xds.SocketAddressProtocolTCP).
					Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
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
				Resource: NewOutboundListenerBuilder(envoy_common.APIV3, "127.0.0.1", 27777, core_xds.SocketAddressProtocolTCP).
					Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
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
                                            [%START_TIME%] %RESPONSE_FLAGS% default 127.0.0.1(backend)->%UPSTREAM_HOST%(other-service) took %DURATION%ms, sent %BYTES_SENT% bytes, received: %BYTES_RECEIVED% bytes
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
				Resource: NewOutboundListenerBuilder(envoy_common.APIV3, "127.0.0.1", 27777, core_xds.SocketAddressProtocolTCP).
					Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
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
				Resource: NewOutboundListenerBuilder(envoy_common.APIV3, "127.0.0.1", 27778, core_xds.SocketAddressProtocolTCP).
					Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
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
				Resource: NewOutboundListenerBuilder(envoy_common.APIV3, "127.0.0.1", 27779, core_xds.SocketAddressProtocolTCP).
					Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
						Configure(TcpProxyDeprecated(
							"127.0.0.1:27779",
							envoy_common.NewCluster(
								envoy_common.WithService("bar-service"),
								envoy_common.WithWeight(100),
							),
						)),
					)).MustBuild(),
			}},
			outbounds: []*builders.OutboundBuilder{
				builders.Outbound().
					WithService("foo-service").
					WithAddress("127.0.0.1").
					WithPort(27778),
				builders.Outbound().
					WithService("bar-service").
					WithAddress("127.0.0.1").
					WithPort(27779),
			},
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
			expectedClusters: []string{
				`
            altStatName: meshaccesslog_opentelemetry_0
            connectTimeout: 5s
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
            connectTimeout: 5s
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
            `,
			},
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
                                    stringValue: '[%START_TIME%] %RESPONSE_FLAGS% default 127.0.0.1(backend)->%UPSTREAM_HOST%(other-service) took %DURATION%ms, sent %BYTES_SENT% bytes, received: %BYTES_RECEIVED% bytes'
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
				Resource: NewOutboundListenerBuilder(envoy_common.APIV3, "127.0.0.1", 27777, core_xds.SocketAddressProtocolTCP).
					Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
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
				Resource: NewOutboundListenerBuilder(envoy_common.APIV3, "127.0.0.1", 27777, core_xds.SocketAddressProtocolTCP).
					Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
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
				Resource: NewOutboundListenerBuilder(envoy_common.APIV3, "127.0.0.1", 27777, core_xds.SocketAddressProtocolTCP).
					Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
						Configure(HttpConnectionManager("127.0.0.1:27777", false, nil)).
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
                  internalAddressConfig:
                      cidrRanges:
                          - addressPrefix: 127.0.0.1
                            prefixLen: 32
                          - addressPrefix: ::1
                            prefixLen: 128
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
				Resource: NewInboundListenerBuilder(envoy_common.APIV3, "127.0.0.1", 17777, core_xds.SocketAddressProtocolTCP).
					Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
						Configure(HttpConnectionManager("127.0.0.1:17777", false, nil)).
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
                                [%START_TIME%] default "%REQ(:METHOD)% %REQ(X-ENVOY-ORIGINAL-PATH?:PATH)% %PROTOCOL%" %RESPONSE_CODE% %RESPONSE_FLAGS% %BYTES_RECEIVED% %BYTES_SENT% %DURATION% %RESP(X-ENVOY-UPSTREAM-SERVICE-TIME)% "%REQ(X-FORWARDED-FOR)%" "%REQ(USER-AGENT)%" "%REQ(X-B3-TRACEID?X-DATADOG-TRACEID)%" "%REQ(X-REQUEST-ID)%" "%REQ(:AUTHORITY)%" "unknown" "backend" "127.0.0.1" "%UPSTREAM_HOST%"
                      path: /tmp/log
                  httpFilters:
                  - name: envoy.filters.http.router
                    typedConfig:
                      '@type': type.googleapis.com/envoy.extensions.filters.http.router.v3.Router
                  internalAddressConfig:
                      cidrRanges:
                          - addressPrefix: 127.0.0.1
                            prefixLen: 32
                          - addressPrefix: ::1
                            prefixLen: 128
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
		routes []*core_mesh.MeshGatewayRouteResource
		rules  core_rules.GatewayRules
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

			xdsCtx := *xds_builders.Context().
				WithMesh(samples.MeshDefaultBuilder()).
				WithResources(resources).
				AddServiceProtocol("backend", core_mesh.ProtocolHTTP).
				AddServiceProtocol("other-service", core_mesh.ProtocolHTTP).
				Build()
			proxy := xds_builders.Proxy().
				WithMetadata(&core_xds.DataplaneMetadata{
					AccessLogSocketPath: "/tmp/foo",
				}).
				WithDataplane(
					builders.Dataplane().
						WithName("gateway").
						WithMesh("default").
						WithBuiltInGateway("gateway"),
				).
				WithPolicies(xds_builders.MatchedPolicies().WithGatewayPolicy(api.MeshAccessLogType, given.rules)).
				Build()

			for n, p := range core_plugins.Plugins().ProxyPlugins() {
				Expect(p.Apply(context.Background(), xdsCtx.Mesh, proxy)).To(Succeed(), n)
			}

			gatewayGenerator := gateway_plugin.NewGenerator("test-zone")
			generatedResources, err := gatewayGenerator.Generate(context.Background(), nil, xdsCtx, proxy)
			Expect(err).NotTo(HaveOccurred())

			plugin := plugin.NewPlugin().(core_plugins.PolicyPlugin)
			Expect(plugin.Apply(generatedResources, xdsCtx, proxy)).To(Succeed())

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
			rules: core_rules.GatewayRules{
				FromRules: map[core_rules.InboundListener]core_rules.Rules{
					{Address: "127.0.0.1", Port: 8080}: {
						{
							Subset: core_rules.Subset{},
							Conf: api.Conf{
								Backends: &[]api.Backend{{
									File: &api.FileBackend{
										Path: "/tmp/from-log",
									},
								}},
							},
						},
					},
				},
				ToRules: core_rules.GatewayToRules{
					ByListener: map[core_rules.InboundListener]core_rules.Rules{
						{Address: "127.0.0.1", Port: 8080}: {
							{
								Subset: core_rules.Subset{},
								Conf: api.Conf{
									Backends: &[]api.Backend{{
										File: &api.FileBackend{
											Path: "/tmp/to-log",
										},
									}},
								},
							},
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
