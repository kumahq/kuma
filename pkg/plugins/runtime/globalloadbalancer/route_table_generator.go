package globalloadbalancer

import (
	"net/http"

	"google.golang.org/protobuf/types/known/wrapperspb"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/plugins/runtime/globalloadbalancer/route"
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
func GenerateVirtualHost(proxy *core_xds.Proxy, routes []*route.RouteBuilder, domains []string) (*envoy_virtual_hosts.VirtualHostBuilder, error) {
	datacenter := proxy.Dataplane.Spec.TagSet().Values(mesh_proto.KoyebDatacenterTag)[0]

	vh := envoy_virtual_hosts.NewVirtualHostBuilder(proxy.APIVersion, domains[0]).Configure(
		envoy_virtual_hosts.DomainNames(domains...),
		envoy_virtual_hosts.SetRequestHeader("X-Koyeb-GLB", datacenter),
		// TODO(nicoche) verify the protocol here
		envoy_virtual_hosts.Retry(&defaultRetryPolicy, "http"),
	)

	for _, rb := range routes {
		vh.Configure(route.VirtualHostRoute(rb))
	}

	return vh, nil
}

func GenerateFallbackVirtualHost(proxy *core_xds.Proxy) *envoy_virtual_hosts.VirtualHostBuilder {
	healthChecksRouteBuilder := &route.RouteBuilder{}
	healthChecksRouteBuilder.Configure(route.RouteMatchExactPath("/health"))
	healthChecksRouteBuilder.Configure(route.RouteActionDirectResponse(http.StatusOK, "OK"))

	fallbackRoute := GenerateNotFoundRouteBuilder()

	vh := envoy_virtual_hosts.NewVirtualHostBuilder(proxy.APIVersion, "fallback").Configure(
		envoy_virtual_hosts.DomainNames("*"),
		// TODO(nicoche) verify the protocol here
		envoy_virtual_hosts.Retry(&defaultRetryPolicy, "http"),

		// /health should always return "OK"
		route.VirtualHostRoute(healthChecksRouteBuilder),
		// any other route should fallback on the 404 page
		route.VirtualHostRoute(fallbackRoute),
	)

	return vh
}
