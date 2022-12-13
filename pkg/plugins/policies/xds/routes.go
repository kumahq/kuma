package xds

import (
	envoy_route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"

	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/plugins/runtime/gateway/metadata"
)

type Routes struct {
	Gateway map[string]*envoy_route.RouteConfiguration
}

func GatherRoutes(rs *core_xds.ResourceSet) Routes {
	routes := Routes{
		Gateway: map[string]*envoy_route.RouteConfiguration{},
	}
	for _, res := range rs.Resources(envoy_resource.RouteType) {
		if res.Origin == metadata.OriginGateway {
			routeConfig := res.Resource.(*envoy_route.RouteConfiguration)
			routes.Gateway[routeConfig.Name] = routeConfig
		}
	}
	return routes
}
