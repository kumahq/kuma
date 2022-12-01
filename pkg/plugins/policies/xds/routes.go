package xds

import (
	envoy_route_api "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
)

func GatherRoutes(vh *envoy_route_api.VirtualHost, isInbound bool) []*envoy_route_api.Route {
	vhRoutes := []*envoy_route_api.Route{}
	vhRoutes = append(vhRoutes, vh.Routes...)
	return vhRoutes
}
