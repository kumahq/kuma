package xds

import (
	"fmt"
	"net/http"

	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"

	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	plugins_xds "github.com/kumahq/kuma/pkg/plugins/policies/core/xds"
	xds_listeners "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
	xds_routes "github.com/kumahq/kuma/pkg/xds/envoy/routes"
	xds_virtual_hosts "github.com/kumahq/kuma/pkg/xds/envoy/virtualhosts"
)

const NoMatchMsg = "This response comes from Kuma Sidecar. No routes matched this domain - check configuration of your MeshPassthrough policy.\n"

type FilterChainConfigurer struct {
	APIVersion core_xds.APIVersion
	Protocol   core_mesh.Protocol
	Port       uint32
	MatchType  MatchType
	MatchValue string
	Routes     []Route // should be set only for http and domains/wildcard
	IsIPv6     bool
}

func (c FilterChainConfigurer) Configure(listener *envoy_listener.Listener, clustersAccumulator map[string]core_mesh.Protocol) error {
	if listener == nil {
		return nil
	}
	if err := c.addFilterChainConfiguration(listener, clustersAccumulator); err != nil {
		return err
	}
	return nil
}

func (c FilterChainConfigurer) addFilterChainConfiguration(listener *envoy_listener.Listener, clustersAccumulator map[string]core_mesh.Protocol) error {
	switch c.Protocol {
	case core_mesh.ProtocolTCP:
		if err := c.configureTcpFilterChain(listener, clustersAccumulator); err != nil {
			return err
		}
	case core_mesh.ProtocolTLS:
		if err := c.configureTlsFilterChain(listener, clustersAccumulator); err != nil {
			return err
		}
	default:
		switch c.MatchType {
		case Domain, WildcardDomain:
			routeFn := func(routeBuilder *xds_routes.RouteConfigurationBuilder, clustersAccumulator map[string]core_mesh.Protocol) {
				for _, route := range c.Routes {
					if c.IsIPv6 && route.MatchType == IP {
						continue
					}
					if !c.IsIPv6 && route.MatchType == IPV6 {
						continue
					}
					switch route.MatchType {
					case Domain, WildcardDomain, IP, IPV6:
						domains := []string{route.Value}
						// based on the RFC, the host header might include a port, so we add another entry with the port defined
						if route.MatchType == IP || route.MatchType == Domain || route.MatchType == WildcardDomain {
							domains = append(domains, fmt.Sprintf("%s:%d", route.Value, c.Port))
						}
						if route.MatchType == IPV6 {
							domains = append(domains, fmt.Sprintf("[%s]", route.Value), fmt.Sprintf("[%s]:%d", route.Value, c.Port))
						}
						clusterName := ClusterName(route.Value, c.Protocol, c.Port)
						routeBuilder.Configure(xds_routes.VirtualHost(xds_virtual_hosts.NewVirtualHostBuilder(c.APIVersion, route.Value).
							Configure(
								xds_virtual_hosts.BasicRoute(clusterName),
								xds_virtual_hosts.DomainNames(domains...),
							)))
						clustersAccumulator[clusterName] = c.Protocol
					}
				}
				routeBuilder.Configure(xds_routes.VirtualHost(xds_virtual_hosts.NewVirtualHostBuilder(c.APIVersion, "no_match").
					Configure(
						xds_virtual_hosts.DirectResponseRoute(http.StatusServiceUnavailable, NoMatchMsg),
					)))
			}
			filterChainName := FilterChainName(string(c.Protocol), c.Protocol, c.Port)
			if err := c.configureHttpFilterChain(listener, routeFn, clustersAccumulator, filterChainName); err != nil {
				return err
			}
		default:
			routeFn := func(routeBuilder *xds_routes.RouteConfigurationBuilder, clustersAccumulator map[string]core_mesh.Protocol) {
				clusterName := ClusterName(c.MatchValue, c.Protocol, c.Port)
				routeBuilder.Configure(xds_routes.VirtualHost(xds_virtual_hosts.NewVirtualHostBuilder(c.APIVersion, c.MatchValue).
					Configure(
						xds_virtual_hosts.BasicRoute(clusterName),
						xds_virtual_hosts.DomainNames("*"),
					)))
				clustersAccumulator[clusterName] = c.Protocol
			}
			filterChainName := FilterChainName(c.MatchValue, c.Protocol, c.Port)
			if err := c.configureHttpFilterChain(listener, routeFn, clustersAccumulator, filterChainName); err != nil {
				return err
			}
		}
	}
	return nil
}

func (c FilterChainConfigurer) configureHttpFilterChain(
	listener *envoy_listener.Listener,
	routeFn func(routeBuilder *xds_routes.RouteConfigurationBuilder, clustersAccumulator map[string]core_mesh.Protocol),
	clustersAccumulator map[string]core_mesh.Protocol,
	filterChainName string,
) error {
	if c.IsIPv6 && (c.MatchType == IP || c.MatchType == CIDR) {
		return nil
	}
	if !c.IsIPv6 && (c.MatchType == IPV6 || c.MatchType == CIDRV6) {
		return nil
	}
	routeBuilder := xds_routes.NewRouteConfigurationBuilder(c.APIVersion, filterChainName)
	routeFn(routeBuilder, clustersAccumulator)
	filterChainBuilder := xds_listeners.NewFilterChainBuilder(c.APIVersion, filterChainName).
		Configure(xds_listeners.MatchApplicationProtocols("http/1.1", "h2c")).
		Configure(xds_listeners.MatchTransportProtocol("raw_buffer")).
		Configure(xds_listeners.HttpConnectionManager(filterChainName, false)).
		Configure(xds_listeners.HttpStaticRoute(routeBuilder))
	c.configureAddressMatch(filterChainBuilder)
	if c.Port != 0 {
		filterChainBuilder.
			Configure(xds_listeners.MatchDestiantionPort(c.Port))
	}

	filterChain, err := filterChainBuilder.Build()
	if err != nil {
		return err
	}
	listener.FilterChains = append(listener.FilterChains, filterChain.(*envoy_listener.FilterChain))
	return nil
}

func (c FilterChainConfigurer) configureTcpFilterChain(listener *envoy_listener.Listener, clustersAccumulator map[string]core_mesh.Protocol) error {
	if c.IsIPv6 && (c.MatchType == IP || c.MatchType == CIDR) {
		return nil
	}
	if !c.IsIPv6 && (c.MatchType == IPV6 || c.MatchType == CIDRV6) {
		return nil
	}
	chainName := FilterChainName(c.MatchValue, core_mesh.ProtocolTCP, c.Port)
	clusterName := ClusterName(c.MatchValue, core_mesh.ProtocolTCP, c.Port)
	split := plugins_xds.NewSplitBuilder().
		WithClusterName(clusterName).
		Build()
	filterChainBuilder := xds_listeners.NewFilterChainBuilder(c.APIVersion, chainName).
		Configure(xds_listeners.TCPProxy(clusterName, split)).
		Configure(xds_listeners.MatchTransportProtocol("raw_buffer"))
	c.configureAddressMatch(filterChainBuilder)
	if c.Port != 0 {
		filterChainBuilder.
			Configure(xds_listeners.MatchDestiantionPort(c.Port))
	}

	filterChain, err := filterChainBuilder.Build()
	if err != nil {
		return err
	}
	listener.FilterChains = append(listener.FilterChains, filterChain.(*envoy_listener.FilterChain))
	clustersAccumulator[clusterName] = c.Protocol
	return nil
}

func (c FilterChainConfigurer) configureTlsFilterChain(listener *envoy_listener.Listener, clustersAccumulator map[string]core_mesh.Protocol) error {
	if c.IsIPv6 && (c.MatchType == IP || c.MatchType == CIDR) {
		return nil
	}
	if !c.IsIPv6 && (c.MatchType == IPV6 || c.MatchType == CIDRV6) {
		return nil
	}
	chainName := FilterChainName(c.MatchValue, core_mesh.ProtocolTLS, c.Port)
	clusterName := ClusterName(c.MatchValue, core_mesh.ProtocolTLS, c.Port)
	split := plugins_xds.NewSplitBuilder().
		WithClusterName(clusterName).
		Build()
	filterChainBuilder := xds_listeners.NewFilterChainBuilder(c.APIVersion, chainName).
		Configure(xds_listeners.TCPProxy(clusterName, split)).
		Configure(xds_listeners.MatchTransportProtocol("tls"))
	c.configureAddressMatch(filterChainBuilder)
	switch c.MatchType {
	case WildcardDomain, Domain:
		filterChainBuilder.
			Configure(xds_listeners.MatchServerNames(c.MatchValue))
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
	clustersAccumulator[clusterName] = c.Protocol
	return nil
}

func (c FilterChainConfigurer) configureAddressMatch(builder *xds_listeners.FilterChainBuilder) {
	switch c.MatchType {
	case CIDR, CIDRV6:
		ip, mask := getIpAndMask(c.MatchValue)
		builder.Configure(xds_listeners.MatchDestiantionAddressesRange(ip, mask))
	case IP, IPV6:
		builder.Configure(xds_listeners.MatchDestiantionAddress(c.MatchValue, c.IsIPv6))
	}
}

func ClusterName(matchValue string, protocol core_mesh.Protocol, port uint32) string {
	if port == 0 {
		return fmt.Sprintf("meshpassthrough_%s_*", matchValue)
	}
	return fmt.Sprintf("meshpassthrough_%s_%d", matchValue, port)
}

func FilterChainName(name string, protocol core_mesh.Protocol, port uint32) string {
	displayPort := "*"
	if port != 0 {
		displayPort = fmt.Sprintf("%d", port)
	}
	if name == string(protocol) {
		return fmt.Sprintf("meshpassthrough_%s_%s", name, displayPort)
	}
	return fmt.Sprintf("meshpassthrough_%s_%s_%s", protocol, name, displayPort)
}
