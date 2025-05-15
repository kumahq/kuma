package samples

import (
	"github.com/kumahq/kuma/pkg/core/kri"
	plugins_xds "github.com/kumahq/kuma/pkg/plugins/policies/core/xds"
	meshhttproute_api "github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	meshhttproute_xds "github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/xds"
	"github.com/kumahq/kuma/pkg/util/pointer"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
)

func MeshHttpOutboundWithSeveralRoutes(serviceName string) *meshhttproute_xds.HttpOutboundRouteConfigurer {
	anotherBackendMatch := meshhttproute_api.Match{
		Path: &meshhttproute_api.PathMatch{
			Type:  meshhttproute_api.Exact,
			Value: "/another-backend",
		},
		Method: pointer.To[meshhttproute_api.Method]("GET"),
	}
	rootPrefixMatch := meshhttproute_api.Match{
		Path: &meshhttproute_api.PathMatch{
			Type:  meshhttproute_api.PathPrefix,
			Value: "/",
		},
	}
	return &meshhttproute_xds.HttpOutboundRouteConfigurer{
		Service: serviceName,
		Routes: []meshhttproute_xds.OutboundRoute{
			{
				Split: []envoy_common.Split{
					plugins_xds.NewSplitBuilder().WithClusterName(serviceName).WithWeight(100).Build(),
				},
				Match: anotherBackendMatch,
				Name:  string(meshhttproute_api.HashMatches([]meshhttproute_api.Match{anotherBackendMatch})),
			},
			{
				Split: []envoy_common.Split{
					plugins_xds.NewSplitBuilder().WithClusterName(serviceName).WithWeight(100).Build(),
				},
				Match: rootPrefixMatch,
				Name:  string(meshhttproute_api.HashMatches([]meshhttproute_api.Match{rootPrefixMatch})),
			},
		},
		DpTags: map[string]map[string]bool{
			"kuma.io/service": {
				"web": true,
			},
		},
	}
}

func RealMeshHTTPRouteOutboundRoutes(serviceName string, meshHTTPRoute kri.Identifier) *meshhttproute_xds.HttpOutboundRouteConfigurer {
	anotherBackendMatch := meshhttproute_api.Match{
		Path: &meshhttproute_api.PathMatch{
			Type:  meshhttproute_api.Exact,
			Value: "/another-backend",
		},
		Method: pointer.To[meshhttproute_api.Method]("GET"),
	}
	rootPrefixMatch := meshhttproute_api.Match{
		Path: &meshhttproute_api.PathMatch{
			Type:  meshhttproute_api.PathPrefix,
			Value: "/",
		},
	}
	return &meshhttproute_xds.HttpOutboundRouteConfigurer{
		Service: serviceName,
		Routes: []meshhttproute_xds.OutboundRoute{
			{
				Split: []envoy_common.Split{
					plugins_xds.NewSplitBuilder().WithClusterName(serviceName).WithWeight(100).Build(),
				},
				Match: anotherBackendMatch,
				Name:  meshHTTPRoute.String(),
			},
			{
				Split: []envoy_common.Split{
					plugins_xds.NewSplitBuilder().WithClusterName(serviceName).WithWeight(100).Build(),
				},
				Match: rootPrefixMatch,
				Name:  string(meshhttproute_api.HashMatches([]meshhttproute_api.Match{rootPrefixMatch})),
			},
		},
		DpTags: map[string]map[string]bool{
			"kuma.io/service": {
				"web": true,
			},
		},
	}
}

func MeshHttpOutboudWithSingleRoute(serviceName string) *meshhttproute_xds.HttpOutboundRouteConfigurer {
	rootPrefixMatch := meshhttproute_api.Match{
		Path: &meshhttproute_api.PathMatch{
			Type:  meshhttproute_api.PathPrefix,
			Value: "/",
		},
	}
	return &meshhttproute_xds.HttpOutboundRouteConfigurer{
		Service: serviceName,
		Routes: []meshhttproute_xds.OutboundRoute{
			{
				Split: []envoy_common.Split{
					plugins_xds.NewSplitBuilder().WithClusterName(serviceName).WithWeight(100).Build(),
				},
				Match: rootPrefixMatch,
				Name:  string(meshhttproute_api.HashMatches([]meshhttproute_api.Match{rootPrefixMatch})),
			},
		},
		DpTags: map[string]map[string]bool{
			"kuma.io/service": {
				"web": true,
			},
		},
	}
}
