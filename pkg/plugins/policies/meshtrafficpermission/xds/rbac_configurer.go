package xds

import (
	"fmt"

	matcher_config "github.com/cncf/xds/go/xds/type/matcher/v3"
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	rbac_config "github.com/envoyproxy/go-control-plane/envoy/config/rbac/v3"
	http_rbac "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/rbac/v3"
	envoy_hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	network_rbac "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/rbac/v3"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/kri"
	bldrs_matchers "github.com/kumahq/kuma/pkg/envoy/builders/xds/matchers"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/rules/common"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/rules/inbound"
	policies_api "github.com/kumahq/kuma/pkg/plugins/policies/meshtrafficpermission/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/util/pointer"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	util_xds "github.com/kumahq/kuma/pkg/util/xds"
	listeners_v3 "github.com/kumahq/kuma/pkg/xds/envoy/listeners/v3"
)

type RBACConfigurer struct {
	StatsName    string
	InboundRules []*inbound.Rule
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

	matcher, err := c.createMatcher()
	if err != nil {
		return err
	}
	shadowMatcher, err := c.createShadowMatcher()
	if err != nil {
		return err
	}

	// When the filter chain contains a `http_connection_manager`, it is more
	// appropriate to configure the `envoy.filters.http.rbac` filter within the
	// HTTP connection manager than to use the `envoy.filters.network.rbac`
	// filter at the listener level. One of the advantages of this approach is
	// that it can provide better envoy stats.
	for _, filter := range filterChain.Filters {
		if filter.GetName() == "envoy.filters.network.http_connection_manager" {
			return listeners_v3.UpdateHTTPConnectionManager(
				filterChain,
				rbacUpdater(matcher, shadowMatcher),
			)
		}
	}

	return c.addRBACFilterToFilterChain(filterChain, matcher, shadowMatcher)
}

func (c *RBACConfigurer) addRBACFilterToFilterChain(
	filterChain *envoy_listener.FilterChain,
	matcher *matcher_config.Matcher,
	shadowMatcher *matcher_config.Matcher,
) error {
	typedConfig, err := util_proto.MarshalAnyDeterministic(&network_rbac.RBAC{
		// we include dot to change "inbound:127.0.0.1:21011rbac.allowed" metric
		// to "inbound:127.0.0.1:21011.rbac.allowed"
		StatPrefix:    fmt.Sprintf("%s.", util_xds.SanitizeMetric(c.StatsName)),
		Matcher:       matcher,
		ShadowMatcher: shadowMatcher,
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

func rbacUpdater(
	matcher *matcher_config.Matcher,
	shadowMatcher *matcher_config.Matcher,
) func(manager *envoy_hcm.HttpConnectionManager) error {
	return func(manager *envoy_hcm.HttpConnectionManager) error {
		typedConfig, err := util_proto.MarshalAnyDeterministic(&http_rbac.RBAC{
			Matcher:       matcher,
			ShadowMatcher: shadowMatcher,
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

func (c *RBACConfigurer) createMatcher() (*matcher_config.Matcher, error) {
	var fieldMatchers []*matcher_config.Matcher_MatcherList_FieldMatcher
	for _, rule := range c.InboundRules {
		conf := rule.Conf.GetDefault().(policies_api.RuleConf)
		if conf.Deny != nil {
			denyMatchers, err := buildMatchers(pointer.Deref(conf.Deny), rbac_config.RBAC_DENY, rule.Origin)
			if err != nil {
				return nil, err
			}
			fieldMatchers = append(fieldMatchers, denyMatchers)
		}
		if conf.Allow != nil || conf.AllowWithShadowDeny != nil {
			allowMatchers, err := buildMatchers(append(pointer.Deref(conf.Allow), pointer.Deref(conf.AllowWithShadowDeny)...), rbac_config.RBAC_ALLOW, rule.Origin)
			if err != nil {
				return nil, err
			}
			fieldMatchers = append(fieldMatchers, allowMatchers)
		}
	}

	return bldrs_matchers.NewMatcherBuilder().
		Configure(bldrs_matchers.MatchersList(fieldMatchers)).
		Configure(bldrs_matchers.OnNoMatch(
			bldrs_matchers.NewOnMatch().Configure(bldrs_matchers.RbacAction(rbac_config.RBAC_DENY, "default")),
		)).
		Build()
}

func (c *RBACConfigurer) createShadowMatcher() (*matcher_config.Matcher, error) {
	var fieldMatchers []*matcher_config.Matcher_MatcherList_FieldMatcher
	for _, rule := range c.InboundRules {
		conf := rule.Conf.GetDefault().(policies_api.RuleConf)
		if conf.AllowWithShadowDeny == nil {
			continue
		}
		shadowDenyMatchers, err := buildMatchers(pointer.Deref(conf.AllowWithShadowDeny), rbac_config.RBAC_DENY, rule.Origin)
		if err != nil {
			return nil, err
		}
		fieldMatchers = append(fieldMatchers, shadowDenyMatchers)
	}

	return bldrs_matchers.NewMatcherBuilder().
		Configure(bldrs_matchers.MatchersList(fieldMatchers)).
		Configure(bldrs_matchers.OnNoMatch(
			bldrs_matchers.NewOnMatch().Configure(bldrs_matchers.RbacAction(rbac_config.RBAC_DENY, "default")),
		)).
		Build()
}

func buildMatchers(matches []common_api.Match, action rbac_config.RBAC_Action, origin common.Origin) (*matcher_config.Matcher_MatcherList_FieldMatcher, error) {
	if len(matches) == 0 {
		return nil, nil
	}
	return bldrs_matchers.NewFieldMatcherList().
		Configure(bldrs_matchers.Matches(
			matches,
			bldrs_matchers.NewOnMatch().Configure(bldrs_matchers.RbacAction(action, kri.FromResourceMeta(origin.Resource, policies_api.MeshTrafficPermissionType).String())),
		)).
		Build()
}
