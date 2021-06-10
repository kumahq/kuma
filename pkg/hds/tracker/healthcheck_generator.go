package tracker

import (
	"context"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_endpoint "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	envoy_service_health "github.com/envoyproxy/go-control-plane/envoy/service/health/v3"
	"github.com/golang/protobuf/ptypes/wrappers"
	"google.golang.org/protobuf/types/known/durationpb"

	dp_server "github.com/kumahq/kuma/pkg/config/dp-server"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/hds/cache"
	util_xds_v3 "github.com/kumahq/kuma/pkg/util/xds/v3"
	"github.com/kumahq/kuma/pkg/xds/envoy/names"
)

type SnapshotGenerator struct {
	config                  *dp_server.HdsConfig
	readOnlyResourceManager manager.ReadOnlyResourceManager
}

func NewSnapshotGenerator(readOnlyResourceManager manager.ReadOnlyResourceManager, config *dp_server.HdsConfig) *SnapshotGenerator {
	return &SnapshotGenerator{
		readOnlyResourceManager: readOnlyResourceManager,
		config:                  config,
	}
}

func (g *SnapshotGenerator) GenerateSnapshot(node *envoy_core.Node) (util_xds_v3.Snapshot, error) {
	proxyId, err := xds.ParseProxyIdFromString(node.Id)
	if err != nil {
		return nil, err
	}
	dp := mesh.NewDataplaneResource()
	if err := g.readOnlyResourceManager.Get(context.Background(), dp, store.GetBy(proxyId.ToResourceKey())); err != nil {
		return nil, err
	}

	var healthChecks []*envoy_service_health.ClusterHealthCheck
	for _, inbound := range dp.Spec.GetNetworking().GetInbound() {
		if inbound.ServiceProbe == nil {
			continue
		}
		serviceProbe := inbound.ServiceProbe
		intf := dp.Spec.GetNetworking().ToInboundInterface(inbound)

		var timeout *durationpb.Duration
		if serviceProbe.Timeout == nil {
			timeout = durationpb.New(g.config.CheckDefaults.Timeout)
		} else {
			timeout = serviceProbe.Timeout
		}

		var interval *durationpb.Duration
		if serviceProbe.Timeout == nil {
			interval = durationpb.New(g.config.CheckDefaults.Interval)
		} else {
			interval = serviceProbe.Interval
		}

		var healthyThreshold *wrappers.UInt32Value
		if serviceProbe.HealthyThreshold == nil {
			healthyThreshold = &wrappers.UInt32Value{Value: g.config.CheckDefaults.HealthyThreshold}
		} else {
			healthyThreshold = serviceProbe.HealthyThreshold
		}

		var unhealthyThreshold *wrappers.UInt32Value
		if serviceProbe.UnhealthyThreshold == nil {
			unhealthyThreshold = &wrappers.UInt32Value{Value: g.config.CheckDefaults.UnhealthyThreshold}
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
					NoTrafficInterval:  durationpb.New(g.config.CheckDefaults.NoTrafficInterval),
					HealthChecker: &envoy_core.HealthCheck_TcpHealthCheck_{
						TcpHealthCheck: &envoy_core.HealthCheck_TcpHealthCheck{},
					},
				},
			},
		})
	}

	hcs := &envoy_service_health.HealthCheckSpecifier{
		ClusterHealthChecks: healthChecks,
		Interval:            durationpb.New(g.config.Interval),
	}

	return cache.NewSnapshot("", hcs), nil
}
