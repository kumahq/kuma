package xds

import (
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"

	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshpassthrough/api/v1alpha1"
	xds_listeners "github.com/kumahq/kuma/pkg/xds/envoy/listeners/v3"
)

const (
	OriginMeshPassthrough = "meshpassthrough"
)

type Configurer struct {
	APIVersion core_xds.APIVersion
	Conf       api.Conf
}

func (c Configurer) Configure(ipv4 *envoy_listener.Listener, ipv6 *envoy_listener.Listener, rs *core_xds.ResourceSet) error {
	clustersAccumulator := map[string]bool{}
	orderedMatchers, _ := GetOrderedMatchers(c.Conf)
	if err := c.configureListener(orderedMatchers, ipv4, clustersAccumulator, false); err != nil {
		return err
	}
	if err := c.configureListener(orderedMatchers, ipv6, clustersAccumulator, true); err != nil {
		return err
	}
	for name := range clustersAccumulator {
		config, err := CreateCluster(c.APIVersion, name)
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
	orderedMatchers []FilterChainMatcher,
	listener *envoy_listener.Listener,
	clustersAccumulator map[string]bool,
	isIPv6 bool,
) error {
	if listener == nil {
		return nil
	}
	// remove default filter chain provided by `transparent_proxy_generator`
	listener.FilterChains = []*envoy_listener.FilterChain{}
	for _, matcher := range orderedMatchers {
		configurer := FilterChainConfigurer{
			APIVersion: c.APIVersion,
			Protocol:   matcher.Protocol,
			Port:       matcher.Port,
			Routes:     matcher.Routes,
			IsIPv6:     isIPv6,
		}
		err := configurer.Configure(listener, clustersAccumulator)
		if err != nil {
			return err
		}
	}
	return c.configureListenerFilter(listener)
}

func (c Configurer) configureListenerFilter(listener *envoy_listener.Listener) error {
	hasTlsInspector := false
	hasHttpInspector := false
	for _, filter := range listener.ListenerFilters {
		if filter.Name == xds_listeners.TlsInspectorName {
			hasTlsInspector = true
		}
		if filter.Name == xds_listeners.HttpInspectorName {
			hasHttpInspector = true
		}
	}
	var err error
	if !hasTlsInspector {
		configurer := xds_listeners.TLSInspectorConfigurer{}
		err = configurer.Configure(listener)
	}
	if err != nil {
		return err
	}
	if !hasHttpInspector {
		configurer := xds_listeners.HTTPInspectorConfigurer{}
		err = configurer.Configure(listener)
	}
	return err
}
