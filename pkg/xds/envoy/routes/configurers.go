package routes

import (
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	v2 "github.com/kumahq/kuma/pkg/xds/envoy/routes/v2"
	v3 "github.com/kumahq/kuma/pkg/xds/envoy/routes/v3"
)

func CommonVirtualHost(name string) VirtualHostBuilderOpt {
	return VirtualHostBuilderOptFunc(func(config *VirtualHostBuilderConfig) {
		config.AddV2(&v2.CommonVirtualHostConfigurer{
			Name: name,
		})
		config.AddV3(&v3.CommonVirtualHostConfigurer{
			Name: name,
		})
	})
}

func DefaultRoute(clusters ...envoy_common.Cluster) VirtualHostBuilderOpt {
	return VirtualHostBuilderOptFunc(func(config *VirtualHostBuilderConfig) {
		config.AddV2(&v2.DefaultRouteConfigurer{
			Clusters: clusters,
		})
		config.AddV3(&v3.DefaultRouteConfigurer{
			Clusters: clusters,
		})
	})
}

// Redirect for paths that match to matchPath returns 301 status code with new port and path
func Redirect(matchPath, newPath string, allowGetOnly bool, port uint32) VirtualHostBuilderOpt {
	return VirtualHostBuilderOptFunc(func(config *VirtualHostBuilderConfig) {
		config.AddV2(&v2.RedirectConfigurer{
			MatchPath:    matchPath,
			NewPath:      newPath,
			Port:         port,
			AllowGetOnly: allowGetOnly,
		})
		config.AddV3(&v3.RedirectConfigurer{
			MatchPath:    matchPath,
			NewPath:      newPath,
			Port:         port,
			AllowGetOnly: allowGetOnly,
		})
	})
}

// ResetTagsHeader adds x-kuma-tags header to the RequestHeadersToRemove list. x-kuma-tags header is planned to be used
// internally, so we don't want to expose it to the destination application.
func ResetTagsHeader() RouteConfigurationBuilderOpt {
	return RouteConfigurationBuilderOptFunc(func(config *RouteConfigurationBuilderConfig) {
		config.AddV2(&v2.ResetTagsHeaderConfigurer{})
		config.AddV3(&v3.ResetTagsHeaderConfigurer{})
	})
}

func TagsHeader(tags mesh_proto.MultiValueTagSet) RouteConfigurationBuilderOpt {
	return RouteConfigurationBuilderOptFunc(func(config *RouteConfigurationBuilderConfig) {
		config.AddV2(&v2.TagsHeaderConfigurer{
			Tags: tags,
		})
		config.AddV3(&v3.TagsHeaderConfigurer{
			Tags: tags,
		})
	})
}

func Route(matchPath, newPath, cluster string, allowGetOnly bool) VirtualHostBuilderOpt {
	return VirtualHostBuilderOptFunc(func(config *VirtualHostBuilderConfig) {
		config.AddV2(&v2.RoutesConfigurer{
			MatchPath:    matchPath,
			NewPath:      newPath,
			Cluster:      cluster,
			AllowGetOnly: allowGetOnly,
		})
		config.AddV3(&v3.RoutesConfigurer{
			MatchPath:    matchPath,
			NewPath:      newPath,
			Cluster:      cluster,
			AllowGetOnly: allowGetOnly,
		})
	})
}

func VirtualHost(builder *VirtualHostBuilder) RouteConfigurationBuilderOpt {
	return RouteConfigurationBuilderOptFunc(func(config *RouteConfigurationBuilderConfig) {
		config.AddV2(&RouteConfigurationVirtualHostConfigurerV2{
			builder: builder,
		})
		config.AddV3(&RouteConfigurationVirtualHostConfigurerV3{
			builder: builder,
		})
	})
}

func CommonRouteConfiguration(name string) RouteConfigurationBuilderOpt {
	return RouteConfigurationBuilderOptFunc(func(config *RouteConfigurationBuilderConfig) {
		config.AddV2(&v2.CommonRouteConfigurationConfigurer{
			Name: name,
		})
		config.AddV3(&v3.CommonRouteConfigurationConfigurer{
			Name: name,
		})
	})
}
