package envoy

import (
	"net"
	"time"

	"github.com/Kong/kuma/api/mesh/v1alpha1"

	"github.com/Kong/kuma/pkg/sds/server"
	"github.com/envoyproxy/go-control-plane/envoy/api/v2/auth"

	"github.com/gogo/protobuf/types"

	mesh_core "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/Kong/kuma/pkg/core/xds"
	util_error "github.com/Kong/kuma/pkg/util/error"
	xds_context "github.com/Kong/kuma/pkg/xds/context"
	v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	"github.com/envoyproxy/go-control-plane/envoy/api/v2/endpoint"
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/api/v2/listener"
	accesslog "github.com/envoyproxy/go-control-plane/envoy/config/accesslog/v2"
	filter_accesslog "github.com/envoyproxy/go-control-plane/envoy/config/filter/accesslog/v2"
	tcp "github.com/envoyproxy/go-control-plane/envoy/config/filter/network/tcp_proxy/v2"
	grpc_credential "github.com/envoyproxy/go-control-plane/envoy/config/grpc_credential/v2alpha"
	"github.com/envoyproxy/go-control-plane/pkg/util"
)

const (
	defaultConnectTimeout = 5 * time.Second
)

func CreateStaticEndpoint(clusterName string, address string, port uint32) *v2.ClusterLoadAssignment {
	return &v2.ClusterLoadAssignment{
		ClusterName: clusterName,
		Endpoints: []*endpoint.LocalityLbEndpoints{{
			LbEndpoints: []*endpoint.LbEndpoint{{
				HostIdentifier: &endpoint.LbEndpoint_Endpoint{
					Endpoint: &endpoint.Endpoint{
						Address: &core.Address{
							Address: &core.Address_SocketAddress{
								SocketAddress: &core.SocketAddress{
									Protocol: core.TCP,
									Address:  address,
									PortSpecifier: &core.SocketAddress_PortValue{
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

func CreateClusterLoadAssignment(clusterName string, endpoints []net.SRV) *v2.ClusterLoadAssignment {
	lbEndpoints := make([]*endpoint.LbEndpoint, 0, len(endpoints))
	for _, ep := range endpoints {
		lbEndpoints = append(lbEndpoints, &endpoint.LbEndpoint{
			HostIdentifier: &endpoint.LbEndpoint_Endpoint{
				Endpoint: &endpoint.Endpoint{
					Address: &core.Address{
						Address: &core.Address_SocketAddress{
							SocketAddress: &core.SocketAddress{
								Protocol: core.TCP,
								Address:  ep.Target,
								PortSpecifier: &core.SocketAddress_PortValue{
									PortValue: uint32(ep.Port),
								},
							},
						},
					},
				}},
		})
	}
	return &v2.ClusterLoadAssignment{
		ClusterName: clusterName,
		Endpoints: []*endpoint.LocalityLbEndpoints{{
			LbEndpoints: lbEndpoints,
		}},
	}
}

func CreateLocalCluster(clusterName string, address string, port uint32) *v2.Cluster {
	connectTimeout := defaultConnectTimeout
	return &v2.Cluster{
		Name:                 clusterName,
		ConnectTimeout:       &connectTimeout,
		ClusterDiscoveryType: &v2.Cluster_Type{Type: v2.Cluster_STATIC},
		LoadAssignment:       CreateStaticEndpoint(clusterName, address, port),
	}
}

func CreateEdsCluster(ctx xds_context.Context, clusterName string, metadata *core_xds.DataplaneMetadata) *v2.Cluster {
	connectTimeout := defaultConnectTimeout
	return &v2.Cluster{
		Name:                 clusterName,
		ConnectTimeout:       &connectTimeout,
		ClusterDiscoveryType: &v2.Cluster_Type{Type: v2.Cluster_EDS},
		EdsClusterConfig: &v2.Cluster_EdsClusterConfig{
			EdsConfig: &core.ConfigSource{
				ConfigSourceSpecifier: &core.ConfigSource_Ads{
					Ads: &core.AggregatedConfigSource{},
				},
			},
		},
		TlsContext: CreateUpstreamTlsContext(ctx, metadata),
	}
}

func CreatePassThroughCluster(clusterName string) *v2.Cluster {
	connectTimeout := defaultConnectTimeout
	return &v2.Cluster{
		Name:                 clusterName,
		ConnectTimeout:       &connectTimeout,
		ClusterDiscoveryType: &v2.Cluster_Type{Type: v2.Cluster_ORIGINAL_DST},
		LbPolicy:             v2.Cluster_ORIGINAL_DST_LB,
	}
}

func CreateOutboundListener(ctx xds_context.Context, listenerName string, address string, port uint32, clusterName string, virtual bool, sourceService string, destinationService string, backends []*v1alpha1.LoggingBackend, proxy *core_xds.Proxy) (*v2.Listener, error) {
	accessLog, err := convertLoggingBackends(sourceService, destinationService, backends, proxy)
	if err != nil {
		return nil, err
	}
	config := &tcp.TcpProxy{
		StatPrefix: clusterName,
		ClusterSpecifier: &tcp.TcpProxy_Cluster{
			Cluster: clusterName,
		},
		AccessLog: accessLog,
	}
	pbst, err := types.MarshalAny(config)
	util_error.MustNot(err)
	listener := &v2.Listener{
		Name: listenerName,
		Address: &core.Address{
			Address: &core.Address_SocketAddress{
				SocketAddress: &core.SocketAddress{
					Protocol: core.TCP,
					Address:  address,
					PortSpecifier: &core.SocketAddress_PortValue{
						PortValue: port,
					},
				},
			},
		},
		FilterChains: []*envoy_listener.FilterChain{{
			Filters: []*envoy_listener.Filter{{
				Name: util.TCPProxy,
				ConfigType: &envoy_listener.Filter_TypedConfig{
					TypedConfig: pbst,
				},
			}},
		}},
	}
	if virtual {
		// TODO(yskopets): What is the up-to-date alternative ?
		listener.DeprecatedV1 = &v2.Listener_DeprecatedV1{
			BindToPort: &types.BoolValue{Value: false},
		}
	}
	return listener, nil
}

func CreateInboundListener(ctx xds_context.Context, listenerName string, address string, port uint32, clusterName string, virtual bool, permissions *mesh_core.TrafficPermissionResourceList, metadata *core_xds.DataplaneMetadata) *v2.Listener {
	config := &tcp.TcpProxy{
		StatPrefix: clusterName,
		ClusterSpecifier: &tcp.TcpProxy_Cluster{
			Cluster: clusterName,
		},
		AccessLog: accessLog(ctx),
	}
	pbst, err := types.MarshalAny(config)
	util_error.MustNot(err)
	listener := &v2.Listener{
		Name: listenerName,
		Address: &core.Address{
			Address: &core.Address_SocketAddress{
				SocketAddress: &core.SocketAddress{
					Protocol: core.TCP,
					Address:  address,
					PortSpecifier: &core.SocketAddress_PortValue{
						PortValue: port,
					},
				},
			},
		},
		FilterChains: []*envoy_listener.FilterChain{{
			TlsContext: CreateDownstreamTlsContext(ctx, metadata),
			Filters: []*envoy_listener.Filter{{
				Name: util.TCPProxy,
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
			BindToPort: &types.BoolValue{Value: false},
		}
	}
	return listener
}

func accessLog(ctx xds_context.Context) []*filter_accesslog.AccessLog {
	if !ctx.Mesh.LoggingEnabled {
		return []*filter_accesslog.AccessLog{}
	}
	fileAccessLog := &accesslog.FileAccessLog{
		AccessLogFormat: &accesslog.FileAccessLog_Format{
			Format: "[%START_TIME%] %DOWNSTREAM_REMOTE_ADDRESS%->%UPSTREAM_HOST% took %DURATION%ms, sent %BYTES_SENT% bytes, received: %BYTES_RECEIVED% bytes\n",
		},
		Path: ctx.Mesh.LoggingPath,
	}
	pbst, err := types.MarshalAny(fileAccessLog)
	util_error.MustNot(err)
	logs := &filter_accesslog.AccessLog{
		Name: util.FileAccessLog,
		ConfigType: &filter_accesslog.AccessLog_TypedConfig{
			TypedConfig: pbst,
		},
	}
	return []*filter_accesslog.AccessLog{logs}
}

func CreateDownstreamTlsContext(ctx xds_context.Context, metadata *core_xds.DataplaneMetadata) *auth.DownstreamTlsContext {
	if !ctx.Mesh.TlsEnabled {
		return nil
	}
	return &auth.DownstreamTlsContext{
		CommonTlsContext:         CreateCommonTlsContext(ctx, metadata),
		RequireClientCertificate: &types.BoolValue{Value: true},
	}
}

func CreateUpstreamTlsContext(ctx xds_context.Context, metadata *core_xds.DataplaneMetadata) *auth.UpstreamTlsContext {
	if !ctx.Mesh.TlsEnabled {
		return nil
	}
	return &auth.UpstreamTlsContext{
		CommonTlsContext: CreateCommonTlsContext(ctx, metadata),
	}
}

func CreateCommonTlsContext(ctx xds_context.Context, metadata *core_xds.DataplaneMetadata) *auth.CommonTlsContext {
	return &auth.CommonTlsContext{
		ValidationContextType: &auth.CommonTlsContext_ValidationContextSdsSecretConfig{
			ValidationContextSdsSecretConfig: sdsSecretConfig(ctx, server.MeshCaResource, metadata),
		},
		TlsCertificateSdsSecretConfigs: []*auth.SdsSecretConfig{
			sdsSecretConfig(ctx, server.IdentityCertResource, metadata),
		},
	}
}

func sdsSecretConfig(context xds_context.Context, name string, metadata *core_xds.DataplaneMetadata) *auth.SdsSecretConfig {
	withCallCredentials := func(grpc *core.GrpcService_GoogleGrpc) *core.GrpcService_GoogleGrpc {
		if metadata.DataplaneTokenPath == "" {
			return grpc
		}

		config := &grpc_credential.FileBasedMetadataConfig{
			SecretData: &core.DataSource{
				Specifier: &core.DataSource_Filename{
					Filename: metadata.DataplaneTokenPath,
				},
			},
		}
		typedConfig, err := types.MarshalAny(config)
		util_error.MustNot(err)

		grpc.CallCredentials = append(grpc.CallCredentials, &core.GrpcService_GoogleGrpc_CallCredentials{
			CredentialSpecifier: &core.GrpcService_GoogleGrpc_CallCredentials_FromPlugin{
				FromPlugin: &core.GrpcService_GoogleGrpc_CallCredentials_MetadataCredentialsFromPlugin{
					Name: "envoy.grpc_credentials.file_based_metadata",
					ConfigType: &core.GrpcService_GoogleGrpc_CallCredentials_MetadataCredentialsFromPlugin_TypedConfig{
						TypedConfig: typedConfig,
					},
				},
			},
		})
		grpc.CredentialsFactoryName = "envoy.grpc_credentials.file_based_metadata"

		return grpc
	}
	return &auth.SdsSecretConfig{
		Name: name,
		SdsConfig: &core.ConfigSource{
			ConfigSourceSpecifier: &core.ConfigSource_ApiConfigSource{
				ApiConfigSource: &core.ApiConfigSource{
					ApiType: core.ApiConfigSource_GRPC,
					GrpcServices: []*core.GrpcService{
						{
							TargetSpecifier: &core.GrpcService_GoogleGrpc_{
								GoogleGrpc: withCallCredentials(&core.GrpcService_GoogleGrpc{
									TargetUri:  context.ControlPlane.SdsLocation,
									StatPrefix: "sds_" + name,
									ChannelCredentials: &core.GrpcService_GoogleGrpc_ChannelCredentials{
										CredentialSpecifier: &core.GrpcService_GoogleGrpc_ChannelCredentials_SslCredentials{
											SslCredentials: &core.GrpcService_GoogleGrpc_SslCredentials{
												RootCerts: &core.DataSource{
													Specifier: &core.DataSource_InlineBytes{
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
	config := &tcp.TcpProxy{
		StatPrefix: clusterName,
		ClusterSpecifier: &tcp.TcpProxy_Cluster{
			Cluster: clusterName,
		},
		AccessLog: accessLog(ctx),
	}
	pbst, err := types.MarshalAny(config)
	util_error.MustNot(err)
	return &v2.Listener{
		Name: listenerName,
		Address: &core.Address{
			Address: &core.Address_SocketAddress{
				SocketAddress: &core.SocketAddress{
					Protocol: core.TCP,
					Address:  address,
					PortSpecifier: &core.SocketAddress_PortValue{
						PortValue: port,
					},
				},
			},
		},
		FilterChains: []*envoy_listener.FilterChain{{
			Filters: []*envoy_listener.Filter{{
				Name: util.TCPProxy,
				ConfigType: &envoy_listener.Filter_TypedConfig{
					TypedConfig: pbst,
				},
			}},
		}},
		// TODO(yskopets): What is the up-to-date alternative ?
		UseOriginalDst: &types.BoolValue{Value: true},
		// TODO(yskopets): Apparently, `envoy.envoy_listener.original_dst` has different effect than `UseOriginalDst`
		//ListenerFilters: []envoy_listener.ListenerFilter{{
		//	Name: util.OriginalDestination,
		//}},
	}
}
