package virtualhosts

import (
	envoy_config_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
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

// Routes configures routes for inbound virtual hosts and keeps the legacy route
// timeout override of 0s so Envoy request timeouts stay disabled by default.
func Routes(routes envoy_common.Routes) VirtualHostBuilderOpt {
	return AddVirtualHostConfigurer(
		&RoutesConfigurer{
			Routes:                routes,
			ConfigureRouteTimeout: true,
		})
}

// OutboundRoutes configures routes for outbound and egress virtual hosts
// without setting a route timeout so MeshTimeout or Envoy defaults can apply.
func OutboundRoutes(routes envoy_common.Routes) VirtualHostBuilderOpt {
	return AddVirtualHostConfigurer(
		&RoutesConfigurer{
			Routes:                routes,
			ConfigureRouteTimeout: false,
		})
}

// Redirect for paths that match to matchPath returns 301 status code with new port and path
func Redirect(matchPath, newPath string, allowGetOnly bool, port uint32) VirtualHostBuilderOpt {
	return AddVirtualHostConfigurer(&RedirectConfigurer{
		MatchPath:    matchPath,
		NewPath:      newPath,
		Port:         port,
		AllowGetOnly: allowGetOnly,
	})
}

// RequireTLS specifies that this virtual host must only accept TLS connections.
func RequireTLS() VirtualHostBuilderOpt {
	return AddVirtualHostConfigurer(
		VirtualHostMustConfigureFunc(func(vh *envoy_config_route_v3.VirtualHost) {
			vh.RequireTls = envoy_config_route_v3.VirtualHost_ALL
		}),
	)
}

// SetResponseHeader unconditionally sets the named response header to the given value.
func SetResponseHeader(name string, value string) VirtualHostBuilderOpt {
	return AddVirtualHostConfigurer(
		VirtualHostMustConfigureFunc(func(vh *envoy_config_route_v3.VirtualHost) {
			hsts := &envoy_config_core_v3.HeaderValueOption{
				AppendAction: envoy_config_core_v3.HeaderValueOption_OVERWRITE_IF_EXISTS_OR_ADD,
				Header: &envoy_config_core_v3.HeaderValue{
					Key:   name,
					Value: value,
				},
			}

			vh.ResponseHeadersToAdd = append(vh.ResponseHeadersToAdd, hsts)
		}),
	)
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

func Route(matchPath string, newPath string, cluster string, allowGetOnly bool) VirtualHostBuilderOpt {
	return AddVirtualHostConfigurer(
		&VirtualHostRouteConfigurer{
			MatchPath:    matchPath,
			NewPath:      newPath,
			Cluster:      cluster,
			AllowGetOnly: allowGetOnly,
		})
}
