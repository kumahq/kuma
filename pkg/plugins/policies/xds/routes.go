package xds

import (
	envoy_route_api "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
)

func RouteExist(rule *core_xds.Rule, route *envoy_route_api.Route)bool{
	if len(rule.Subset) == 0 {
		return true
	}
	
	

	return false
}

//envoy_routes.TagsHeaderName, tags.MatchSourceRegex(rl)