package virtualhosts

import (
	envoy_config_route_v3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"

	envoy_common "github.com/kumahq/kuma/v3/pkg/xds/envoy"
)

func DomainNames(domainNames ...string) VirtualHostBuilderOpt {
	if len(domainNames) == 0 {
		return VirtualHostBuilderOptFunc(nil)
	}

	return AddVirtualHostConfigurer(
		VirtualHostMustConfigureFunc(func(vh *envoy_config_route_v3.VirtualHost) {
			vh.Domains = domainNames
		}),
	)
}

// CatchAllRoute configures the single prefix "/" route of an inbound virtual host.
func CatchAllRoute(cluster envoy_common.Cluster) VirtualHostBuilderOpt {
	return AddVirtualHostConfigurer(&CatchAllRouteConfigurer{Cluster: cluster})
}

func BasicRoute(cluster string) VirtualHostBuilderOpt {
	return AddVirtualHostConfigurer(
		&VirtualHostBasicRouteConfigurer{
			Cluster: cluster,
		})
}

func DirectResponseRoute(status uint32, responseMsg string) VirtualHostBuilderOpt {
	return AddVirtualHostConfigurer(
		&VirtualHostDirectResponseRouteConfigurer{
			status:      status,
			responseMsg: responseMsg,
		})
}
