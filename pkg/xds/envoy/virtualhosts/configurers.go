package virtualhosts

import (
	envoy_route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"

	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	envoy_routes "github.com/kumahq/kuma/pkg/xds/envoy/routes/v3"
)

func VirtualHostCommon(name string) VirtualHostConfigurer {
	return VirtualHostMustConfigureFunc(func(vh *envoy_route.VirtualHost) {
		vh.Name = name
		vh.Domains = []string{"*"}
	})
}

func VirtualHostDomains(domains ...string) VirtualHostConfigurer {
	if len(domains) == 0 {
		return VirtualHostConfigureFunc(nil)
	}

	return VirtualHostMustConfigureFunc(func(vh *envoy_route.VirtualHost) {
		vh.Domains = domains
	})
}

func VirtualHostRoutes(routes envoy_common.Routes) VirtualHostConfigurer {
	return VirtualHostConfigurer(&envoy_routes.RoutesConfigurer{Routes: routes})
}
