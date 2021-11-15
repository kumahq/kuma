package bootstrap

import (
	"strconv"

	"github.com/asaskevich/govalidator"
	envoy_bootstrap_v3 "github.com/envoyproxy/go-control-plane/envoy/config/bootstrap/v3"
	envoy_cluster_v3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_config_endpoint_v3 "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	envoy_metrics_v3 "github.com/envoyproxy/go-control-plane/envoy/config/metrics/v3"
	envoy_tls "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	envoy_type_matcher_v3 "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func mustNewStruct(in map[string]interface{}) *structpb.Struct {
	r, err := structpb.NewStruct(in)
	if err != nil {
		panic(err)
	}
	return r
}

func genConfig(parameters configParameters) (*envoy_bootstrap_v3.Bootstrap, error) {
	res := &envoy_bootstrap_v3.Bootstrap{
		Node: &envoy_core_v3.Node{
			Id:      parameters.Id,
			Cluster: parameters.Service,
			Metadata: mustNewStruct(map[string]interface{}{
				"version": map[string]interface{}{
					"kumaDp": map[string]interface{}{
						"version":   parameters.KumaDpVersion,
						"gitTag":    parameters.KumaDpGitTag,
						"gitCommit": parameters.KumaDpGitCommit,
						"buildDate": parameters.KumaDpBuildDate,
					},
					"envoy": map[string]interface{}{
						"version": parameters.EnvoyVersion,
						"build":   parameters.EnvoyBuild,
					},
				},
			}),
		},
		LayeredRuntime: &envoy_bootstrap_v3.LayeredRuntime{
			Layers: []*envoy_bootstrap_v3.RuntimeLayer{
				{
					Name: "kuma",
					LayerSpecifier: &envoy_bootstrap_v3.RuntimeLayer_StaticLayer{
						StaticLayer: mustNewStruct(map[string]interface{}{
							"envoy.restart_features.use_apple_api_for_dns_lookups": false,
							"re2.max_program_size.error_level":                     4294967295,
							"re2.max_program_size.warn_level":                      1000,
						}),
					},
				},
			},
		},
		StatsConfig: &envoy_metrics_v3.StatsConfig{
			StatsTags: []*envoy_metrics_v3.TagSpecifier{
				{
					TagName:  "name",
					TagValue: &envoy_metrics_v3.TagSpecifier_Regex{Regex: "^grpc\\.((.+)\\.)"},
				},
				{
					TagName:  "status",
					TagValue: &envoy_metrics_v3.TagSpecifier_Regex{Regex: "^grpc.*streams_closed(_([0-9]+))"},
				},
				{
					TagName:  "kafka_name",
					TagValue: &envoy_metrics_v3.TagSpecifier_Regex{Regex: "^kafka(\\.(\\S*[0-9]))\\."},
				},
				{
					TagName:  "kafka_type",
					TagValue: &envoy_metrics_v3.TagSpecifier_Regex{Regex: "^kafka\\..*\\.(.*)"},
				},
				{
					TagName:  "worker",
					TagValue: &envoy_metrics_v3.TagSpecifier_Regex{Regex: "(worker_([0-9]+)\\.)"},
				},
				{
					TagName:  "listener",
					TagValue: &envoy_metrics_v3.TagSpecifier_Regex{Regex: "((.+?)\\.)rbac\\."},
				},
			},
		},
		DynamicResources: &envoy_bootstrap_v3.Bootstrap_DynamicResources{
			LdsConfig: &envoy_core_v3.ConfigSource{
				ConfigSourceSpecifier: &envoy_core_v3.ConfigSource_Ads{Ads: &envoy_core_v3.AggregatedConfigSource{}},
				ResourceApiVersion:    envoy_core_v3.ApiVersion_V3,
			},
			CdsConfig: &envoy_core_v3.ConfigSource{
				ConfigSourceSpecifier: &envoy_core_v3.ConfigSource_Ads{Ads: &envoy_core_v3.AggregatedConfigSource{}},
				ResourceApiVersion:    envoy_core_v3.ApiVersion_V3,
			},
			AdsConfig: &envoy_core_v3.ApiConfigSource{
				ApiType:                   envoy_core_v3.ApiConfigSource_GRPC,
				TransportApiVersion:       envoy_core_v3.ApiVersion_V3,
				SetNodeOnFirstMessageOnly: true,
				GrpcServices: []*envoy_core_v3.GrpcService{
					{
						TargetSpecifier: &envoy_core_v3.GrpcService_EnvoyGrpc_{
							EnvoyGrpc: &envoy_core_v3.GrpcService_EnvoyGrpc{
								ClusterName: "ads_cluster",
							},
						},
					},
				},
			},
		},
		StaticResources: &envoy_bootstrap_v3.Bootstrap_StaticResources{
			Clusters: []*envoy_cluster_v3.Cluster{
				{
					// TODO does timeout and keepAlive make sense on this as it uses unix domain sockets?
					Name:                 "access_log_sink",
					ConnectTimeout:       durationpb.New(parameters.XdsConnectTimeout),
					Http2ProtocolOptions: &envoy_core_v3.Http2ProtocolOptions{},
					LbPolicy:             envoy_cluster_v3.Cluster_ROUND_ROBIN,
					UpstreamConnectionOptions: &envoy_cluster_v3.UpstreamConnectionOptions{
						TcpKeepalive: &envoy_core_v3.TcpKeepalive{
							KeepaliveProbes:   wrapperspb.UInt32(3),
							KeepaliveTime:     wrapperspb.UInt32(10),
							KeepaliveInterval: wrapperspb.UInt32(10),
						},
					},
					ClusterDiscoveryType: &envoy_cluster_v3.Cluster_Type{Type: envoy_cluster_v3.Cluster_STATIC},
					LoadAssignment: &envoy_config_endpoint_v3.ClusterLoadAssignment{
						ClusterName: "access_log_sink",
						Endpoints: []*envoy_config_endpoint_v3.LocalityLbEndpoints{
							{
								LbEndpoints: []*envoy_config_endpoint_v3.LbEndpoint{
									{
										HostIdentifier: &envoy_config_endpoint_v3.LbEndpoint_Endpoint{
											Endpoint: &envoy_config_endpoint_v3.Endpoint{
												Address: &envoy_core_v3.Address{
													Address: &envoy_core_v3.Address_Pipe{Pipe: &envoy_core_v3.Pipe{Path: parameters.AccessLogPipe}},
												},
											},
										},
									},
								},
							},
						},
					},
				},
				{
					Name:                 "ads_cluster",
					ConnectTimeout:       durationpb.New(parameters.XdsConnectTimeout),
					Http2ProtocolOptions: &envoy_core_v3.Http2ProtocolOptions{},
					LbPolicy:             envoy_cluster_v3.Cluster_ROUND_ROBIN,
					UpstreamConnectionOptions: &envoy_cluster_v3.UpstreamConnectionOptions{
						TcpKeepalive: &envoy_core_v3.TcpKeepalive{
							KeepaliveProbes:   wrapperspb.UInt32(3),
							KeepaliveTime:     wrapperspb.UInt32(10),
							KeepaliveInterval: wrapperspb.UInt32(10),
						},
					},
					ClusterDiscoveryType: &envoy_cluster_v3.Cluster_Type{Type: clusterTypeFromHost(parameters.XdsHost)},
					LoadAssignment: &envoy_config_endpoint_v3.ClusterLoadAssignment{
						ClusterName: "ads_cluster",
						Endpoints: []*envoy_config_endpoint_v3.LocalityLbEndpoints{
							{
								LbEndpoints: []*envoy_config_endpoint_v3.LbEndpoint{
									{
										HostIdentifier: &envoy_config_endpoint_v3.LbEndpoint_Endpoint{
											Endpoint: &envoy_config_endpoint_v3.Endpoint{
												Address: &envoy_core_v3.Address{
													Address: &envoy_core_v3.Address_SocketAddress{
														SocketAddress: &envoy_core_v3.SocketAddress{
															Address:       parameters.XdsHost,
															PortSpecifier: &envoy_core_v3.SocketAddress_PortValue{PortValue: parameters.XdsPort},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	for _, r := range res.StaticResources.Clusters {
		if r.Name == "ads_cluster" {
			transport := &envoy_tls.UpstreamTlsContext{
				Sni: parameters.XdsHost,
				CommonTlsContext: &envoy_tls.CommonTlsContext{
					TlsParams: &envoy_tls.TlsParameters{
						TlsMinimumProtocolVersion: envoy_tls.TlsParameters_TLSv1_2,
					},
					ValidationContextType: &envoy_tls.CommonTlsContext_ValidationContext{
						ValidationContext: &envoy_tls.CertificateValidationContext{
							MatchSubjectAltNames: []*envoy_type_matcher_v3.StringMatcher{
								{
									MatchPattern: &envoy_type_matcher_v3.StringMatcher_Exact{Exact: parameters.XdsHost},
								},
							},
						},
					},
				},
			}
			if parameters.CertBytes != nil {
				transport.CommonTlsContext.GetValidationContext().TrustedCa = &envoy_core_v3.DataSource{
					Specifier: &envoy_core_v3.DataSource_InlineBytes{
						InlineBytes: parameters.CertBytes,
					},
				}
			}
			any, err := anypb.New(transport)
			if err != nil {
				return nil, err
			}
			r.TransportSocket = &envoy_core_v3.TransportSocket{
				Name: "envoy.transport_sockets.tls",
				ConfigType: &envoy_core_v3.TransportSocket_TypedConfig{
					TypedConfig: any,
				},
			}
		}
	}
	if parameters.HdsEnabled {
		res.HdsConfig = &envoy_core_v3.ApiConfigSource{
			ApiType:                   envoy_core_v3.ApiConfigSource_GRPC,
			TransportApiVersion:       envoy_core_v3.ApiVersion_V3,
			SetNodeOnFirstMessageOnly: true,
			GrpcServices: []*envoy_core_v3.GrpcService{
				{
					TargetSpecifier: &envoy_core_v3.GrpcService_EnvoyGrpc_{
						EnvoyGrpc: &envoy_core_v3.GrpcService_EnvoyGrpc{
							ClusterName: "ads_cluster",
						},
					},
				},
			},
		}
	}

	if parameters.DataplaneToken != "" {
		res.Node.Metadata.Fields["dataplane.token"] = structpb.NewStringValue(parameters.DataplaneToken)
		if res.HdsConfig != nil {
			for _, n := range res.HdsConfig.GrpcServices {
				n.InitialMetadata = []*envoy_core_v3.HeaderValue{
					{Key: "authorization", Value: parameters.DataplaneToken},
				}
			}
		}
		for _, n := range res.DynamicResources.AdsConfig.GrpcServices {
			n.InitialMetadata = []*envoy_core_v3.HeaderValue{
				{Key: "authorization", Value: parameters.DataplaneToken},
			}
		}
	}
	if parameters.DataplaneResource != "" {
		res.Node.Metadata.Fields["dataplane.resource"] = structpb.NewStringValue(parameters.DataplaneResource)
	}
	if parameters.AdminPort != 0 {
		res.Node.Metadata.Fields["dataplane.admin.port"] = structpb.NewStringValue(strconv.Itoa(int(parameters.AdminPort)))
		res.Admin = &envoy_bootstrap_v3.Admin{
			AccessLogPath: parameters.AdminAccessLogPath,
			Address: &envoy_core_v3.Address{
				Address: &envoy_core_v3.Address_SocketAddress{
					SocketAddress: &envoy_core_v3.SocketAddress{
						Address:  parameters.AdminAddress,
						Protocol: envoy_core_v3.SocketAddress_TCP,
						PortSpecifier: &envoy_core_v3.SocketAddress_PortValue{
							PortValue: parameters.AdminPort,
						},
					},
				},
			},
		}
	}
	if parameters.DNSPort != 0 {
		res.Node.Metadata.Fields["dataplane.dns.port"] = structpb.NewStringValue(strconv.Itoa(int(parameters.DNSPort)))
	}
	if parameters.EmptyDNSPort != 0 {
		res.Node.Metadata.Fields["dataplane.dns.empty.port"] = structpb.NewStringValue(strconv.Itoa(int(parameters.EmptyDNSPort)))
	}
	if parameters.ProxyType != "" {
		res.Node.Metadata.Fields["dataplane.proxyType"] = structpb.NewStringValue(parameters.ProxyType)
	}
	if len(parameters.DynamicMetadata) > 0 {
		dm, _ := structpb.NewStruct(map[string]interface{}{})
		for k, v := range parameters.DynamicMetadata {
			dm.Fields[k] = structpb.NewStringValue(v)
		}
		res.Node.Metadata.Fields["dynamicMetadata"] = structpb.NewStructValue(dm)
	}
	return res, nil
}

func clusterTypeFromHost(host string) envoy_cluster_v3.Cluster_DiscoveryType {
	if govalidator.IsIP(host) {
		return envoy_cluster_v3.Cluster_STATIC
	}
	return envoy_cluster_v3.Cluster_STRICT_DNS
}
