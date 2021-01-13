package hds

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
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/hds/cache"
	"github.com/kumahq/kuma/pkg/xds/envoy/names"
)

type generator struct {
	config                  *dp_server.HdsConfig
	readOnlyResourceManager manager.ReadOnlyResourceManager
}

func (g *generator) GenerateSnapshot(node *envoy_core.Node) (cache.Snapshot, error) {
	proxyID, err := xds.ParseProxyIdFromString(node.Id)
	if err != nil {
		return nil, err
	}
	dpKey := model.ResourceKey{Mesh: proxyID.Mesh, Name: proxyID.Name}
	dp := mesh.NewDataplaneResource()
	if err := g.readOnlyResourceManager.Get(context.Background(), dp, store.GetBy(dpKey)); err != nil {
		return nil, err
	}

	inbounds, err := dp.Spec.GetNetworking().GetInboundInterfaces()
	if err != nil {
		return nil, err
	}

	var healthChecks []*envoy_service_health.ClusterHealthCheck
	for _, inbound := range inbounds {
		healthChecks = append(healthChecks, &envoy_service_health.ClusterHealthCheck{
			ClusterName: names.GetLocalClusterName(inbound.WorkloadPort),
			LocalityEndpoints: []*envoy_service_health.LocalityEndpoints{{
				Endpoints: []*envoy_endpoint.Endpoint{{
					Address: &envoy_core.Address{
						Address: &envoy_core.Address_SocketAddress{
							SocketAddress: &envoy_core.SocketAddress{
								Address: inbound.WorkloadIP,
								PortSpecifier: &envoy_core.SocketAddress_PortValue{
									PortValue: inbound.WorkloadPort,
								},
							},
						},
					},
				}},
			}},
			HealthChecks: []*envoy_core.HealthCheck{
				{
					Timeout:            durationpb.New(g.config.Check.Timeout),
					Interval:           durationpb.New(g.config.Check.Interval),
					NoTrafficInterval:  durationpb.New(g.config.Check.NoTrafficInterval),
					HealthyThreshold:   &wrappers.UInt32Value{Value: g.config.Check.HealthyThreshold},
					UnhealthyThreshold: &wrappers.UInt32Value{Value: g.config.Check.UnhealthyThreshold},
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

	return cache.NewSnapshot(hcs), nil
}
