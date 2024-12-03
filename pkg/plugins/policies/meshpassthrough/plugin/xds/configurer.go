package xds

import (
	"slices"

	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"

	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshpassthrough/api/v1alpha1"
	xds_listeners_v3 "github.com/kumahq/kuma/pkg/xds/envoy/listeners/v3"
)

const (
	OriginMeshPassthrough = "meshpassthrough"
)

type Configurer struct {
	APIVersion core_xds.APIVersion
	Conf       api.Conf
}

func (c Configurer) Configure(ipv4 *envoy_listener.Listener, ipv6 *envoy_listener.Listener, rs *core_xds.ResourceSet) error {
	clustersAccumulator := map[string]core_mesh.Protocol{}
	filterChainMatches, err := GetOrderedMatchers(c.Conf)
	if err != nil {
		return err
	}
<<<<<<< HEAD
	if err := c.configureListener(orderedMatchers, ipv4, clustersAccumulator, false); err != nil {
		return err
	}
	if err := c.configureListener(orderedMatchers, ipv6, clustersAccumulator, true); err != nil {
		return err
=======

	if hasIPv4Matches(filterChainMatches) {
		if err := c.configureListener(filterChainMatches, ipv4, clustersAccumulator, false); err != nil {
			return err
		}
	}
	if hasIPv6Matches(filterChainMatches) {
		if err := c.configureListener(filterChainMatches, ipv6, clustersAccumulator, true); err != nil {
			return err
		}
>>>>>>> a5d4745e6 (fix(meshpassthrough): generate separate filter chain  when ip/cidr fo… (#12054))
	}
	for name, protocol := range clustersAccumulator {
		config, err := CreateCluster(c.APIVersion, name, protocol)
		if err != nil {
			return err
		}
		rs.Add(&core_xds.Resource{
			Name:     config.GetName(),
			Origin:   OriginMeshPassthrough,
			Resource: config,
		})
	}
	return nil
}

func (c Configurer) configureListener(
	orderedFilterChainMatches []FilterChainMatch,
	listener *envoy_listener.Listener,
	clustersAccumulator map[string]core_mesh.Protocol,
	isIPv6 bool,
) error {
	if listener == nil {
		return nil
	}
	// remove default filter chain provided by `transparent_proxy_generator`
	listener.FilterChains = []*envoy_listener.FilterChain{}
	for _, matcher := range orderedFilterChainMatches {
		configurer := FilterChainConfigurer{
			APIVersion: c.APIVersion,
			Protocol:   matcher.Protocol,
			Port:       matcher.Port,
			MatchType:  matcher.MatchType,
			MatchValue: matcher.Value,
			Routes:     matcher.Routes,
			IsIPv6:     isIPv6,
		}
		err := configurer.Configure(listener, clustersAccumulator)
		if err != nil {
			return err
		}
	}
	if err := c.configureListenerFilter(listener); err != nil {
		return err
	}
	return nil
}

func (c Configurer) configureListenerFilter(listener *envoy_listener.Listener) error {
	hasTlsInspector := false
	hasHttpInspector := false
	for _, filter := range listener.ListenerFilters {
		if filter.Name == xds_listeners_v3.TlsInspectorName {
			hasTlsInspector = true
		}
		if filter.Name == xds_listeners_v3.HttpInspectorName {
			hasHttpInspector = true
		}
	}
	var err error
	if !hasTlsInspector {
		configurer := xds_listeners_v3.TLSInspectorConfigurer{}
		err = configurer.Configure(listener)
	}
	if err != nil {
		return err
	}
	if !hasHttpInspector {
		configurer := xds_listeners_v3.HTTPInspectorConfigurer{}
		err = configurer.Configure(listener)
	}
	return err
}
<<<<<<< HEAD
=======

func hasIPv4Matches(orderedMatchers []FilterChainMatch) bool {
	for _, matcher := range orderedMatchers {
		if slices.Contains([]MatchType{Domain, WildcardDomain, CIDR, IP}, matcher.MatchType) {
			return true
		}
	}
	return false
}

func hasIPv6Matches(orderedMatchers []FilterChainMatch) bool {
	for _, matcher := range orderedMatchers {
		if slices.Contains([]MatchType{Domain, WildcardDomain, CIDRV6, IPV6}, matcher.MatchType) {
			return true
		}
	}
	return false
}
>>>>>>> a5d4745e6 (fix(meshpassthrough): generate separate filter chain  when ip/cidr fo… (#12054))
