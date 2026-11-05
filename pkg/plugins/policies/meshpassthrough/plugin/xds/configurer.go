package xds

import (
	"slices"

	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"

	core_xds "github.com/kumahq/kuma/v3/pkg/core/xds"
	api "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshpassthrough/api/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/meshpassthrough/metadata"
	"github.com/kumahq/kuma/v3/pkg/util/pointer"
	xds_listeners_v3 "github.com/kumahq/kuma/v3/pkg/xds/envoy/listeners/v3"
)

type Configurer struct {
	APIVersion        core_xds.APIVersion
	InternalAddresses []core_xds.InternalAddress
	Conf              api.Conf
	IPv6Enabled       bool
	// DataplaneIPv6 is the address family of the proxy, it picks the family a domain
	// cluster resolves in. A domain chain is on both listeners and shares one cluster,
	// so the cluster follows the proxy rather than the listener.
	DataplaneIPv6 bool
}

func (c Configurer) Configure(ipv4 *envoy_listener.Listener, ipv6 *envoy_listener.Listener, rs *core_xds.ResourceSet) error {
	clustersAccumulator := map[string]Cluster{}
	filterChainMatches := GetOrderedMatchers(c.Conf)

	if hasIPv4Matches(filterChainMatches) {
		if err := c.configureListener(filterChainMatches, ipv4, clustersAccumulator, false, c.IPv6Enabled); err != nil {
			return err
		}
	}
	if hasIPv6Matches(filterChainMatches) {
		if err := c.configureListener(filterChainMatches, ipv6, clustersAccumulator, true, c.IPv6Enabled); err != nil {
			return err
		}
	}

	for name, cluster := range clustersAccumulator {
		config, err := CreateCluster(c.APIVersion, name, cluster, c.DataplaneIPv6)
		if err != nil {
			return err
		}
		rs.Add(&core_xds.Resource{
			Name:     config.GetName(),
			Origin:   metadata.OriginMeshPassthrough,
			Resource: config,
		})
	}
	return nil
}

func (c Configurer) configureListener(
	orderedFilterChainMatches []FilterChainMatch,
	listener *envoy_listener.Listener,
	clustersAccumulator map[string]Cluster,
	isIPv6 bool,
	ipv6Enabled bool,
) error {
	if listener == nil {
		return nil
	}
	listenerFiltersExcludedOnPorts := mysqlPorts(c.Conf)
	// remove default filter chain provided by `transparent_proxy_generator`
	listener.FilterChains = []*envoy_listener.FilterChain{}
	for _, matcher := range orderedFilterChainMatches {
		configurer := FilterChainConfigurer{
			APIVersion:        c.APIVersion,
			InternalAddresses: c.InternalAddresses,
			Protocol:          matcher.Protocol,
			Port:              matcher.Port,
			MatchType:         matcher.MatchType,
			MatchValue:        matcher.Value,
			Routes:            matcher.Routes,
			IsIPv6:            isIPv6,
			IPv6Enabled:       ipv6Enabled,
		}
		err := configurer.Configure(listener, clustersAccumulator)
		if err != nil {
			return err
		}
	}
	if err := c.configureListenerFilter(listener, listenerFiltersExcludedOnPorts); err != nil {
		return err
	}
	return nil
}

func (c Configurer) configureListenerFilter(listener *envoy_listener.Listener, listenerFiltersExcludedOnPorts []uint32) error {
	hasHttpInspector := false
	for _, filter := range listener.ListenerFilters {
		if filter.Name == xds_listeners_v3.HttpInspectorName {
			hasHttpInspector = true
		}
	}

	originalDstConfigurer := xds_listeners_v3.OriginalDstFilterConfigurer{}
	err := originalDstConfigurer.Configure(listener)
	if err != nil {
		return err
	}
	if err := xds_listeners_v3.EnsureTLSInspector(listener, listenerFiltersExcludedOnPorts...); err != nil {
		return err
	}
	if !hasHttpInspector {
		configurer := xds_listeners_v3.HTTPInspectorConfigurer{
			DisabledPorts: listenerFiltersExcludedOnPorts,
		}
		err = configurer.Configure(listener)
	}
	return err
}

// mysqlPorts lists the ports the inspectors are disabled on: the mysql server speaks
// first, so an inspector waiting for the client stalls the connection. A mysql match
// dropped for a tcp match on the same address and port still counts, the tcp chain
// carries its traffic.
func mysqlPorts(conf api.Conf) []uint32 {
	conflicts := api.FindConflicts(conf)
	ports := []uint32{}
	for i, match := range pointer.Deref(conf.AppendMatch) {
		if match.Protocol != api.MysqlProtocol || match.Port == nil || conflicts.IsInvalid(i) || slices.Contains(ports, *match.Port) {
			continue
		}
		ports = append(ports, *match.Port)
	}
	return ports
}

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
