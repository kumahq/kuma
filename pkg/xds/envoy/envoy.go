package envoy

import (
	"time"

	"github.com/Kong/kuma/pkg/sds/server"
	envoy_auth "github.com/envoyproxy/go-control-plane/envoy/api/v2/auth"

	"github.com/golang/protobuf/ptypes"
	pstruct "github.com/golang/protobuf/ptypes/struct"
	"github.com/golang/protobuf/ptypes/wrappers"

	mesh_core "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/Kong/kuma/pkg/core/xds"
	util_error "github.com/Kong/kuma/pkg/util/error"
	util_xds "github.com/Kong/kuma/pkg/util/xds"
	xds_context "github.com/Kong/kuma/pkg/xds/context"
	v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/api/v2/cluster"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	envoy_endpoint "github.com/envoyproxy/go-control-plane/envoy/api/v2/endpoint"
	envoy_grpc_credential "github.com/envoyproxy/go-control-plane/envoy/config/grpc_credential/v2alpha"
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
	return clusterWithAltStatName(&v2.Cluster{
		Name:                 clusterName,
		ConnectTimeout:       ptypes.DurationProto(defaultConnectTimeout),
		ClusterDiscoveryType: &v2.Cluster_Type{Type: v2.Cluster_STATIC},
		LoadAssignment:       CreateStaticEndpoint(clusterName, address, port),
	})
}

func CreateEdsCluster(ctx xds_context.Context, clusterName string, metadata *core_xds.DataplaneMetadata) *v2.Cluster {
	return clusterWithAltStatName(&v2.Cluster{
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
	})
}

func clusterWithAltStatName(cluster *v2.Cluster) *v2.Cluster {
	sanitizedName := util_xds.SanitizeMetric(cluster.Name)
	if sanitizedName != cluster.Name {
		cluster.AltStatName = sanitizedName
	}
	return cluster
}

func ClusterWithHealthChecks(cluster *v2.Cluster, healthCheck *mesh_core.HealthCheckResource) *v2.Cluster {
	if healthCheck == nil {
		return cluster
	}
	if healthCheck.HasActiveChecks() {
		activeChecks := healthCheck.Spec.Conf.GetActiveChecks()
		cluster.HealthChecks = append(cluster.HealthChecks, &envoy_core.HealthCheck{
			HealthChecker: &envoy_core.HealthCheck_TcpHealthCheck_{
				TcpHealthCheck: &envoy_core.HealthCheck_TcpHealthCheck{},
			},
			Interval:           activeChecks.Interval,
			Timeout:            activeChecks.Timeout,
			UnhealthyThreshold: &wrappers.UInt32Value{Value: activeChecks.UnhealthyThreshold},
			HealthyThreshold:   &wrappers.UInt32Value{Value: activeChecks.HealthyThreshold},
		})
	}
	if healthCheck.HasPassiveChecks() {
		passiveChecks := healthCheck.Spec.Conf.GetPassiveChecks()
		cluster.OutlierDetection = &envoy_cluster.OutlierDetection{
			Interval:        passiveChecks.PenaltyInterval,
			Consecutive_5Xx: &wrappers.UInt32Value{Value: passiveChecks.UnhealthyThreshold},
		}
	}
	return cluster
}

func CreatePassThroughCluster(clusterName string) *v2.Cluster {
	return clusterWithAltStatName(&v2.Cluster{
		Name:                 clusterName,
		ConnectTimeout:       ptypes.DurationProto(defaultConnectTimeout),
		ClusterDiscoveryType: &v2.Cluster_Type{Type: v2.Cluster_ORIGINAL_DST},
		LbPolicy:             v2.Cluster_ORIGINAL_DST_LB,
	})
}

func CreateDownstreamTlsContext(ctx xds_context.Context, metadata *core_xds.DataplaneMetadata) *envoy_auth.DownstreamTlsContext {
	if !ctx.Mesh.Resource.Spec.GetMtls().GetEnabled() {
		return nil
	}
	return &envoy_auth.DownstreamTlsContext{
		CommonTlsContext:         CreateCommonTlsContext(ctx, metadata),
		RequireClientCertificate: &wrappers.BoolValue{Value: true},
	}
}

func CreateUpstreamTlsContext(ctx xds_context.Context, metadata *core_xds.DataplaneMetadata) *envoy_auth.UpstreamTlsContext {
	if !ctx.Mesh.Resource.Spec.GetMtls().GetEnabled() {
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
		if metadata.GetDataplaneTokenPath() == "" {
			return grpc
		}

		config := &envoy_grpc_credential.FileBasedMetadataConfig{
			SecretData: &envoy_core.DataSource{
				Specifier: &envoy_core.DataSource_Filename{
					Filename: metadata.GetDataplaneTokenPath(),
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
									StatPrefix: util_xds.SanitizeMetric("sds_" + name),
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
