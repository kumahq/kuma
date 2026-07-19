package tracker

import (
	"context"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_endpoint "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	envoy_service_health "github.com/envoyproxy/go-control-plane/envoy/service/health/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	envoy_cache "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	dp_server "github.com/kumahq/kuma/v3/pkg/config/dp-server"
	"github.com/kumahq/kuma/v3/pkg/core"
	core_meta "github.com/kumahq/kuma/v3/pkg/core/metadata"
	unified_naming "github.com/kumahq/kuma/v3/pkg/core/naming/unified-naming"
	"github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/v3/pkg/core/resources/manager"
	"github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/core/resources/store"
	"github.com/kumahq/kuma/v3/pkg/core/user"
	"github.com/kumahq/kuma/v3/pkg/core/xds"
	v3 "github.com/kumahq/kuma/v3/pkg/hds/v3"
	tproxy_dp "github.com/kumahq/kuma/v3/pkg/transparentproxy/config/dataplane"
	"github.com/kumahq/kuma/v3/pkg/util/net"
	util_proto "github.com/kumahq/kuma/v3/pkg/util/proto"
	util_xds_v3 "github.com/kumahq/kuma/v3/pkg/util/xds/v3"
	"github.com/kumahq/kuma/v3/pkg/xds/envoy/names"
	"github.com/kumahq/kuma/v3/pkg/xds/generator/metadata"
	"github.com/kumahq/kuma/v3/pkg/xds/generator/system_names"
)

type SnapshotGenerator struct {
	config                  *dp_server.HdsConfig
	readOnlyResourceManager manager.ReadOnlyResourceManager
	defaultAdminPort        uint32
}

func NewSnapshotGenerator(
	readOnlyResourceManager manager.ReadOnlyResourceManager,
	config *dp_server.HdsConfig,
	defaultAdminPort uint32,
) *SnapshotGenerator {
	return &SnapshotGenerator{
		readOnlyResourceManager: readOnlyResourceManager,
		config:                  config,
		defaultAdminPort:        defaultAdminPort,
	}
}

func (g *SnapshotGenerator) GenerateSnapshot(ctx context.Context, node *envoy_core.Node) (envoy_cache.ResourceSnapshot, error) {
	ctx = user.Ctx(ctx, user.ControlPlane)
	proxyId, err := xds.ParseProxyIdFromString(node.Id)
	if err != nil {
		return nil, err
	}
	dp := mesh.NewDataplaneResource()
	if err := g.readOnlyResourceManager.Get(ctx, dp, store.GetBy(proxyId.ToResourceKey())); err != nil {
		return nil, err
	}

	meshResource := mesh.NewMeshResource()
	meshFound := true
	if err := g.readOnlyResourceManager.Get(ctx, meshResource, store.GetByKey(proxyId.ToResourceKey().Mesh, model.NoMesh)); err != nil {
		if !store.IsNotFound(err) {
			return nil, err
		}
		// Mesh deletion doesn't cascade to Dataplanes (mesh_manager only deletes
		// the Mesh itself), so an orphaned Dataplane can still reach HDS. Fall
		// back to the legacy admin cluster name instead of failing the snapshot.
		meshFound = false
	}

	// TODO(unified-resource-naming): adjust when legacy naming is removed
	md := xds.DataplaneMetadataFromXdsMetadata(node.Metadata)
	// Match the admin xDS gate exactly: the unified admin cluster name is only
	// published when unified_naming.Enabled is true, which additionally requires
	// an Exclusive mesh. Keying off the feature bit alone would target the
	// non-existent unified admin cluster on non-Exclusive meshes.
	unifiedNamingEnabled := meshFound && unified_naming.Enabled(md, meshResource)
	clusterName := names.GetEnvoyAdminClusterName()
	if unifiedNamingEnabled {
		clusterName = system_names.SystemResourceNameEnvoyAdmin
	}

	healthChecks := []*envoy_service_health.ClusterHealthCheck{
		g.envoyHealthCheck(dp.AdminPort(g.defaultAdminPort), md.GetAdminSocketPath(), clusterName),
	}

	for _, inbound := range dp.Spec.GetNetworking().GetInbound() {
		if inbound.ServiceProbe == nil {
			continue
		}
		serviceProbe := inbound.ServiceProbe
		intf := dp.Spec.GetNetworking().ToInboundInterface(inbound)

		var timeout *durationpb.Duration
		if serviceProbe.Timeout == nil {
			timeout = util_proto.Duration(g.config.CheckDefaults.Timeout.Duration)
		} else {
			timeout = serviceProbe.Timeout
		}

		var interval *durationpb.Duration
		if serviceProbe.Interval == nil {
			interval = util_proto.Duration(g.config.CheckDefaults.Interval.Duration)
		} else {
			interval = serviceProbe.Interval
		}

		var healthyThreshold *wrapperspb.UInt32Value
		if serviceProbe.HealthyThreshold == nil {
			healthyThreshold = util_proto.UInt32(g.config.CheckDefaults.HealthyThreshold)
		} else {
			healthyThreshold = serviceProbe.HealthyThreshold
		}

		var unhealthyThreshold *wrapperspb.UInt32Value
		if serviceProbe.UnhealthyThreshold == nil {
			unhealthyThreshold = util_proto.UInt32(g.config.CheckDefaults.UnhealthyThreshold)
		} else {
			unhealthyThreshold = serviceProbe.UnhealthyThreshold
		}

		hc := &envoy_service_health.ClusterHealthCheck{
			ClusterName: names.GetLocalClusterName(intf.WorkloadPort),
			LocalityEndpoints: []*envoy_service_health.LocalityEndpoints{{
				Endpoints: []*envoy_endpoint.Endpoint{{
					Address: &envoy_core.Address{
						Address: &envoy_core.Address_SocketAddress{
							SocketAddress: &envoy_core.SocketAddress{
								Address: intf.WorkloadIP,
								PortSpecifier: &envoy_core.SocketAddress_PortValue{
									PortValue: intf.WorkloadPort,
								},
							},
						},
					},
				}},
			}},
			HealthChecks: []*envoy_core.HealthCheck{
				{
					Timeout:            timeout,
					Interval:           interval,
					HealthyThreshold:   healthyThreshold,
					UnhealthyThreshold: unhealthyThreshold,
					NoTrafficInterval:  util_proto.Duration(g.config.CheckDefaults.NoTrafficInterval.Duration),
					HealthChecker: &envoy_core.HealthCheck_TcpHealthCheck_{
						TcpHealthCheck: &envoy_core.HealthCheck_TcpHealthCheck{},
					},
				},
			},
		}

		meta := xds.DataplaneMetadataFromXdsMetadata(node.GetMetadata())
		tpCfg := tproxy_dp.GetDataplaneConfig(dp, meta)
		if tpCfg.Enabled() && (intf.WorkloadIP != core_meta.LoopbackIPv4.String() && intf.WorkloadIP != core_meta.LoopbackIPv6.String()) {
			if net.IsAddressIPv6(intf.WorkloadIP) {
				hc.UpstreamBindConfig = g.upstreamBindConfig(metadata.TransparentInPassThroughIPv6, 0)
			} else {
				hc.UpstreamBindConfig = g.upstreamBindConfig(metadata.TransparentInPassThroughIPv4, 0)
			}
		}

		healthChecks = append(healthChecks, hc)
	}

	hcs := &envoy_service_health.HealthCheckSpecifier{
		ClusterHealthChecks: healthChecks,
		Interval:            util_proto.Duration(g.config.Interval.Duration),
	}

	return util_xds_v3.NewSingleTypeSnapshot(core.NewUUID(), v3.HealthCheckSpecifierType, []types.Resource{hcs}), nil
}

// envoyHealthCheck builds a HC for Envoy itself so when Envoy is in draining state HDS can report that DP is offline
func (g *SnapshotGenerator) envoyHealthCheck(port uint32, adminSocketPath, clusterName string) *envoy_service_health.ClusterHealthCheck {
	var addr *envoy_core.Address
	if adminSocketPath != "" {
		addr = &envoy_core.Address{
			Address: &envoy_core.Address_Pipe{
				Pipe: &envoy_core.Pipe{
					Path: adminSocketPath,
				},
			},
		}
	} else {
		addr = &envoy_core.Address{
			Address: &envoy_core.Address_SocketAddress{
				SocketAddress: &envoy_core.SocketAddress{
					Address: "127.0.0.1",
					PortSpecifier: &envoy_core.SocketAddress_PortValue{
						PortValue: port,
					},
				},
			},
		}
	}
	return &envoy_service_health.ClusterHealthCheck{
		ClusterName: clusterName,
		LocalityEndpoints: []*envoy_service_health.LocalityEndpoints{{
			Endpoints: []*envoy_endpoint.Endpoint{{
				Address: addr,
			}},
		}},
		HealthChecks: []*envoy_core.HealthCheck{
			{
				Timeout:            util_proto.Duration(g.config.CheckDefaults.Timeout.Duration),
				Interval:           util_proto.Duration(g.config.CheckDefaults.Interval.Duration),
				HealthyThreshold:   util_proto.UInt32(g.config.CheckDefaults.HealthyThreshold),
				UnhealthyThreshold: util_proto.UInt32(g.config.CheckDefaults.UnhealthyThreshold),
				NoTrafficInterval:  util_proto.Duration(g.config.CheckDefaults.NoTrafficInterval.Duration),
				HealthChecker: &envoy_core.HealthCheck_HttpHealthCheck_{
					HttpHealthCheck: &envoy_core.HealthCheck_HttpHealthCheck{
						Path: "/ready",
					},
				},
			},
		},
	}
}

func (g *SnapshotGenerator) upstreamBindConfig(addr string, port uint32) *envoy_core.BindConfig {
	return &envoy_core.BindConfig{
		SourceAddress: &envoy_core.SocketAddress{
			Address: addr,
			PortSpecifier: &envoy_core.SocketAddress_PortValue{
				PortValue: port,
			},
		},
	}
}
