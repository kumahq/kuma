package gateway

import (
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/plugins/runtime/gateway/match"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
)

func filterGatewayRoutes(in []model.Resource, accept func(resource *core_mesh.GatewayRouteResource) bool) []*core_mesh.GatewayRouteResource {
	routes := make([]*core_mesh.GatewayRouteResource, 0, len(in))

	for _, r := range in {
		if trafficRoute, ok := r.(*core_mesh.GatewayRouteResource); ok {
			if accept(trafficRoute) {
				routes = append(routes, trafficRoute)
			}
		}
	}

	return routes
}

// GatewayRouteGenerator generates Kuma gateway routes from GatewayRoute resources.
type GatewayRouteGenerator struct {
}

func (*GatewayRouteGenerator) SupportsProtocol(p mesh_proto.Gateway_Listener_Protocol) bool {
	return p == mesh_proto.Gateway_Listener_HTTP || p == mesh_proto.Gateway_Listener_HTTPS
}

func (*GatewayRouteGenerator) GenerateHost(ctx xds_context.Context, info *GatewayResourceInfo) (*core_xds.ResourceSet, error) {
	gatewayRoutes := filterGatewayRoutes(info.Host.Routes, func(route *core_mesh.GatewayRouteResource) bool {
		// Wildcard virtual host accepts all routes.
		if info.Host.Hostname == WildcardHostname {
			return true
		}

		// If the route has no hostnames, it matches all virtualhosts.
		names := route.Spec.GetConf().GetHttp().GetHostnames()
		if len(names) == 0 {
			return true
		}

		// Otherwise, match the virtualhost name to the route names.
		return match.Hostnames(info.Host.Hostname, names...)
	})

	if len(gatewayRoutes) == 0 {
		return nil, nil
	}

	resources := ResourceAggregator{}

	log.V(1).Info("applying merged traffic routes",
		"listener-port", info.Listener.Port,
		"listener-name", info.Listener.ResourceName,
	)

	return resources.Get(), nil
}
