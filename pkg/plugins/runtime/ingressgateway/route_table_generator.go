package ingressgateway

import (
	"net/http"

	"google.golang.org/protobuf/types/known/wrapperspb"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/plugins/runtime/ingressgateway/route"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_virtual_hosts "github.com/kumahq/kuma/pkg/xds/envoy/virtualhosts"
)

var defaultRetryPolicy = core_mesh.RetryResource{
	Spec: &mesh_proto.Retry{
		Conf: &mesh_proto.Retry_Conf{
			Http: &mesh_proto.Retry_Conf_Http{
				NumRetries: &wrapperspb.UInt32Value{
					Value: 3,
				},
				// Required even if empty
				// RetriableStatusCodes: []uint32{},
			},
		},
	},
}

// GenerateVirtualHost generates xDS resources for the current route table.
// Note that the routes should be configured in the intended match order!
func GenerateVirtualHost(xdsCtx xds_context.Context, proxy *core_xds.Proxy, routes []*route.RouteBuilder) (*envoy_virtual_hosts.VirtualHostBuilder, error) {
	region := proxy.Dataplane.Spec.TagSet().Values(mesh_proto.KoyebRegionTag)[0]

	vh := envoy_virtual_hosts.NewVirtualHostBuilder(proxy.APIVersion, "wildcard").Configure(
		envoy_virtual_hosts.DomainNames("*"),
		envoy_virtual_hosts.SetRequestHeader("X-KOYEB-BACKEND", region),
		// TODO(nicoche) verify the protocol here
		envoy_virtual_hosts.Retry(&defaultRetryPolicy, "http"),
	)

	routes = append(routes, getHealthRouteBuilder())
	routes = append(routes, getFallbackRouteBuilder())

	for _, rb := range routes {
		vh.Configure(route.VirtualHostRoute(rb))
	}

	return vh, nil
}

func getHealthRouteBuilder() *route.RouteBuilder {
	routeBuilder := &route.RouteBuilder{}
	routeBuilder.Configure(route.RouteMatchPrefixPath("/health"))
	routeBuilder.Configure(route.RouteActionDirectResponse(http.StatusOK, "OK"))

	return routeBuilder
}

func getFallbackRouteBuilder() *route.RouteBuilder {
	routeBuilder := &route.RouteBuilder{}
	routeBuilder.Configure(route.RouteMatchPrefixPath("/"))
	routeBuilder.Configure(route.RouteActionDirectResponse(http.StatusBadGateway, "No supported routes for this path"))

	return routeBuilder
}
