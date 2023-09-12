package globalloadbalancer

import (
	"fmt"
	"sort"
	"strings"

	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/plugins/runtime/globalloadbalancer/route"
)

type pathRoute struct {
	path    string
	service *core_xds.KoyebService
}

type pathRoutes []*pathRoute

func GenerateRouteBuilders(proxy *core_xds.Proxy, koyebServices []*core_xds.KoyebService) ([]*route.RouteBuilder, error) {
	routeBuilders := []*route.RouteBuilder{}

	// Generate a slice with one path per slice.
	// We need to have one path per item because we want to configure the routing,
	// longest paths first. That is because envoy matches rules in order. So if we
	// have:
	// - (1) match prefix /
	// - (2) match prefix /boom
	// all requests would be caught by (1) when we'd want /boom/test to be caught
	// by (2).
	//
	// Hence, we want to lay out every single path for this set of services and
	// sort them all from longest to shortest
	pathRoutes := getSortedPathRoutes(koyebServices)

	for _, pathRoute := range pathRoutes {
		routeBuilders = append(routeBuilders, generateServiceRoutes(pathRoute.path, pathRoute.service)...)
	}

	// If someone does not define a catch all (/) path, then some requests have
	// nowhere to be routed to.
	// For example, consider a service with routes like /api/v1 and /api/v2,
	// then we do not know where to route requests going to /api/v3, / or /test.
	// In that case, add a catch all route for a pretty 404
	shortestPathDeclared := pathRoutes[len(pathRoutes)-1].path
	if shortestPathDeclared != "" && shortestPathDeclared != "/" {
		routeBuilders = append(routeBuilders, GenerateNotFoundRouteBuilder())
	}

	return routeBuilders, nil
}

func generateServiceRoutes(path string, service *core_xds.KoyebService) []*route.RouteBuilder {
	routeBuilders := []*route.RouteBuilder{}
	clusterName := fmt.Sprintf("aggr_service_%s", service.ID)

	// cf. https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/route/v3/route_components.proto#envoy-v3-api-field-config-route-v3-routeaction-prefix-rewrite
	//
	// Pay careful attention to the use of trailing slashes in the route’s match prefix value. Stripping a prefix from a path requires multiple Routes to handle all cases. For example, rewriting /prefix to / and /prefix/etc to /etc cannot be done in a single Route, as shown by the below config entries:
	//   - match:
	//       prefix: "/prefix/"
	//     route:
	//       prefix_rewrite: "/"
	//   - match:
	//       prefix: "/prefix"
	//     route:
	//       prefix_rewrite: "/"
	// Having above entries in the config, requests to /prefix will be stripped to /, while requests to /prefix/etc will be stripped to /etc.
	//
	// Hence we need to declare two distinct RouteAction, one for paths ending with a final '/' and one for paths without
	if !strings.HasSuffix(path, "/") {
		path = fmt.Sprint(path, "/")
	}

	generateRouteBuilder := func(path string) *route.RouteBuilder {
		routeBuilder := &route.RouteBuilder{}
		routeBuilder.Configure(route.RouteMatchPrefixPath(path))
		routeBuilder.Configure(route.RouteAddRequestHeader(route.RouteReplaceHeader(
			"X-Koyeb-Route",
			fmt.Sprintf("%s-%d_%s", service.ID, service.Port, service.DeploymentGroup),
		)))
		routeBuilder.Configure(route.RouteActionCluster(clusterName, false, "/"))

		return routeBuilder
	}

	routeBuilders = append(routeBuilders, generateRouteBuilder(path))

	// Now, defining a route for mydomain.com/something does not include requests that end with a trailing slash except if our path is
	// an empty string:
	// GET mydomain.com == GET mydomain.com/
	// GET mydomain.com/something != GET mydomain.com/something/
	path = path[:len(path)-1]
	if path != "" {
		routeBuilders = append(routeBuilders, generateRouteBuilder(path))
	}

	return routeBuilders
}

func GenerateNotFoundRouteBuilder() *route.RouteBuilder {
	clusterName := "not_found"

	routeBuilder := &route.RouteBuilder{}
	routeBuilder.Configure(route.RouteMatchPrefixPath("/"))
	routeBuilder.Configure(route.RouteActionCluster(clusterName, true, "/cloudflare-404/"))

	return routeBuilder
}

func getSortedPathRoutes(koyebServices []*core_xds.KoyebService) pathRoutes {
	pathRoutes := pathRoutes{}

	// Note that we expect the total number of paths for this set of services
	// to be relatively small (<2 in 90% 0f cases, < 10 in 99% of cases) so
	// we do not really care about those nested for loops.

	for _, service := range koyebServices {
		for _, path := range service.Paths {
			pathRoutes = append(pathRoutes, &pathRoute{
				path:    path,
				service: service,
			})
		}
	}

	sort.SliceStable(pathRoutes, func(i int, j int) bool {
		return len(pathRoutes[i].path) > len(pathRoutes[j].path)
	})

	return pathRoutes
}
