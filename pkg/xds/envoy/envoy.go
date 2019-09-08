package envoy

import (
	"github.com/Kong/kuma/pkg/sds/server"
	"github.com/envoyproxy/go-control-plane/envoy/api/v2/auth"
	"time"

	"github.com/gogo/protobuf/types"

	util_error "github.com/Kong/kuma/pkg/util/error"
	xds_context "github.com/Kong/kuma/pkg/xds/context"
	v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	"github.com/envoyproxy/go-control-plane/envoy/api/v2/endpoint"
	"github.com/envoyproxy/go-control-plane/envoy/api/v2/listener"
	tcp "github.com/envoyproxy/go-control-plane/envoy/config/filter/network/tcp_proxy/v2"
	"github.com/envoyproxy/go-control-plane/pkg/util"
)

func CreateStaticEndpoint(clusterName string, address string, port uint32) *v2.ClusterLoadAssignment {
	return &v2.ClusterLoadAssignment{
		ClusterName: clusterName,
		Endpoints: []endpoint.LocalityLbEndpoints{{
			LbEndpoints: []endpoint.LbEndpoint{{
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

func CreateLocalCluster(clusterName string, address string, port uint32) *v2.Cluster {
	return &v2.Cluster{
		Name:                 clusterName,
		ConnectTimeout:       5 * time.Second,
		ClusterDiscoveryType: &v2.Cluster_Type{Type: v2.Cluster_STATIC},
		LoadAssignment:       CreateStaticEndpoint(clusterName, address, port),
	}
}

func CreatePassThroughCluster(clusterName string) *v2.Cluster {
	return &v2.Cluster{
		Name:                 clusterName,
		ConnectTimeout:       5 * time.Second,
		ClusterDiscoveryType: &v2.Cluster_Type{Type: v2.Cluster_ORIGINAL_DST},
		LbPolicy:             v2.Cluster_ORIGINAL_DST_LB,
	}
}

func CreateInboundListener(ctx xds_context.Context, listenerName string, address string, port uint32, clusterName string, virtual bool) *v2.Listener {
	config := &tcp.TcpProxy{
		StatPrefix: clusterName,
		ClusterSpecifier: &tcp.TcpProxy_Cluster{
			Cluster: clusterName,
		},
	}
	pbst, err := types.MarshalAny(config)
	util_error.MustNot(err)
	listener := &v2.Listener{
		Name: listenerName,
		Address: core.Address{
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
		FilterChains: []listener.FilterChain{{
			TlsContext: downstreamTlsContext(ctx),
			Filters: []listener.Filter{{
				Name: util.TCPProxy,
				ConfigType: &listener.Filter_TypedConfig{
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
	return listener
}

func downstreamTlsContext(ctx xds_context.Context) *auth.DownstreamTlsContext {
	if !ctx.Mesh.TlsEnabled {
		return nil
	}
	return &auth.DownstreamTlsContext{
		CommonTlsContext: &auth.CommonTlsContext{
			ValidationContextType: &auth.CommonTlsContext_ValidationContextSdsSecretConfig{
				ValidationContextSdsSecretConfig: sdsSecretConfig(ctx, server.MeshCaResource),
			},
			TlsCertificateSdsSecretConfigs: []*auth.SdsSecretConfig{
				sdsSecretConfig(ctx, server.IdentityCertResource),
			},
		},
	}
}

func sdsSecretConfig(context xds_context.Context, name string) *auth.SdsSecretConfig {
	return &auth.SdsSecretConfig{
		Name: name,
		SdsConfig: &core.ConfigSource{
			ConfigSourceSpecifier: &core.ConfigSource_ApiConfigSource{
				ApiConfigSource: &core.ApiConfigSource{
					ApiType: core.ApiConfigSource_GRPC,
					GrpcServices: []*core.GrpcService{
						{
							TargetSpecifier: &core.GrpcService_GoogleGrpc_{
								GoogleGrpc: &core.GrpcService_GoogleGrpc{
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
								},
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
	}
	pbst, err := types.MarshalAny(config)
	util_error.MustNot(err)
	return &v2.Listener{
		Name: listenerName,
		Address: core.Address{
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
		FilterChains: []listener.FilterChain{{
			Filters: []listener.Filter{{
				Name: util.TCPProxy,
				ConfigType: &listener.Filter_TypedConfig{
					TypedConfig: pbst,
				},
			}},
		}},
		// TODO(yskopets): What is the up-to-date alternative ?
		UseOriginalDst: &types.BoolValue{Value: true},
		// TODO(yskopets): Apparently, `envoy.listener.original_dst` has different effect than `UseOriginalDst`
		//ListenerFilters: []listener.ListenerFilter{{
		//	Name: util.OriginalDestination,
		//}},
	}
}
