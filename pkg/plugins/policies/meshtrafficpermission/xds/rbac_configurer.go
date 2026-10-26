package xds

import (
	"fmt"
	"slices"

	matcher_config "github.com/cncf/xds/go/xds/type/matcher/v3"
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	rbac_config "github.com/envoyproxy/go-control-plane/envoy/config/rbac/v3"
	http_rbac "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/rbac/v3"
	envoy_hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	network_rbac "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/rbac/v3"

	common_api "github.com/kumahq/kuma/v2/api/common/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/core/kri"
	bldrs_matchers "github.com/kumahq/kuma/v2/pkg/envoy/builders/xds/matchers"
	core_rules "github.com/kumahq/kuma/v2/pkg/plugins/policies/core/rules"
	"github.com/kumahq/kuma/v2/pkg/plugins/policies/core/rules/common"
	"github.com/kumahq/kuma/v2/pkg/plugins/policies/core/rules/inbound"
	"github.com/kumahq/kuma/v2/pkg/plugins/policies/core/rules/subsetutils"
	policies_api "github.com/kumahq/kuma/v2/pkg/plugins/policies/meshtrafficpermission/api/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/util/pointer"
	util_proto "github.com/kumahq/kuma/v2/pkg/util/proto"
	util_xds "github.com/kumahq/kuma/v2/pkg/util/xds"
	listeners_v3 "github.com/kumahq/kuma/v2/pkg/xds/envoy/listeners/v3"
)

type RBACConfigurer struct {
	StatsName    string
	InboundRules []*inbound.Rule
	// LegacyRules are 'from' rules merged into the same matcher as InboundRules
	LegacyRules core_rules.Rules
	Mesh        string
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
	return c.buildMatcher(
		func(conf policies_api.RuleConf) []common_api.Match {
			return pointer.Deref(conf.Deny)
		},
		func(conf policies_api.RuleConf) []common_api.Match {
			return slices.Concat(pointer.Deref(conf.Allow), pointer.Deref(conf.AllowWithShadowDeny))
		},
		[]policies_api.Action{policies_api.Deny},
		[]policies_api.Action{policies_api.Allow, policies_api.AllowWithShadowDeny},
	)
}

// createShadowMatcher mirrors createMatcher with allowWithShadowDeny turned
// into deny, so callers matched only by allow stay shadow-allowed.
func (c *RBACConfigurer) createShadowMatcher() (*matcher_config.Matcher, error) {
	hasShadowDeny := slices.ContainsFunc(c.InboundRules, func(rule *inbound.Rule) bool {
		return rule.Conf.(policies_api.RuleConf).AllowWithShadowDeny != nil
	})
	for _, rule := range c.LegacyRules {
		if pointer.Deref(rule.Conf.(policies_api.Conf).Action) == policies_api.AllowWithShadowDeny {
			hasShadowDeny = true
		}
	}
	if !hasShadowDeny {
		return nil, nil
	}

	return c.buildMatcher(
		func(conf policies_api.RuleConf) []common_api.Match {
			return slices.Concat(pointer.Deref(conf.Deny), pointer.Deref(conf.AllowWithShadowDeny))
		},
		func(conf policies_api.RuleConf) []common_api.Match {
			return pointer.Deref(conf.Allow)
		},
		[]policies_api.Action{policies_api.Deny, policies_api.AllowWithShadowDeny},
		[]policies_api.Action{policies_api.Allow},
	)
}

func (c *RBACConfigurer) buildMatcher(
	deny, allow func(policies_api.RuleConf) []common_api.Match,
	legacyDeny, legacyAllow []policies_api.Action,
) (*matcher_config.Matcher, error) {
	var fieldMatchers []*matcher_config.Matcher_MatcherList_FieldMatcher

	// Merged policies can arrive in a broad-to-specific order (for example a
	// mesh-wide allow before a listener-scoped deny). Envoy stops at the first
	// matching action, so all deny matchers need to be emitted before allow
	// matchers across the merged rule set.
	for _, rule := range c.InboundRules {
		denyMatchers, err := buildMatchers(deny(rule.Conf.(policies_api.RuleConf)), rbac_config.RBAC_DENY, rule.Origin)
		if err != nil {
			return nil, err
		}
		fieldMatchers = append(fieldMatchers, denyMatchers)
	}
	legacyDenyMatchers, err := c.legacyMatchers(rbac_config.RBAC_DENY, legacyDeny...)
	if err != nil {
		return nil, err
	}
	fieldMatchers = append(fieldMatchers, legacyDenyMatchers...)

	for _, rule := range c.InboundRules {
		allowMatchers, err := buildMatchers(allow(rule.Conf.(policies_api.RuleConf)), rbac_config.RBAC_ALLOW, rule.Origin)
		if err != nil {
			return nil, err
		}
		fieldMatchers = append(fieldMatchers, allowMatchers)
	}
	legacyAllowMatchers, err := c.legacyMatchers(rbac_config.RBAC_ALLOW, legacyAllow...)
	if err != nil {
		return nil, err
	}
	fieldMatchers = append(fieldMatchers, legacyAllowMatchers...)

	return bldrs_matchers.NewMatcherBuilder().
		Configure(bldrs_matchers.MatchersList(fieldMatchers)).
		Configure(bldrs_matchers.OnNoMatch(
			bldrs_matchers.NewOnMatch().Configure(bldrs_matchers.RbacAction(rbac_config.RBAC_DENY, "default")),
		)).
		Build()
}

// legacyMatchers translates 'from' rules with one of the given actions.
func (c *RBACConfigurer) legacyMatchers(rbacAction rbac_config.RBAC_Action, actions ...policies_api.Action) ([]*matcher_config.Matcher_MatcherList_FieldMatcher, error) {
	var fieldMatchers []*matcher_config.Matcher_MatcherList_FieldMatcher
	for _, rule := range c.LegacyRules {
		action := pointer.Deref(rule.Conf.(policies_api.Conf).Action)
		if !slices.Contains(actions, action) {
			continue
		}
		// A catch-all deny ('kind: Mesh' or negations only) is what OnNoMatch already does,
		// emitted before allows it would shadow every 'rules' allow.
		catchAll := !slices.ContainsFunc(rule.Subset, func(t subsetutils.Tag) bool { return !t.Not })
		if rbacAction == rbac_config.RBAC_DENY && catchAll {
			continue
		}
		predicate, err := predicateFromSubset(c.Mesh, rule.Subset)
		if err != nil {
			return nil, err
		}
		name := "MeshTrafficPermission"
		if len(rule.Origin) > 0 {
			// origins are ordered by specificity and the last one decides the merged action
			name = kri.FromResourceMeta(rule.Origin[len(rule.Origin)-1], policies_api.MeshTrafficPermissionType).String()
		}
		onMatch, err := bldrs_matchers.NewOnMatch().Configure(bldrs_matchers.RbacAction(rbacAction, name)).Build()
		if err != nil {
			return nil, err
		}
		fieldMatchers = append(fieldMatchers, &matcher_config.Matcher_MatcherList_FieldMatcher{
			Predicate: predicate,
			OnMatch:   onMatch,
		})
	}
	return fieldMatchers, nil
}

func buildMatchers(matches []common_api.Match, action rbac_config.RBAC_Action, origin common.Origin) (*matcher_config.Matcher_MatcherList_FieldMatcher, error) {
	return bldrs_matchers.NewFieldMatcher().
		Configure(bldrs_matchers.Matches(
			matches,
			bldrs_matchers.NewOnMatch().Configure(bldrs_matchers.RbacAction(action, kri.FromResourceMeta(origin.Resource, policies_api.MeshTrafficPermissionType).String())),
		)).
		Build()
}
