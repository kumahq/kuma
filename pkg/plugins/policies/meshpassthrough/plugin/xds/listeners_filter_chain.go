package xds

import (
	"errors"

	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"

	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	plugins_xds "github.com/kumahq/kuma/pkg/plugins/policies/core/xds"
	xds_listeners "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
	xds_routes "github.com/kumahq/kuma/pkg/xds/envoy/routes"
	xds_virtual_hosts "github.com/kumahq/kuma/pkg/xds/envoy/virtualhosts"
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
	case core_mesh.ProtocolTCP, core_mesh.ProtocolTLS:
		if len(c.Routes) != 1 {
			return nil, errors.New("there should be exactly one route")
		}
		split := plugins_xds.NewSplitBuilder().
			WithClusterName(c.Routes[0].ClusterName).
			WithExternalService(true).
			Build()
		tcpProxy := xds_listeners.TCPProxy(c.Name, split)
		builder := xds_listeners.NewFilterChainBuilder(c.APIVersion, c.Name)
		filterChain, err := builder.Configure(tcpProxy).Build()
		return filterChain.(*envoy_listener.FilterChain), err
	case core_mesh.ProtocolGRPC, core_mesh.ProtocolHTTP, core_mesh.ProtocolHTTP2:
		routeBuilder := xds_routes.NewRouteConfigurationBuilder(c.APIVersion, c.Name)
		for _, route := range c.Routes {
			virtualHostBuilder := xds_virtual_hosts.NewVirtualHostBuilder(c.APIVersion, route.Domain).
				Configure(
					xds_virtual_hosts.BasicRoute(route.ClusterName),
					xds_virtual_hosts.DomainNames(route.Domain),
				)
			routeBuilder.Configure(xds_routes.VirtualHost(virtualHostBuilder))
		}
		filterChain, err := xds_listeners.NewFilterChainBuilder(c.APIVersion, c.Name).
			Configure(xds_listeners.HttpConnectionManager(c.Name, false)).
			Configure(xds_listeners.HttpStaticRoute(routeBuilder)).
			Build()
		return filterChain.(*envoy_listener.FilterChain), err
	}
	return nil, nil
}
