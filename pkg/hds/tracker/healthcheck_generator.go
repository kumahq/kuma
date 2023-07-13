package tracker

import (
	"context"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_endpoint "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	envoy_service_health "github.com/envoyproxy/go-control-plane/envoy/service/health/v3"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	dp_server "github.com/kumahq/kuma/pkg/config/dp-server"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/core/user"
	"github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/hds/cache"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	util_xds_v3 "github.com/kumahq/kuma/pkg/util/xds/v3"
	"github.com/kumahq/kuma/pkg/xds/envoy/names"
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

func (g *SnapshotGenerator) GenerateSnapshot(ctx context.Context, node *envoy_core.Node) (util_xds_v3.Snapshot, error) {
	ctx = user.Ctx(ctx, user.ControlPlane)
	proxyId, err := xds.ParseProxyIdFromString(node.Id)
	if err != nil {
		return nil, err
	}
	dp := mesh.NewDataplaneResource()
	if err := g.readOnlyResourceManager.Get(ctx, dp, store.GetBy(proxyId.ToResourceKey())); err != nil {
		return nil, err
	}

	healthChecks := []*envoy_service_health.ClusterHealthCheck{
		g.envoyHealthCheck(dp.AdminPort(g.defaultAdminPort)),
	}

	for _, inbound := range dp.Spec.GetNetworking().GetInbound() {
		if inbound.ServiceProbe == nil {
			continue
		}
		serviceProbe := inbound.ServiceProbe
		intf := dp.Spec.GetNetworking().ToInboundInterface(inbound)

		var timeout *durationpb.Duration
		if serviceProbe.Timeout == nil {
			timeout = util_proto.Duration(g.config.CheckDefaults.Timeout)
		} else {
			timeout = serviceProbe.Timeout
		}

		var interval *durationpb.Duration
		if serviceProbe.Timeout == nil {
			interval = util_proto.Duration(g.config.CheckDefaults.Interval)
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

		healthChecks = append(healthChecks, &envoy_service_health.ClusterHealthCheck{
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
					NoTrafficInterval:  util_proto.Duration(g.config.CheckDefaults.NoTrafficInterval),
					HealthChecker: &envoy_core.HealthCheck_TcpHealthCheck_{
						TcpHealthCheck: &envoy_core.HealthCheck_TcpHealthCheck{},
					},
				},
			},
		})
	}

	hcs := &envoy_service_health.HealthCheckSpecifier{
		ClusterHealthChecks: healthChecks,
		Interval:            util_proto.Duration(g.config.Interval),
	}

	return cache.NewSnapshot("", hcs), nil
}

// envoyHealthCheck builds a HC for Envoy itself so when Envoy is in draining state HDS can report that DP is offline
func (g *SnapshotGenerator) envoyHealthCheck(port uint32) *envoy_service_health.ClusterHealthCheck {
	return &envoy_service_health.ClusterHealthCheck{
		ClusterName: names.GetEnvoyAdminClusterName(),
		LocalityEndpoints: []*envoy_service_health.LocalityEndpoints{{
			Endpoints: []*envoy_endpoint.Endpoint{{
				Address: &envoy_core.Address{
					Address: &envoy_core.Address_SocketAddress{
						SocketAddress: &envoy_core.SocketAddress{
							Address: "127.0.0.1",
							PortSpecifier: &envoy_core.SocketAddress_PortValue{
								PortValue: port,
							},
						},
					},
				},
			}},
		}},
		HealthChecks: []*envoy_core.HealthCheck{
			{
				Timeout:            util_proto.Duration(g.config.CheckDefaults.Timeout),
				Interval:           util_proto.Duration(g.config.CheckDefaults.Interval),
				HealthyThreshold:   util_proto.UInt32(g.config.CheckDefaults.HealthyThreshold),
				UnhealthyThreshold: util_proto.UInt32(g.config.CheckDefaults.UnhealthyThreshold),
				NoTrafficInterval:  util_proto.Duration(g.config.CheckDefaults.NoTrafficInterval),
				HealthChecker: &envoy_core.HealthCheck_HttpHealthCheck_{
					HttpHealthCheck: &envoy_core.HealthCheck_HttpHealthCheck{
						Path: "/ready",
					},
				},
			},
		},
	}
}
