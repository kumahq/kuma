package routes

import (
	envoy_config_route_v3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	v3 "github.com/kumahq/kuma/pkg/xds/envoy/routes/v3"
	envoy_virtual_hosts "github.com/kumahq/kuma/pkg/xds/envoy/virtualhosts"
)

// ResetTagsHeader adds x-kuma-tags header to the RequestHeadersToRemove list. x-kuma-tags header is planned to be used
// internally, so we don't want to expose it to the destination application.
func ResetTagsHeader() RouteConfigurationBuilderOpt {
	return AddRouteConfigurationConfigurer(&v3.ResetTagsHeaderConfigurer{})
}

func TagsHeader(tags mesh_proto.MultiValueTagSet) RouteConfigurationBuilderOpt {
	return AddRouteConfigurationConfigurer(
		&v3.TagsHeaderConfigurer{
			Tags: tags,
		})
}

func VirtualHost(builder *envoy_virtual_hosts.VirtualHostBuilder) RouteConfigurationBuilderOpt {
	return AddRouteConfigurationConfigurer(
		v3.RouteConfigurationConfigureFunc(func(rc *envoy_config_route_v3.RouteConfiguration) error {
			virtualHost, err := builder.Build()
			if err != nil {
				return err
			}
			rc.VirtualHosts = append(rc.VirtualHosts, virtualHost.(*envoy_config_route_v3.VirtualHost))
			return nil
		}))
}

func CommonRouteConfiguration() RouteConfigurationBuilderOpt {
	return AddRouteConfigurationConfigurer(
		&v3.CommonRouteConfigurationConfigurer{})
}

func IgnorePortInHostMatching() RouteConfigurationBuilderOpt {
	return AddRouteConfigurationConfigurer(
		v3.RouteConfigurationConfigureFunc(func(rc *envoy_config_route_v3.RouteConfiguration) error {
			rc.IgnorePortInHostMatching = true
			return nil
		}),
	)
}
