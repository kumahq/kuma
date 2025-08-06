package xds

import (
	"fmt"

	xds_config "github.com/cncf/xds/go/xds/core/v3"
	matcher_config "github.com/cncf/xds/go/xds/type/matcher/v3"
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	rbac_config "github.com/envoyproxy/go-control-plane/envoy/config/rbac/v3"
	http_rbac "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/rbac/v3"
	envoy_hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	network_rbac "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/rbac/v3"
	sslv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/matching/common_inputs/ssl/v3"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/kri"
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

	matcher := c.createMatcher()
	shadowMatcher := c.createShadowMatcher()

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

func (c *RBACConfigurer) createMatcher() *matcher_config.Matcher {
	var matchers []*matcher_config.Matcher_MatcherList_FieldMatcher
	for _, rule := range c.InboundRules {
		conf := rule.Conf.GetDefault().(policies_api.RuleConf)
		denyMatchers := buildMatchers(pointer.Deref(conf.Deny), rbac_config.RBAC_DENY, rule.Origin)
		if denyMatchers != nil {
			matchers = append(matchers, denyMatchers)
		}
		allowMatchers := buildMatchers(append(pointer.Deref(conf.Allow), pointer.Deref(conf.AllowWithShadowDeny)...), rbac_config.RBAC_ALLOW, rule.Origin)
		if allowMatchers != nil {
			matchers = append(matchers, allowMatchers)
		}
	}
	var matchersList *matcher_config.Matcher_MatcherList_
	if len(matchers) > 0 {
		matchersList = &matcher_config.Matcher_MatcherList_{
			MatcherList: &matcher_config.Matcher_MatcherList{
				Matchers: matchers,
			},
		}
	}

	return &matcher_config.Matcher{
		MatcherType: matchersList,
		OnNoMatch:   onMatch(rbac_config.RBAC_DENY, "default"),
	}
}

func (c *RBACConfigurer) createShadowMatcher() *matcher_config.Matcher {
	var matchers []*matcher_config.Matcher_MatcherList_FieldMatcher
	for _, rule := range c.InboundRules {
		conf := rule.Conf.GetDefault().(policies_api.RuleConf)
		shadowDenyMatchers := buildMatchers(pointer.Deref(conf.AllowWithShadowDeny), rbac_config.RBAC_DENY, rule.Origin)
		if shadowDenyMatchers != nil {
			matchers = append(matchers, shadowDenyMatchers)
		}
	}

	var matchersList *matcher_config.Matcher_MatcherList_
	if len(matchers) > 0 {
		matchersList = &matcher_config.Matcher_MatcherList_{
			MatcherList: &matcher_config.Matcher_MatcherList{
				Matchers: matchers,
			},
		}
	}

	return &matcher_config.Matcher{
		MatcherType: matchersList,
		OnNoMatch:   onMatch(rbac_config.RBAC_DENY, "default"),
	}
}

func onMatch(action rbac_config.RBAC_Action, name string) *matcher_config.Matcher_OnMatch {
	return &matcher_config.Matcher_OnMatch{OnMatch: &matcher_config.Matcher_OnMatch_Action{
		Action: &xds_config.TypedExtensionConfig{
			Name: "envoy.filters.rbac.action",
			TypedConfig: util_proto.MustMarshalAny(&rbac_config.Action{
				Name:   name,
				Action: action,
			}),
		},
	}}
}

func spiffeIdMatcher(spiffeId *common_api.SpiffeIdMatch) *matcher_config.Matcher_MatcherList_Predicate {
	var stringMatcher matcher_config.StringMatcher
	switch spiffeId.Type {
	case common_api.ExactMatchType:
		stringMatcher = matcher_config.StringMatcher{
			MatchPattern: &matcher_config.StringMatcher_Exact{
				Exact: spiffeId.Value,
			},
		}
	case common_api.PrefixMatchType:
		stringMatcher = matcher_config.StringMatcher{
			MatchPattern: &matcher_config.StringMatcher_Prefix{
				Prefix: spiffeId.Value,
			},
		}
	}

	return &matcher_config.Matcher_MatcherList_Predicate{
		MatchType: &matcher_config.Matcher_MatcherList_Predicate_SinglePredicate_{
			SinglePredicate: &matcher_config.Matcher_MatcherList_Predicate_SinglePredicate{
				Input: &xds_config.TypedExtensionConfig{
					Name:        "envoy.matching.inputs.uri_san",
					TypedConfig: util_proto.MustMarshalAny(&sslv3.UriSanInput{}),
				},
				Matcher: &matcher_config.Matcher_MatcherList_Predicate_SinglePredicate_ValueMatch{
					ValueMatch: &stringMatcher,
				},
			},
		},
	}
}

func buildPredicateFrom(matchers []*matcher_config.Matcher_MatcherList_Predicate) *matcher_config.Matcher_MatcherList_Predicate {
	var predicate matcher_config.Matcher_MatcherList_Predicate
	if len(matchers) == 1 {
		predicate = matcher_config.Matcher_MatcherList_Predicate{
			MatchType: matchers[0].MatchType,
		}
	} else if len(matchers) > 1 {
		predicate = matcher_config.Matcher_MatcherList_Predicate{
			MatchType: &matcher_config.Matcher_MatcherList_Predicate_OrMatcher{
				OrMatcher: &matcher_config.Matcher_MatcherList_Predicate_PredicateList{
					Predicate: matchers,
				},
			},
		}
	}
	return &predicate
}

func buildMatchers(matches []common_api.Match, action rbac_config.RBAC_Action, origin common.Origin) *matcher_config.Matcher_MatcherList_FieldMatcher {
	if len(matches) == 0 {
		return nil
	}
	var matchers []*matcher_config.Matcher_MatcherList_Predicate
	for _, match := range matches {
		if match.SpiffeId == nil {
			continue
		}
		matchers = append(matchers, spiffeIdMatcher(match.SpiffeId))
	}

	return &matcher_config.Matcher_MatcherList_FieldMatcher{
		Predicate: buildPredicateFrom(matchers),
		OnMatch:   onMatch(action, kri.FromResourceMeta(origin.Resource, policies_api.MeshTrafficPermissionType, "").String()),
	}
}
