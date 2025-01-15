package xds

import (
	"fmt"

	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	rbac_config "github.com/envoyproxy/go-control-plane/envoy/config/rbac/v3"
	http_rbac "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/rbac/v3"
	envoy_hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	network_rbac "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/rbac/v3"
	matcherv3 "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_xds "github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/rules/subsetutils"
	policies_api "github.com/kumahq/kuma/pkg/plugins/policies/meshtrafficpermission/api/v1alpha1"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	util_xds "github.com/kumahq/kuma/pkg/util/xds"
	listeners_v3 "github.com/kumahq/kuma/pkg/xds/envoy/listeners/v3"
	tls "github.com/kumahq/kuma/pkg/xds/envoy/tls/v3"
)

type RBACConfigurer struct {
	StatsName string
	Rules     core_xds.Rules
	Mesh      string
}

func (c *RBACConfigurer) Configure(filterChain *envoy_listener.FilterChain) error {
	for idx, filter := range filterChain.Filters {
		if filter.GetName() == "envoy.filters.network.rbac" {
			// new MeshTrafficPermission takes over this filter chain,
			// it's safe to delete RBAC from old TrafficPermissions
			filterChain.Filters = append(filterChain.Filters[:idx], filterChain.Filters[idx+1:]...)
			break
		}
	}

	principalByAction := c.principalsByAction()
	rules := createRules(principalByAction)
	shadowRules := createShadowRules(principalByAction)

	// When the filter chain contains a `http_connection_manager`, it is more
	// appropriate to configure the `envoy.filters.http.rbac` filter within the
	// HTTP connection manager than to use the `envoy.filters.network.rbac`
	// filter at the listener level. One of the advantages of this approach is
	// that it can provide better envoy stats.
	for _, filter := range filterChain.Filters {
		if filter.GetName() == "envoy.filters.network.http_connection_manager" {
			return listeners_v3.UpdateHTTPConnectionManager(
				filterChain,
				httpRBACUpdater(rules, shadowRules),
			)
		}
	}

	return c.addRBACFilterToFilterChain(filterChain, rules, shadowRules)
}

func (c *RBACConfigurer) addRBACFilterToFilterChain(
	filterChain *envoy_listener.FilterChain,
	rules *rbac_config.RBAC,
	shadowRules *rbac_config.RBAC,
) error {
	typedConfig, err := util_proto.MarshalAnyDeterministic(&network_rbac.RBAC{
		// we include dot to change "inbound:127.0.0.1:21011rbac.allowed" metric
		// to "inbound:127.0.0.1:21011.rbac.allowed"
		StatPrefix:  fmt.Sprintf("%s.", util_xds.SanitizeMetric(c.StatsName)),
		Rules:       rules,
		ShadowRules: shadowRules,
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

func httpRBACUpdater(
	rules *rbac_config.RBAC,
	shadowRules *rbac_config.RBAC,
) func(manager *envoy_hcm.HttpConnectionManager) error {
	return func(manager *envoy_hcm.HttpConnectionManager) error {
		typedConfig, err := util_proto.MarshalAnyDeterministic(&http_rbac.RBAC{
			Rules:       rules,
			ShadowRules: shadowRules,
		})
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

type PrincipalMap map[policies_api.Action][]*rbac_config.Principal

func (c *RBACConfigurer) principalsByAction() PrincipalMap {
	pm := PrincipalMap{}
	for _, rule := range c.Rules {
		action := rule.Conf.(policies_api.Conf).Action
		pm[action] = append(pm[action], c.principalFromSubset(rule.Subset))
	}
	return pm
}

// createRules always returns not-nil result regardless the number of principals
func createRules(pm PrincipalMap) *rbac_config.RBAC {
	rules := &rbac_config.RBAC{
		Action:   rbac_config.RBAC_ALLOW,
		Policies: map[string]*rbac_config.Policy{},
	}

	principals := []*rbac_config.Principal{}
	principals = append(principals, pm[policies_api.Allow]...)
	principals = append(principals, pm[policies_api.AllowWithShadowDeny]...)

	if len(principals) != 0 {
		rules.Policies["MeshTrafficPermission"] = &rbac_config.Policy{
			Permissions: []*rbac_config.Permission{
				{
					Rule: &rbac_config.Permission_Any{
						Any: true,
					},
				},
			},
			Principals: principals,
		}
	}

	return rules
}

func createShadowRules(pm PrincipalMap) *rbac_config.RBAC {
	deny := pm[policies_api.AllowWithShadowDeny]
	if len(deny) == 0 {
		return nil
	}
	return &rbac_config.RBAC{
		Action: rbac_config.RBAC_DENY,
		Policies: map[string]*rbac_config.Policy{
			"MeshTrafficPermission": {
				Permissions: []*rbac_config.Permission{
					{
						Rule: &rbac_config.Permission_Any{
							Any: true,
						},
					},
				},
				Principals: deny,
			},
		},
	}
}

func (c *RBACConfigurer) principalFromSubset(ss subsetutils.Subset) *rbac_config.Principal {
	principals := []*rbac_config.Principal{}

	for _, t := range ss {
		var principalName *matcherv3.StringMatcher
		switch t.Key {
		case mesh_proto.ServiceTag:
			service := t.Value
			principalName = tls.ServiceSpiffeIDMatcher(c.Mesh, service)
		default:
			principalName = tls.KumaIDMatcher(t.Key, t.Value)
		}
		principal := &rbac_config.Principal{
			Identifier: &rbac_config.Principal_Authenticated_{
				Authenticated: &rbac_config.Principal_Authenticated{
					PrincipalName: principalName,
				},
			},
		}
		if t.Not {
			principal = c.not(principal)
		}
		principals = append(principals, principal)
	}

	switch len(principals) {
	case 0:
		return &rbac_config.Principal{
			Identifier: &rbac_config.Principal_Any{
				Any: true,
			},
		}
	case 1:
		return principals[0]
	default:
		return &rbac_config.Principal{
			Identifier: &rbac_config.Principal_AndIds{
				AndIds: &rbac_config.Principal_Set{
					Ids: principals,
				},
			},
		}
	}
}

func (c *RBACConfigurer) not(p *rbac_config.Principal) *rbac_config.Principal {
	return &rbac_config.Principal{
		Identifier: &rbac_config.Principal_NotId{
			NotId: p,
		},
	}
}
