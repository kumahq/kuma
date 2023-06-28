package routes

import (
	envoy_route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	v3 "github.com/kumahq/kuma/pkg/xds/envoy/routes/v3"
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

func CommonRouteConfiguration() RouteConfigurationBuilderOpt {
	return AddRouteConfigurationConfigurer(
		&v3.CommonRouteConfigurationConfigurer{})
}

func IgnorePortInHostMatching() RouteConfigurationBuilderOpt {
	return AddRouteConfigurationConfigurer(
		v3.RouteConfigurationConfigureFunc(func(rc *envoy_route.RouteConfiguration) error {
			rc.IgnorePortInHostMatching = true
			return nil
		}),
	)
}
