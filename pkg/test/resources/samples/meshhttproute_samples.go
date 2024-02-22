package samples

import (
	plugins_xds "github.com/kumahq/kuma/pkg/plugins/policies/core/xds"
	meshhttproute_api "github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	meshhttproute_xds "github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/xds"
	"github.com/kumahq/kuma/pkg/util/pointer"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
)

func MeshHttpOutboundRouteWithSeveralRoutes(serviceName string) *meshhttproute_xds.HttpOutboundRouteConfigurer {
	return &meshhttproute_xds.HttpOutboundRouteConfigurer{
		Service: serviceName,
		Routes: []meshhttproute_xds.OutboundRoute{
			{
				Split: []envoy_common.Split{
					plugins_xds.NewSplitBuilder().WithClusterName(serviceName).WithWeight(100).Build(),
				},
				Matches: []meshhttproute_api.Match{
					{
						Path: &meshhttproute_api.PathMatch{
							Type:  meshhttproute_api.Exact,
							Value: "/another-backend",
						},
						Method: pointer.To[meshhttproute_api.Method]("GET"),
					},
				},
			},
			{
				Split: []envoy_common.Split{
					plugins_xds.NewSplitBuilder().WithClusterName(serviceName).WithWeight(100).Build(),
				},
				Matches: []meshhttproute_api.Match{
					{
						Path: &meshhttproute_api.PathMatch{
							Type:  meshhttproute_api.PathPrefix,
							Value: "/",
						},
					},
				},
			},
		},
		DpTags: map[string]map[string]bool{
			"kuma.io/service": {
				"web": true,
			},
		},
	}
}

func MeshHttpOutboundRouteWithSingleRoute(serviceName string) *meshhttproute_xds.HttpOutboundRouteConfigurer {
	return &meshhttproute_xds.HttpOutboundRouteConfigurer{
		Service: serviceName,
		Routes: []meshhttproute_xds.OutboundRoute{
			{
				Split: []envoy_common.Split{
					plugins_xds.NewSplitBuilder().WithClusterName(serviceName).WithWeight(100).Build(),
				},
				Matches: []meshhttproute_api.Match{
					{
						Path: &meshhttproute_api.PathMatch{
							Type:  meshhttproute_api.PathPrefix,
							Value: "/",
						},
					},
				},
			},
		},
		DpTags: map[string]map[string]bool{
			"kuma.io/service": {
				"web": true,
			},
		},
	}
}
