package bootstrap

import (
	"net"
	"strconv"
	"time"

	"github.com/asaskevich/govalidator"
	envoy_accesslog_v3 "github.com/envoyproxy/go-control-plane/envoy/config/accesslog/v3"
	envoy_bootstrap_v3 "github.com/envoyproxy/go-control-plane/envoy/config/bootstrap/v3"
	envoy_cluster_v3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_config_endpoint_v3 "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	envoy_grpc_credentials_v3 "github.com/envoyproxy/go-control-plane/envoy/config/grpc_credential/v3"
	envoy_metrics_v3 "github.com/envoyproxy/go-control-plane/envoy/config/metrics/v3"
	envoy_overload_v3 "github.com/envoyproxy/go-control-plane/envoy/config/overload/v3"
	access_loggers_file "github.com/envoyproxy/go-control-plane/envoy/extensions/access_loggers/file/v3"
	regex_engines "github.com/envoyproxy/go-control-plane/envoy/extensions/regex_engines/v3"
	resource_monitors_fixed_heap "github.com/envoyproxy/go-control-plane/envoy/extensions/resource_monitors/fixed_heap/v3"
	envoy_tls "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	envoy_type_matcher_v3 "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/kumahq/kuma/pkg/config/xds"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	clusters_v3 "github.com/kumahq/kuma/pkg/xds/envoy/clusters/v3"
	"github.com/kumahq/kuma/pkg/xds/envoy/names"
	"github.com/kumahq/kuma/pkg/xds/envoy/tls"
)

var BootstrapClusters = map[string]struct{}{}

func RegisterBootstrapCluster(c string) string {
	BootstrapClusters[c] = struct{}{}
	return c
}

var (
	adsClusterName           = RegisterBootstrapCluster(names.GetAdsClusterName())
	accessLogSinkClusterName = RegisterBootstrapCluster(names.GetAccessLogSinkClusterName())
)

func genConfig(parameters configParameters, proxyConfig xds.Proxy, enableReloadableTokens bool) (*envoy_bootstrap_v3.Bootstrap, error) {
	staticClusters, err := buildStaticClusters(parameters, enableReloadableTokens)
	if err != nil {
		return nil, err
	}

	features := []interface{}{}
	for _, feature := range parameters.Features {
		features = append(features, feature)
	}

	runtimeLayers := []*envoy_bootstrap_v3.RuntimeLayer{{
		Name: "kuma",
		LayerSpecifier: &envoy_bootstrap_v3.RuntimeLayer_StaticLayer{
			StaticLayer: util_proto.MustStruct(map[string]interface{}{
				"re2.max_program_size.error_level": 4294967295,
				"re2.max_program_size.warn_level":  1000,
			}),
		},
	}}

	if parameters.IsGatewayDataplane {
		connections := proxyConfig.Gateway.GlobalDownstreamMaxConnections
		if connections == 0 {
			connections = 50000
		}

		runtimeLayers = append(runtimeLayers,
			&envoy_bootstrap_v3.RuntimeLayer{
				Name: "gateway",
				LayerSpecifier: &envoy_bootstrap_v3.RuntimeLayer_StaticLayer{
					StaticLayer: util_proto.MustStruct(map[string]interface{}{
						"overload.global_downstream_max_connections": connections,
					}),
				},
			},
			&envoy_bootstrap_v3.RuntimeLayer{
				Name: "gateway.listeners",
				LayerSpecifier: &envoy_bootstrap_v3.RuntimeLayer_RtdsLayer_{
					RtdsLayer: &envoy_bootstrap_v3.RuntimeLayer_RtdsLayer{
						Name: "gateway.listeners",
						RtdsConfig: &envoy_core_v3.ConfigSource{
							ResourceApiVersion:    envoy_core_v3.ApiVersion_V3,
							ConfigSourceSpecifier: &envoy_core_v3.ConfigSource_Ads{},
						},
					},
				},
			})
	}

	// We create matchers
	var matchNames []*envoy_tls.SubjectAltNameMatcher
	for _, typ := range []envoy_tls.SubjectAltNameMatcher_SanType{
		envoy_tls.SubjectAltNameMatcher_DNS,
		envoy_tls.SubjectAltNameMatcher_IP_ADDRESS,
	} {
		matchNames = append(matchNames, &envoy_tls.SubjectAltNameMatcher{
			SanType: typ,
			Matcher: &envoy_type_matcher_v3.StringMatcher{
				MatchPattern: &envoy_type_matcher_v3.StringMatcher_Exact{Exact: parameters.XdsHost},
			},
		})
	}
	res := &envoy_bootstrap_v3.Bootstrap{
		Node: &envoy_core_v3.Node{
			Id:      parameters.Id,
			Cluster: parameters.Service,
			Metadata: &structpb.Struct{
				Fields: map[string]*structpb.Value{
					core_xds.FieldVersion: {
						Kind: &structpb.Value_StructValue{
							StructValue: util_proto.MustToStruct(parameters.Version),
						},
					},
					core_xds.FieldFeatures:        util_proto.MustNewValueForStruct(features),
					core_xds.FieldWorkdir:         util_proto.MustNewValueForStruct(parameters.Workdir),
					core_xds.FieldMetricsCertPath: util_proto.MustNewValueForStruct(parameters.MetricsCertPath),
					core_xds.FieldMetricsKeyPath:  util_proto.MustNewValueForStruct(parameters.MetricsKeyPath),
					core_xds.FieldSystemCaPath:    util_proto.MustNewValueForStruct(parameters.SystemCaPath),
				},
			},
		},
		LayeredRuntime: &envoy_bootstrap_v3.LayeredRuntime{
			Layers: runtimeLayers,
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
					TagValue: &envoy_metrics_v3.TagSpecifier_Regex{Regex: "^kafka\\..*\\.(.*?(?=_duration|$))"},
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
				InitialFetchTimeout:   durationpb.New(0),
				ResourceApiVersion:    envoy_core_v3.ApiVersion_V3,
			},
			CdsConfig: &envoy_core_v3.ConfigSource{
				ConfigSourceSpecifier: &envoy_core_v3.ConfigSource_Ads{Ads: &envoy_core_v3.AggregatedConfigSource{}},
				InitialFetchTimeout:   durationpb.New(0),
				ResourceApiVersion:    envoy_core_v3.ApiVersion_V3,
			},
			AdsConfig: &envoy_core_v3.ApiConfigSource{
				ApiType:                   envoy_core_v3.ApiConfigSource_GRPC,
				TransportApiVersion:       envoy_core_v3.ApiVersion_V3,
				SetNodeOnFirstMessageOnly: true,
				GrpcServices: []*envoy_core_v3.GrpcService{
					buildGrpcService(parameters, enableReloadableTokens),
				},
			},
		},
		StaticResources: &envoy_bootstrap_v3.Bootstrap_StaticResources{
			Secrets: []*envoy_tls.Secret{
				{
					Name: tls.CpValidationCtx,
					Type: &envoy_tls.Secret_ValidationContext{
						ValidationContext: &envoy_tls.CertificateValidationContext{
							MatchTypedSubjectAltNames: matchNames,
							TrustedCa: &envoy_core_v3.DataSource{
								Specifier: &envoy_core_v3.DataSource_InlineBytes{
									InlineBytes: parameters.CertBytes,
								},
							},
						},
					},
				},
			},
			Clusters: staticClusters,
		},
		DefaultRegexEngine: &envoy_core_v3.TypedExtensionConfig{
			Name:        "envoy.regex_engines.google_re2",
			TypedConfig: util_proto.MustMarshalAny(&regex_engines.GoogleRE2{}),
		},
	}
	for _, r := range res.StaticResources.Clusters {
		if r.Name == adsClusterName {
			transport := &envoy_tls.UpstreamTlsContext{
				Sni: parameters.XdsHost,
				CommonTlsContext: &envoy_tls.CommonTlsContext{
					TlsParams: &envoy_tls.TlsParameters{
						TlsMinimumProtocolVersion: envoy_tls.TlsParameters_TLSv1_2,
					},
					ValidationContextType: &envoy_tls.CommonTlsContext_ValidationContextSdsSecretConfig{
						ValidationContextSdsSecretConfig: &envoy_tls.SdsSecretConfig{
							Name: tls.CpValidationCtx,
						},
					},
				},
			}
			any, err := util_proto.MarshalAnyDeterministic(transport)
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
				buildGrpcService(parameters, enableReloadableTokens),
			},
		}
	}

	if parameters.IsGatewayDataplane {
		if maxBytes := parameters.Resources.MaxHeapSizeBytes; maxBytes > 0 {
			config := &resource_monitors_fixed_heap.FixedHeapConfig{
				MaxHeapSizeBytes: maxBytes,
			}
			marshaledConfig, err := util_proto.MarshalAnyDeterministic(config)
			if err != nil {
				return nil, errors.Wrapf(err, "could not marshall %T", config)
			}

			fixedHeap := "envoy.resource_monitors.fixed_heap"

			res.OverloadManager = &envoy_overload_v3.OverloadManager{
				RefreshInterval: util_proto.Duration(250 * time.Millisecond),
				ResourceMonitors: []*envoy_overload_v3.ResourceMonitor{{
					Name: fixedHeap,
					ConfigType: &envoy_overload_v3.ResourceMonitor_TypedConfig{
						TypedConfig: marshaledConfig,
					},
				}},
				Actions: []*envoy_overload_v3.OverloadAction{{
					Name: "envoy.overload_actions.shrink_heap",
					Triggers: []*envoy_overload_v3.Trigger{{
						Name: fixedHeap,
						TriggerOneof: &envoy_overload_v3.Trigger_Threshold{
							Threshold: &envoy_overload_v3.ThresholdTrigger{
								Value: 0.95,
							},
						},
					}},
				}, {
					Name: "envoy.overload_actions.stop_accepting_requests",
					Triggers: []*envoy_overload_v3.Trigger{{
						Name: fixedHeap,
						TriggerOneof: &envoy_overload_v3.Trigger_Threshold{
							Threshold: &envoy_overload_v3.ThresholdTrigger{
								Value: 0.98,
							},
						},
					}},
				}},
			}
		}
	}

	if (!enableReloadableTokens || parameters.DataplaneTokenPath == "") && parameters.DataplaneToken != "" {
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
		res.Node.Metadata.Fields[core_xds.FieldDataplaneDataplaneResource] = util_proto.MustNewValueForStruct(parameters.DataplaneResource)
	}
	if parameters.AdminPort != 0 {
		res.Node.Metadata.Fields[core_xds.FieldDataplaneAdminPort] = util_proto.MustNewValueForStruct(strconv.Itoa(int(parameters.AdminPort)))
		res.Node.Metadata.Fields[core_xds.FieldDataplaneAdminAddress] = util_proto.MustNewValueForStruct(parameters.AdminAddress)
		res.Admin = &envoy_bootstrap_v3.Admin{
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
		if parameters.AdminAccessLogPath != "" {
			fileAccessLog := &access_loggers_file.FileAccessLog{
				Path: parameters.AdminAccessLogPath,
			}
			marshaled, err := util_proto.MarshalAnyDeterministic(fileAccessLog)
			if err != nil {
				return nil, errors.Wrapf(err, "could not marshall %T", fileAccessLog)
			}
			res.Admin.AccessLog = []*envoy_accesslog_v3.AccessLog{
				{
					Name: "envoy.access_loggers.file",
					ConfigType: &envoy_accesslog_v3.AccessLog_TypedConfig{
						TypedConfig: marshaled,
					},
				},
			}
		}
	}
	if parameters.DNSPort != 0 {
		res.Node.Metadata.Fields[core_xds.FieldDataplaneDNSPort] = util_proto.MustNewValueForStruct(strconv.Itoa(int(parameters.DNSPort)))
	}
	if parameters.ReadinessPort != 0 {
		res.Node.Metadata.Fields[core_xds.FieldDataplaneReadinessPort] = util_proto.MustNewValueForStruct(strconv.Itoa(int(parameters.ReadinessPort)))
	}
	if parameters.ProxyType != "" {
		res.Node.Metadata.Fields[core_xds.FieldDataplaneProxyType] = util_proto.MustNewValueForStruct(parameters.ProxyType)
	}
	if len(parameters.DynamicMetadata) > 0 {
		md := make(map[string]interface{}, len(parameters.DynamicMetadata))
		for k, v := range parameters.DynamicMetadata {
			md[k] = v
		}
		res.Node.Metadata.Fields[core_xds.FieldDynamicMetadata] = util_proto.MustNewValueForStruct(md)
	}
	return res, nil
}

func dnsLookupFamilyFromXdsHost(host string, lookupFn func(host string) ([]net.IP, error)) envoy_cluster_v3.Cluster_DnsLookupFamily {
	if govalidator.IsDNSName(host) && host != "localhost" {
		ips, err := lookupFn(host)
		if err != nil {
			log.Info("[WARNING] error looking up XDS host to determine DnsLookupFamily, falling back to AUTO", "hostname", host)
			return envoy_cluster_v3.Cluster_AUTO
		}
		hasIPv6 := false

		for _, ip := range ips {
			if ip.To4() == nil {
				hasIPv6 = true
			}
		}

		if !hasIPv6 && len(ips) > 0 {
			return envoy_cluster_v3.Cluster_V4_ONLY
		}
	}

	return envoy_cluster_v3.Cluster_AUTO // default
}

func clusterTypeFromHost(host string) envoy_cluster_v3.Cluster_DiscoveryType {
	if govalidator.IsIP(host) {
		return envoy_cluster_v3.Cluster_STATIC
	}
	return envoy_cluster_v3.Cluster_STRICT_DNS
}

func buildGrpcService(params configParameters, useTokenPath bool) *envoy_core_v3.GrpcService {
	if useTokenPath && params.DataplaneTokenPath != "" {
		googleGrpcService := &envoy_core_v3.GrpcService{
			TargetSpecifier: &envoy_core_v3.GrpcService_GoogleGrpc_{
				GoogleGrpc: &envoy_core_v3.GrpcService_GoogleGrpc{
					TargetUri:              net.JoinHostPort(params.XdsHost, strconv.FormatUint(uint64(params.XdsPort), 10)),
					StatPrefix:             "ads",
					CredentialsFactoryName: "envoy.grpc_credentials.file_based_metadata",
					CallCredentials: []*envoy_core_v3.GrpcService_GoogleGrpc_CallCredentials{
						{
							CredentialSpecifier: &envoy_core_v3.GrpcService_GoogleGrpc_CallCredentials_FromPlugin{
								FromPlugin: &envoy_core_v3.GrpcService_GoogleGrpc_CallCredentials_MetadataCredentialsFromPlugin{
									Name: "envoy.grpc_credentials.file_based_metadata",
									ConfigType: &envoy_core_v3.GrpcService_GoogleGrpc_CallCredentials_MetadataCredentialsFromPlugin_TypedConfig{
										TypedConfig: util_proto.MustMarshalAny(&envoy_grpc_credentials_v3.FileBasedMetadataConfig{
											SecretData: &envoy_core_v3.DataSource{
												Specifier: &envoy_core_v3.DataSource_Filename{Filename: params.DataplaneTokenPath},
											},
										}),
									},
								},
							},
						},
					},
				},
			},
		}
		if params.CertBytes != nil {
			googleGrpcService.GetGoogleGrpc().ChannelCredentials = &envoy_core_v3.GrpcService_GoogleGrpc_ChannelCredentials{
				CredentialSpecifier: &envoy_core_v3.GrpcService_GoogleGrpc_ChannelCredentials_SslCredentials{
					SslCredentials: &envoy_core_v3.GrpcService_GoogleGrpc_SslCredentials{
						RootCerts: &envoy_core_v3.DataSource{
							Specifier: &envoy_core_v3.DataSource_InlineBytes{
								InlineBytes: params.CertBytes,
							},
						},
					},
				},
			}
		}
		return googleGrpcService
	} else {
		envoyGrpcSerivce := &envoy_core_v3.GrpcService{
			TargetSpecifier: &envoy_core_v3.GrpcService_EnvoyGrpc_{
				EnvoyGrpc: &envoy_core_v3.GrpcService_EnvoyGrpc{
					ClusterName: adsClusterName,
				},
			},
		}
		return envoyGrpcSerivce
	}
}

func buildStaticClusters(parameters configParameters, enableReloadableTokens bool) ([]*envoy_cluster_v3.Cluster, error) {
	proxyId, err := core_xds.ParseProxyIdFromString(parameters.Id)
	if err != nil {
		return nil, err
	}

	accessLogSink := &envoy_cluster_v3.Cluster{
		// TODO does timeout and keepAlive make sense on this as it uses unix domain sockets?
		Name:           accessLogSinkClusterName,
		ConnectTimeout: util_proto.Duration(parameters.XdsConnectTimeout),
		LbPolicy:       envoy_cluster_v3.Cluster_ROUND_ROBIN,
		UpstreamConnectionOptions: &envoy_cluster_v3.UpstreamConnectionOptions{
			TcpKeepalive: &envoy_core_v3.TcpKeepalive{
				KeepaliveProbes:   util_proto.UInt32(3),
				KeepaliveTime:     util_proto.UInt32(10),
				KeepaliveInterval: util_proto.UInt32(10),
			},
		},
		ClusterDiscoveryType: &envoy_cluster_v3.Cluster_Type{Type: envoy_cluster_v3.Cluster_STATIC},
		LoadAssignment: &envoy_config_endpoint_v3.ClusterLoadAssignment{
			ClusterName: accessLogSinkClusterName,
			Endpoints: []*envoy_config_endpoint_v3.LocalityLbEndpoints{
				{
					LbEndpoints: []*envoy_config_endpoint_v3.LbEndpoint{
						{
							HostIdentifier: &envoy_config_endpoint_v3.LbEndpoint_Endpoint{
								Endpoint: &envoy_config_endpoint_v3.Endpoint{
									Address: &envoy_core_v3.Address{
										Address: &envoy_core_v3.Address_Pipe{Pipe: &envoy_core_v3.Pipe{Path: core_xds.AccessLogSocketName(parameters.Workdir, proxyId.ToResourceKey().Name, proxyId.ToResourceKey().Mesh)}},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	if err := (&clusters_v3.Http2Configurer{}).Configure(accessLogSink); err != nil {
		return nil, err
	}

	clusters := []*envoy_cluster_v3.Cluster{accessLogSink}

	if parameters.DataplaneTokenPath == "" || !enableReloadableTokens {
		adsCluster := &envoy_cluster_v3.Cluster{
			Name:           adsClusterName,
			ConnectTimeout: util_proto.Duration(parameters.XdsConnectTimeout),
			LbPolicy:       envoy_cluster_v3.Cluster_ROUND_ROBIN,
			UpstreamConnectionOptions: &envoy_cluster_v3.UpstreamConnectionOptions{
				TcpKeepalive: &envoy_core_v3.TcpKeepalive{
					KeepaliveProbes:   util_proto.UInt32(3),
					KeepaliveTime:     util_proto.UInt32(10),
					KeepaliveInterval: util_proto.UInt32(10),
				},
			},
			ClusterDiscoveryType: &envoy_cluster_v3.Cluster_Type{Type: clusterTypeFromHost(parameters.XdsHost)},
			DnsLookupFamily:      dnsLookupFamilyFromXdsHost(parameters.XdsHost, net.LookupIP),
			LoadAssignment: &envoy_config_endpoint_v3.ClusterLoadAssignment{
				ClusterName: adsClusterName,
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
		}
		if err := (&clusters_v3.Http2Configurer{}).Configure(adsCluster); err != nil {
			return nil, err
		}

		clusters = append(clusters, adsCluster)
	}
	return clusters, nil
}
