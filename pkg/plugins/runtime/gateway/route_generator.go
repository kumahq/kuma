package gateway

import (
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/xds"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	envoy_routes "github.com/kumahq/kuma/pkg/xds/envoy/routes"
	"github.com/kumahq/kuma/pkg/xds/generator"
)

// RouteGenerator generates Kuma gateway routes. Not implemented.
type RouteGenerator struct {
	Resources manager.ReadOnlyResourceManager
}

var _ generator.ResourceGenerator = &RouteGenerator{}

func (r RouteGenerator) Generate(ctx xds_context.Context, proxy *xds.Proxy) (*xds.ResourceSet, error) {
	log.V(2).Info("Gateway route generation not implemented")
	return nil, nil
}

// DefaultRouteName is the well-known name of the default route configuration.
const DefaultRouteName = "default-route"

// DefaultRouteGenerator generates a default route that can be used to
// satisfy xDS consistency constraints.
type DefaultRouteGenerator struct{}

var _ generator.ResourceGenerator = DefaultRouteGenerator{}

func (DefaultRouteGenerator) Generate(ctx xds_context.Context, proxy *xds.Proxy) (*xds.ResourceSet, error) {
	vhost := envoy_routes.NewVirtualHostBuilder(envoy.APIV3)
	vhost.Configure(
		// Create a virtual host to match '*'.
		envoy_routes.CommonVirtualHost(DefaultRouteName),
	)

	routes := envoy_routes.NewRouteConfigurationBuilder(envoy.APIV3)
	routes.Configure(
		envoy_routes.CommonRouteConfiguration(DefaultRouteName),
		envoy_routes.ResetTagsHeader(),
		envoy_routes.VirtualHost(vhost),
	)

	return BuildResourceSet(routes)
}
