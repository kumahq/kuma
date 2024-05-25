package xds

import (
	"errors"

	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"

	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	plugins_xds "github.com/kumahq/kuma/pkg/plugins/policies/core/xds"
	envoy_listeners "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
	envoy_routes "github.com/kumahq/kuma/pkg/xds/envoy/routes"
	envoy_virtual_hosts "github.com/kumahq/kuma/pkg/xds/envoy/virtualhosts"
)

type FilterChainConfigurer struct {
	Name       string
	Protocol   core_mesh.Protocol
	Routes     []Route
	APIVersion core_xds.APIVersion
}

func (c FilterChainConfigurer) Configure(listener *envoy_listener.Listener) error {
	if listener == nil {
		return nil
	}
	config, err := c.getFilterChainConfiguration()
	if err != nil {
		return err
	}
	listener.FilterChains = append(listener.FilterChains, config)
	return nil
}

func (c FilterChainConfigurer) getFilterChainConfiguration() (*envoy_listener.FilterChain, error) {
	switch c.Protocol {
	case core_mesh.ProtocolTCP:
		if len(c.Routes) != 1 {
			return nil, errors.New("there should be exactly one route")
		}
		split := plugins_xds.NewSplitBuilder().
			WithClusterName(c.Routes[0].ClusterName).
			WithExternalService(true).
			Build()
		tcpProxy := envoy_listeners.TCPProxy(c.Name, split)
		builder := envoy_listeners.NewFilterChainBuilder(c.APIVersion, c.Name)
		filterChain, err := builder.Configure(tcpProxy).Build()
		return filterChain.(*envoy_listener.FilterChain), err
	case core_mesh.ProtocolGRPC, core_mesh.ProtocolHTTP, core_mesh.ProtocolHTTP2:
		routeBuilder := envoy_routes.NewRouteConfigurationBuilder(c.APIVersion, c.Name)
		for _, route := range c.Routes {
			virtualHostBuilder := envoy_virtual_hosts.NewVirtualHostBuilder(c.APIVersion, route.Domain).
				Configure(
					envoy_virtual_hosts.BasicRoute(route.ClusterName),
					envoy_virtual_hosts.DomainNames(route.Domain),
				)
			routeBuilder.Configure(envoy_routes.VirtualHost(virtualHostBuilder))
		}
		filterChain, err := envoy_listeners.NewFilterChainBuilder(c.APIVersion, c.Name).
			Configure(envoy_listeners.HttpConnectionManager(c.Name, false)).
			Configure(envoy_listeners.HttpStaticRoute(routeBuilder)).
			Build()
		return filterChain.(*envoy_listener.FilterChain), err
	}
	return nil, nil
}
