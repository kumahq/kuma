package v1alpha1

import (
	"context"

	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	plugin_gateway "github.com/kumahq/kuma/pkg/plugins/runtime/gateway"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_listeners "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
	envoy_routes "github.com/kumahq/kuma/pkg/xds/envoy/routes"
)

var FilterChainGenerators = map[mesh_proto.MeshGateway_Listener_Protocol]plugin_gateway.FilterChainGenerator{
	mesh_proto.MeshGateway_Listener_HTTP:  &plugin_gateway.HTTPFilterChainGenerator{},
	mesh_proto.MeshGateway_Listener_HTTPS: &plugin_gateway.HTTPSFilterChainGenerator{},
	mesh_proto.MeshGateway_Listener_TCP:   &plugin_gateway.TCPFilterChainGenerator{},
}

func generateGatewayListeners(
	ctx xds_context.Context,
	info plugin_gateway.GatewayListenerInfo,
) (*core_xds.ResourceSet, *plugin_gateway.RuntimeResoureLimitListener, error) {
	resources := core_xds.NewResourceSet()

	listenerBuilder, limit := plugin_gateway.GenerateListener(info)

	protocol := info.Listener.Protocol
	if info.Listener.CrossMesh {
		protocol = mesh_proto.MeshGateway_Listener_HTTPS
	}
	filterGen, found := FilterChainGenerators[protocol]
	if !found {
		return resources, limit, nil
	}

	res, filterChainBuilders, err := filterGen.Generate(ctx, info)
	if err != nil {
		return nil, limit, err
	}
	resources.AddSet(res)

	for _, filterChainBuilder := range filterChainBuilders {
		listenerBuilder.Configure(envoy_listeners.FilterChain(filterChainBuilder))
	}

	res, err = plugin_gateway.BuildResourceSet(listenerBuilder)
	if err != nil {
		return nil, limit, errors.Wrapf(err, "failed to build listener resource")
	}
	resources.AddSet(res)

	return resources, limit, nil
}

func generateGatewayClusters(
	ctx context.Context,
	xdsCtx xds_context.Context,
	info plugin_gateway.GatewayListenerInfo,
) (*core_xds.ResourceSet, error) {
	resources := core_xds.NewResourceSet()

	gen := plugin_gateway.ClusterGenerator{Zone: xdsCtx.ControlPlane.Zone}
	for _, listenerHostname := range info.ListenerHostnames {
		for _, hostInfo := range listenerHostname.HostInfos {
			clusterRes, err := gen.GenerateClusters(ctx, xdsCtx, info, hostInfo.Entries(), hostInfo.Host.Tags)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to generate clusters for dataplane %q", info.Proxy.Id)
			}
			resources.AddSet(clusterRes)
		}
	}

	return resources, nil
}

func generateGatewayRoutes(
	ctx xds_context.Context, info plugin_gateway.GatewayListenerInfo,
) (*core_xds.ResourceSet, error) {
	listenerHostnames := info.ListenerHostnames
	switch info.Listener.Protocol {
	case mesh_proto.MeshGateway_Listener_HTTPS,
		mesh_proto.MeshGateway_Listener_HTTP:
	default:
		return nil, nil
	}

	resources := core_xds.NewResourceSet()
	// Make a pass over the generators for each virtual host.
	for _, hostInfos := range listenerHostnames {
		routeConfig := plugin_gateway.GenerateRouteConfig(
			info.Proxy,
			info.Listener.Protocol,
			hostInfos.EnvoyRouteName(info.Listener.EnvoyListenerName),
		)
		for _, hostInfo := range hostInfos.HostInfos {
			vh, err := plugin_gateway.GenerateVirtualHost(ctx, info, hostInfo.Host, hostInfo.Entries())
			if err != nil {
				return nil, err
			}

			routeConfig.Configure(envoy_routes.VirtualHost(vh))
		}
		res, err := plugin_gateway.BuildResourceSet(routeConfig)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to build route configuration resource")
		}
		resources.AddSet(res)
	}

	return resources, nil
}
