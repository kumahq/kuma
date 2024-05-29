package xds

import (
	"fmt"

	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"

	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	plugins_xds "github.com/kumahq/kuma/pkg/plugins/policies/core/xds"
	xds_listeners "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
	xds_routes "github.com/kumahq/kuma/pkg/xds/envoy/routes"
	xds_virtual_hosts "github.com/kumahq/kuma/pkg/xds/envoy/virtualhosts"
)

type FilterChainConfigurer struct {
	APIVersion core_xds.APIVersion
	Protocol   core_mesh.Protocol
	Port       uint32
	Routes     []Route
	IsIPv6     bool
}

func (c FilterChainConfigurer) Configure(listener *envoy_listener.Listener, clustersAccumulator map[string]bool) error {
	if listener == nil {
		return nil
	}
	if err := c.addFilterChainConfiguration(listener, clustersAccumulator); err != nil {
		return err
	}
	return nil
}

func (c FilterChainConfigurer) addFilterChainConfiguration(listener *envoy_listener.Listener, clustersAccumulator map[string]bool) error {
	switch c.Protocol {
	case core_mesh.ProtocolTCP:
		for _, route := range c.Routes {
			chainName := FilterChainName(route.Value, core_mesh.ProtocolTCP, c.Port)
			clusterName := ClusterName(route, core_mesh.ProtocolTCP, c.Port)
			split := plugins_xds.NewSplitBuilder().
				WithClusterName(clusterName).
				Build()
			filterChainBuilder := xds_listeners.NewFilterChainBuilder(c.APIVersion, chainName).
				Configure(xds_listeners.TCPProxy(clusterName, split)).
				Configure(xds_listeners.MatchTransportProtocol("raw_buffer"))
			c.configureAddressMatch(route, filterChainBuilder)
			if c.Port != 0 {
				filterChainBuilder.
					Configure(xds_listeners.MatchDestiantionPort(c.Port))
			}

			filterChain, err := filterChainBuilder.Build()
			if err != nil {
				return err
			}
			listener.FilterChains = append(listener.FilterChains, filterChain.(*envoy_listener.FilterChain))
			clustersAccumulator[clusterName] = true
		}
	case core_mesh.ProtocolTLS:
		for _, route := range c.Routes {
			chainName := FilterChainName(route.Value, core_mesh.ProtocolTLS, c.Port)
			clusterName := ClusterName(route, core_mesh.ProtocolTLS, c.Port)
			split := plugins_xds.NewSplitBuilder().
				WithClusterName(clusterName).
				Build()
			filterChainBuilder := xds_listeners.NewFilterChainBuilder(c.APIVersion, chainName).
				Configure(xds_listeners.TCPProxy(clusterName, split)).
				Configure(xds_listeners.MatchTransportProtocol("tls"))
			c.configureAddressMatch(route, filterChainBuilder)
			switch route.MatchType {
			case WildcardDomain, Domain:
				filterChainBuilder.
					Configure(xds_listeners.MatchServerNames(route.Value))
			}
			if c.Port != 0 {
				filterChainBuilder.
					Configure(xds_listeners.MatchDestiantionPort(c.Port))
			}

			filterChain, err := filterChainBuilder.Build()
			if err != nil {
				return err
			}
			listener.FilterChains = append(listener.FilterChains, filterChain.(*envoy_listener.FilterChain))
			clustersAccumulator[clusterName] = true
		}
	default:
		chainName := FilterChainName(
			string(c.Protocol),
			c.Protocol,
			c.Port,
		)
		routeBuilder := xds_routes.NewRouteConfigurationBuilder(c.APIVersion, chainName)
		for _, route := range c.Routes {
			switch route.MatchType {
			case Domain, WildcardDomain, IP, IPV6:
				clusterName := ClusterName(route, c.Protocol, c.Port)
				routeBuilder.Configure(xds_routes.VirtualHost(xds_virtual_hosts.NewVirtualHostBuilder(c.APIVersion, route.Value).
					Configure(
						xds_virtual_hosts.BasicRoute(ClusterName(route, c.Protocol, c.Port)),
						xds_virtual_hosts.DomainNames(route.Value),
					)))
				clustersAccumulator[clusterName] = true
			}
		}
		filterChain, err := xds_listeners.NewFilterChainBuilder(c.APIVersion, chainName).
			Configure(xds_listeners.MatchApplicationProtocols("http/1.1", "h2c")).
			Configure(xds_listeners.MatchTransportProtocol("raw_buffer")).
			Configure(xds_listeners.HttpConnectionManager(chainName, false)).
			Configure(xds_listeners.HttpStaticRoute(routeBuilder)).
			Build()
		if err != nil {
			return err
		}
		listener.FilterChains = append(listener.FilterChains, filterChain.(*envoy_listener.FilterChain))
	}
	return nil
}

func (c FilterChainConfigurer) configureAddressMatch(route Route, builder *xds_listeners.FilterChainBuilder) {
	if c.IsIPv6 {
		c.configureIPv6Match(route, builder)
	} else {
		c.configureIPv4Match(route, builder)
	}
}

func (c FilterChainConfigurer) configureIPv4Match(route Route, builder *xds_listeners.FilterChainBuilder) {
	switch route.MatchType {
	case CIDR:
		ip, mask := getIpAndMask(route.Value)
		builder.Configure(xds_listeners.MatchDestiantionAddressesRange(ip, mask))
	case IP:
		builder.Configure(xds_listeners.MatchDestiantionAddress(route.Value, false))
	}
}

func (c FilterChainConfigurer) configureIPv6Match(route Route, builder *xds_listeners.FilterChainBuilder) {
	switch route.MatchType {
	case CIDRV6:
		ip, mask := getIpAndMask(route.Value)
		builder.Configure(xds_listeners.MatchDestiantionAddressesRange(ip, mask))
	case IPV6:
		builder.Configure(xds_listeners.MatchDestiantionAddress(route.Value, true))
	}
}

func ClusterName(route Route, protocol core_mesh.Protocol, port uint32) string {
	if port == 0 {
		return fmt.Sprintf("meshpassthrough_%s_*", route.Value)
	}
	return fmt.Sprintf("meshpassthrough_%s_%d", route.Value, port)
}

func FilterChainName(name string, protocol core_mesh.Protocol, port uint32) string {
	displayPort := "*"
	if port != 0 {
		displayPort = fmt.Sprintf("%d", port)
	}
	if protocol == core_mesh.ProtocolTCP || protocol == core_mesh.ProtocolTLS {
		return fmt.Sprintf("meshpassthrough_%s_%s", name, displayPort)
	}
	return fmt.Sprintf("meshpassthrough_http_%s", displayPort)
}
