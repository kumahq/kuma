package xds

import (
	"fmt"

	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	rbac_config "github.com/envoyproxy/go-control-plane/envoy/config/rbac/v3"
	http_rbac "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/rbac/v3"
	envoy_hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	network_rbac "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/rbac/v3"

	util_proto "github.com/kumahq/kuma/v3/pkg/util/proto"
	util_xds "github.com/kumahq/kuma/v3/pkg/util/xds"
	listeners_v3 "github.com/kumahq/kuma/v3/pkg/xds/envoy/listeners/v3"
)

// DenyAllRBACConfigurer installs an RBAC filter that permits nothing. It is the
// fallback for an inbound with no MeshTrafficPermission rules and no workload
// identity: with no principal to authorize, the inbound must fail closed.
type DenyAllRBACConfigurer struct {
	StatsName string
}

func (c *DenyAllRBACConfigurer) Configure(filterChain *envoy_listener.FilterChain) error {
	for idx, filter := range filterChain.Filters {
		if filter.GetName() == "envoy.filters.network.rbac" {
			// new MeshTrafficPermission takes over this filter chain,
			// it's safe to delete RBAC from old TrafficPermissions
			filterChain.Filters = append(filterChain.Filters[:idx], filterChain.Filters[idx+1:]...)
			break
		}
	}

	// An ALLOW action with no policies authorizes no principal at all.
	rules := &rbac_config.RBAC{
		Action:   rbac_config.RBAC_ALLOW,
		Policies: map[string]*rbac_config.Policy{},
	}

	// Inside an HCM the http filter yields better stats than a listener-level
	// network filter.
	for _, filter := range filterChain.Filters {
		if filter.GetName() == "envoy.filters.network.http_connection_manager" {
			return listeners_v3.UpdateHTTPConnectionManager(filterChain, httpRBACUpdater(rules))
		}
	}

	return c.addRBACFilterToFilterChain(filterChain, rules)
}

func (c *DenyAllRBACConfigurer) addRBACFilterToFilterChain(
	filterChain *envoy_listener.FilterChain,
	rules *rbac_config.RBAC,
) error {
	typedConfig, err := util_proto.MarshalAnyDeterministic(&network_rbac.RBAC{
		// we include dot to change "inbound:127.0.0.1:21011rbac.allowed" metric
		// to "inbound:127.0.0.1:21011.rbac.allowed"
		StatPrefix: fmt.Sprintf("%s.", util_xds.SanitizeMetric(c.StatsName)),
		Rules:      rules,
	})
	if err != nil {
		return err
	}

	filter := &envoy_listener.Filter{
		Name: "envoy.filters.network.rbac",
		ConfigType: &envoy_listener.Filter_TypedConfig{
			TypedConfig: typedConfig,
		},
	}

	// RBAC filter should be the first in the chain
	filterChain.Filters = append([]*envoy_listener.Filter{filter}, filterChain.Filters...)
	return nil
}

func httpRBACUpdater(rules *rbac_config.RBAC) func(manager *envoy_hcm.HttpConnectionManager) error {
	return func(manager *envoy_hcm.HttpConnectionManager) error {
		typedConfig, err := util_proto.MarshalAnyDeterministic(&http_rbac.RBAC{Rules: rules})
		if err != nil {
			return err
		}

		httpFilter := &envoy_hcm.HttpFilter{
			Name: "envoy.filters.http.rbac",
			ConfigType: &envoy_hcm.HttpFilter_TypedConfig{
				TypedConfig: typedConfig,
			},
		}

		manager.HttpFilters = append([]*envoy_hcm.HttpFilter{httpFilter}, manager.HttpFilters...)

		return nil
	}
}
