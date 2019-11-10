package envoy

import (
	"time"

	"github.com/Kong/kuma/api/mesh/v1alpha1"

	"github.com/Kong/kuma/pkg/sds/server"
	envoy_auth "github.com/envoyproxy/go-control-plane/envoy/api/v2/auth"

	"github.com/golang/protobuf/ptypes"
	pstruct "github.com/golang/protobuf/ptypes/struct"
	"github.com/golang/protobuf/ptypes/wrappers"

	mesh_core "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/Kong/kuma/pkg/core/xds"
	util_error "github.com/Kong/kuma/pkg/util/error"
	xds_context "github.com/Kong/kuma/pkg/xds/context"
	v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	envoy_endpoint "github.com/envoyproxy/go-control-plane/envoy/api/v2/endpoint"
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/api/v2/listener"
	envoy_accesslog "github.com/envoyproxy/go-control-plane/envoy/config/accesslog/v2"
	envoy_filter_accesslog "github.com/envoyproxy/go-control-plane/envoy/config/filter/accesslog/v2"
	envoy_tcp "github.com/envoyproxy/go-control-plane/envoy/config/filter/network/tcp_proxy/v2"
	envoy_grpc_credential "github.com/envoyproxy/go-control-plane/envoy/config/grpc_credential/v2alpha"
	wellknown "github.com/envoyproxy/go-control-plane/pkg/wellknown"
)

const (
	defaultConnectTimeout = 5 * time.Second
)

type ClusterInfo struct {
	Name   string
	Weight uint32
	Tags   map[string]string
}

func CreateStaticEndpoint(clusterName string, address string, port uint32) *v2.ClusterLoadAssignment {
	return &v2.ClusterLoadAssignment{
		ClusterName: clusterName,
		Endpoints: []*envoy_endpoint.LocalityLbEndpoints{{
			LbEndpoints: []*envoy_endpoint.LbEndpoint{{
				HostIdentifier: &envoy_endpoint.LbEndpoint_Endpoint{
					Endpoint: &envoy_endpoint.Endpoint{
						Address: &envoy_core.Address{
							Address: &envoy_core.Address_SocketAddress{
								SocketAddress: &envoy_core.SocketAddress{
									Protocol: envoy_core.SocketAddress_TCP,
									Address:  address,
									PortSpecifier: &envoy_core.SocketAddress_PortValue{
										PortValue: port,
									},
								},
							},
						},
					},
				},
			}},
		}},
	}
}

func CreateLbMetadata(tags map[string]string) *envoy_core.Metadata {
	if len(tags) == 0 {
		return nil
	}
	fields := map[string]*pstruct.Value{}
	for key, value := range tags {
		fields[key] = &pstruct.Value{
			Kind: &pstruct.Value_StringValue{
				StringValue: value,
			},
		}
	}
	return &envoy_core.Metadata{
		FilterMetadata: map[string]*pstruct.Struct{
			"envoy.lb": &pstruct.Struct{
				Fields: fields,
			},
		},
	}
}

func CreateClusterLoadAssignment(clusterName string, endpoints []core_xds.Endpoint) *v2.ClusterLoadAssignment {
	lbEndpoints := make([]*envoy_endpoint.LbEndpoint, 0, len(endpoints))
	for _, ep := range endpoints {
		lbEndpoints = append(lbEndpoints, &envoy_endpoint.LbEndpoint{
			Metadata: CreateLbMetadata(ep.Tags),
			HostIdentifier: &envoy_endpoint.LbEndpoint_Endpoint{
				Endpoint: &envoy_endpoint.Endpoint{
					Address: &envoy_core.Address{
						Address: &envoy_core.Address_SocketAddress{
							SocketAddress: &envoy_core.SocketAddress{
								Protocol: envoy_core.SocketAddress_TCP,
								Address:  ep.Target,
								PortSpecifier: &envoy_core.SocketAddress_PortValue{
									PortValue: ep.Port,
								},
							},
						},
					},
				}},
		})
	}
	return &v2.ClusterLoadAssignment{
		ClusterName: clusterName,
		Endpoints: []*envoy_endpoint.LocalityLbEndpoints{{
			LbEndpoints: lbEndpoints,
		}},
	}
}

func CreateLocalCluster(clusterName string, address string, port uint32) *v2.Cluster {
	return &v2.Cluster{
		Name:                 clusterName,
		ConnectTimeout:       ptypes.DurationProto(defaultConnectTimeout),
		ClusterDiscoveryType: &v2.Cluster_Type{Type: v2.Cluster_STATIC},
		LoadAssignment:       CreateStaticEndpoint(clusterName, address, port),
	}
}

func CreateEdsCluster(ctx xds_context.Context, clusterName string, metadata *core_xds.DataplaneMetadata) *v2.Cluster {
	return &v2.Cluster{
		Name:                 clusterName,
		ConnectTimeout:       ptypes.DurationProto(defaultConnectTimeout),
		ClusterDiscoveryType: &v2.Cluster_Type{Type: v2.Cluster_EDS},
		EdsClusterConfig: &v2.Cluster_EdsClusterConfig{
			EdsConfig: &envoy_core.ConfigSource{
				ConfigSourceSpecifier: &envoy_core.ConfigSource_Ads{
					Ads: &envoy_core.AggregatedConfigSource{},
				},
			},
		},
		TlsContext: CreateUpstreamTlsContext(ctx, metadata),
	}
}

func CreatePassThroughCluster(clusterName string) *v2.Cluster {
	return &v2.Cluster{
		Name:                 clusterName,
		ConnectTimeout:       ptypes.DurationProto(defaultConnectTimeout),
		ClusterDiscoveryType: &v2.Cluster_Type{Type: v2.Cluster_ORIGINAL_DST},
		LbPolicy:             v2.Cluster_ORIGINAL_DST_LB,
	}
}

func CreateOutboundListener(ctx xds_context.Context, listenerName string, address string, port uint32, statsName string, clusters []ClusterInfo, virtual bool, sourceService string, destinationService string, backends []*v1alpha1.LoggingBackend, proxy *core_xds.Proxy) (*v2.Listener, error) {
	accessLog, err := convertLoggingBackends(sourceService, destinationService, backends, proxy)
	if err != nil {
		return nil, err
	}
	config := &envoy_tcp.TcpProxy{
		StatPrefix: statsName,
		AccessLog:  accessLog,
	}
	if len(clusters) == 1 {
		config.ClusterSpecifier = &envoy_tcp.TcpProxy_Cluster{
			Cluster: clusters[0].Name,
		}
	} else {
		var weightedClusters []*envoy_tcp.TcpProxy_WeightedCluster_ClusterWeight
		for _, cluster := range clusters {
			weightedClusters = append(weightedClusters, &envoy_tcp.TcpProxy_WeightedCluster_ClusterWeight{
				Name:   cluster.Name,
				Weight: cluster.Weight,
			})
		}
		config.ClusterSpecifier = &envoy_tcp.TcpProxy_WeightedClusters{
			WeightedClusters: &envoy_tcp.TcpProxy_WeightedCluster{
				Clusters: weightedClusters,
			},
		}
	}
	pbst, err := ptypes.MarshalAny(config)
	util_error.MustNot(err)
	listener := &v2.Listener{
		Name: listenerName,
		Address: &envoy_core.Address{
			Address: &envoy_core.Address_SocketAddress{
				SocketAddress: &envoy_core.SocketAddress{
					Protocol: envoy_core.SocketAddress_TCP,
					Address:  address,
					PortSpecifier: &envoy_core.SocketAddress_PortValue{
						PortValue: port,
					},
				},
			},
		},
		FilterChains: []*envoy_listener.FilterChain{{
			Filters: []*envoy_listener.Filter{{
				Name: wellknown.TCPProxy,
				ConfigType: &envoy_listener.Filter_TypedConfig{
					TypedConfig: pbst,
				},
			}},
		}},
	}
	if virtual {
		// TODO(yskopets): What is the up-to-date alternative ?
		listener.DeprecatedV1 = &v2.Listener_DeprecatedV1{
			BindToPort: &wrappers.BoolValue{Value: false},
		}
	}
	return listener, nil
}

func CreateInboundListener(ctx xds_context.Context, listenerName string, address string, port uint32, clusterName string, virtual bool, permissions *mesh_core.TrafficPermissionResourceList, metadata *core_xds.DataplaneMetadata) *v2.Listener {
	config := &envoy_tcp.TcpProxy{
		StatPrefix: clusterName,
		ClusterSpecifier: &envoy_tcp.TcpProxy_Cluster{
			Cluster: clusterName,
		},
		AccessLog: accessLog(ctx),
	}
	pbst, err := ptypes.MarshalAny(config)
	util_error.MustNot(err)
	listener := &v2.Listener{
		Name: listenerName,
		Address: &envoy_core.Address{
			Address: &envoy_core.Address_SocketAddress{
				SocketAddress: &envoy_core.SocketAddress{
					Protocol: envoy_core.SocketAddress_TCP,
					Address:  address,
					PortSpecifier: &envoy_core.SocketAddress_PortValue{
						PortValue: port,
					},
				},
			},
		},
		FilterChains: []*envoy_listener.FilterChain{{
			TlsContext: CreateDownstreamTlsContext(ctx, metadata),
			Filters: []*envoy_listener.Filter{{
				Name: wellknown.TCPProxy,
				ConfigType: &envoy_listener.Filter_TypedConfig{
					TypedConfig: pbst,
				},
			}},
		}},
	}

	if ctx.Mesh.TlsEnabled {
		filter := createRbacFilter(listenerName, permissions)
		// RBAC filter should be first in chain
		listener.FilterChains[0].Filters = append([]*envoy_listener.Filter{&filter}, listener.FilterChains[0].Filters...)
	}

	if virtual {
		// TODO(yskopets): What is the up-to-date alternative ?
		listener.DeprecatedV1 = &v2.Listener_DeprecatedV1{
			BindToPort: &wrappers.BoolValue{Value: false},
		}
	}
	return listener
}

func accessLog(ctx xds_context.Context) []*envoy_filter_accesslog.AccessLog {
	if !ctx.Mesh.LoggingEnabled {
		return []*envoy_filter_accesslog.AccessLog{}
	}
	fileAccessLog := &envoy_accesslog.FileAccessLog{
		AccessLogFormat: &envoy_accesslog.FileAccessLog_Format{
			Format: "[%START_TIME%] %DOWNSTREAM_REMOTE_ADDRESS%->%UPSTREAM_HOST% took %DURATION%ms, sent %BYTES_SENT% bytes, received: %BYTES_RECEIVED% bytes\n",
		},
		Path: ctx.Mesh.LoggingPath,
	}
	pbst, err := ptypes.MarshalAny(fileAccessLog)
	util_error.MustNot(err)
	logs := &envoy_filter_accesslog.AccessLog{
		Name: wellknown.FileAccessLog,
		ConfigType: &envoy_filter_accesslog.AccessLog_TypedConfig{
			TypedConfig: pbst,
		},
	}
	return []*envoy_filter_accesslog.AccessLog{logs}
}

func CreateDownstreamTlsContext(ctx xds_context.Context, metadata *core_xds.DataplaneMetadata) *envoy_auth.DownstreamTlsContext {
	if !ctx.Mesh.TlsEnabled {
		return nil
	}
	return &envoy_auth.DownstreamTlsContext{
		CommonTlsContext:         CreateCommonTlsContext(ctx, metadata),
		RequireClientCertificate: &wrappers.BoolValue{Value: true},
	}
}

func CreateUpstreamTlsContext(ctx xds_context.Context, metadata *core_xds.DataplaneMetadata) *envoy_auth.UpstreamTlsContext {
	if !ctx.Mesh.TlsEnabled {
		return nil
	}
	return &envoy_auth.UpstreamTlsContext{
		CommonTlsContext: CreateCommonTlsContext(ctx, metadata),
	}
}

func CreateCommonTlsContext(ctx xds_context.Context, metadata *core_xds.DataplaneMetadata) *envoy_auth.CommonTlsContext {
	return &envoy_auth.CommonTlsContext{
		ValidationContextType: &envoy_auth.CommonTlsContext_ValidationContextSdsSecretConfig{
			ValidationContextSdsSecretConfig: sdsSecretConfig(ctx, server.MeshCaResource, metadata),
		},
		TlsCertificateSdsSecretConfigs: []*envoy_auth.SdsSecretConfig{
			sdsSecretConfig(ctx, server.IdentityCertResource, metadata),
		},
	}
}

func sdsSecretConfig(context xds_context.Context, name string, metadata *core_xds.DataplaneMetadata) *envoy_auth.SdsSecretConfig {
	withCallCredentials := func(grpc *envoy_core.GrpcService_GoogleGrpc) *envoy_core.GrpcService_GoogleGrpc {
		if metadata.DataplaneTokenPath == "" {
			return grpc
		}

		config := &envoy_grpc_credential.FileBasedMetadataConfig{
			SecretData: &envoy_core.DataSource{
				Specifier: &envoy_core.DataSource_Filename{
					Filename: metadata.DataplaneTokenPath,
				},
			},
		}
		typedConfig, err := ptypes.MarshalAny(config)
		util_error.MustNot(err)

		grpc.CallCredentials = append(grpc.CallCredentials, &envoy_core.GrpcService_GoogleGrpc_CallCredentials{
			CredentialSpecifier: &envoy_core.GrpcService_GoogleGrpc_CallCredentials_FromPlugin{
				FromPlugin: &envoy_core.GrpcService_GoogleGrpc_CallCredentials_MetadataCredentialsFromPlugin{
					Name: "envoy.grpc_credentials.file_based_metadata",
					ConfigType: &envoy_core.GrpcService_GoogleGrpc_CallCredentials_MetadataCredentialsFromPlugin_TypedConfig{
						TypedConfig: typedConfig,
					},
				},
			},
		})
		grpc.CredentialsFactoryName = "envoy.grpc_credentials.file_based_metadata"

		return grpc
	}
	return &envoy_auth.SdsSecretConfig{
		Name: name,
		SdsConfig: &envoy_core.ConfigSource{
			ConfigSourceSpecifier: &envoy_core.ConfigSource_ApiConfigSource{
				ApiConfigSource: &envoy_core.ApiConfigSource{
					ApiType: envoy_core.ApiConfigSource_GRPC,
					GrpcServices: []*envoy_core.GrpcService{
						{
							TargetSpecifier: &envoy_core.GrpcService_GoogleGrpc_{
								GoogleGrpc: withCallCredentials(&envoy_core.GrpcService_GoogleGrpc{
									TargetUri:  context.ControlPlane.SdsLocation,
									StatPrefix: "sds_" + name,
									ChannelCredentials: &envoy_core.GrpcService_GoogleGrpc_ChannelCredentials{
										CredentialSpecifier: &envoy_core.GrpcService_GoogleGrpc_ChannelCredentials_SslCredentials{
											SslCredentials: &envoy_core.GrpcService_GoogleGrpc_SslCredentials{
												RootCerts: &envoy_core.DataSource{
													Specifier: &envoy_core.DataSource_InlineBytes{
														InlineBytes: context.ControlPlane.SdsTlsCert,
													},
												},
											},
										},
									},
								}),
							},
						},
					},
				},
			},
		},
	}
}

func CreateCatchAllListener(ctx xds_context.Context, listenerName string, address string, port uint32, clusterName string) *v2.Listener {
	config := &envoy_tcp.TcpProxy{
		StatPrefix: clusterName,
		ClusterSpecifier: &envoy_tcp.TcpProxy_Cluster{
			Cluster: clusterName,
		},
		AccessLog: accessLog(ctx),
	}
	pbst, err := ptypes.MarshalAny(config)
	util_error.MustNot(err)
	return &v2.Listener{
		Name: listenerName,
		Address: &envoy_core.Address{
			Address: &envoy_core.Address_SocketAddress{
				SocketAddress: &envoy_core.SocketAddress{
					Protocol: envoy_core.SocketAddress_TCP,
					Address:  address,
					PortSpecifier: &envoy_core.SocketAddress_PortValue{
						PortValue: port,
					},
				},
			},
		},
		FilterChains: []*envoy_listener.FilterChain{{
			Filters: []*envoy_listener.Filter{{
				Name: wellknown.TCPProxy,
				ConfigType: &envoy_listener.Filter_TypedConfig{
					TypedConfig: pbst,
				},
			}},
		}},
		// TODO(yskopets): What is the up-to-date alternative ?
		UseOriginalDst: &wrappers.BoolValue{Value: true},
		// TODO(yskopets): Apparently, `envoy.envoy_listener.original_dst` has different effect than `UseOriginalDst`
		//ListenerFilters: []envoy_listener.ListenerFilter{{
		//	Name: wellknown.OriginalDestination,
		//}},
	}
}
