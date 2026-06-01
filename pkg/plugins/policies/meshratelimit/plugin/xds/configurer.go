package xds

import (
	"slices"
	"strings"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	common_set_filter_state "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/common/set_filter_state/v3"
	envoy_extensions_filters_http_local_ratelimit_v3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/local_ratelimit/v3"
	envoy_hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	envoy_extensions_filters_network_local_ratelimit_v3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/local_ratelimit/v3"
	envoy_set_filter_state "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/set_filter_state/v3"
	envoy_matcher "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
	envoy_type_v3 "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	common_api "github.com/kumahq/kuma/v2/api/common/v1alpha1"
	core_rules "github.com/kumahq/kuma/v2/pkg/plugins/policies/core/rules"
	rules_inbound "github.com/kumahq/kuma/v2/pkg/plugins/policies/core/rules/inbound"
	"github.com/kumahq/kuma/v2/pkg/plugins/policies/core/rules/merge"
	"github.com/kumahq/kuma/v2/pkg/plugins/policies/core/rules/subsetutils"
	api "github.com/kumahq/kuma/v2/pkg/plugins/policies/meshratelimit/api/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/util/pointer"
	util_proto "github.com/kumahq/kuma/v2/pkg/util/proto"
	listeners_v3 "github.com/kumahq/kuma/v2/pkg/xds/envoy/listeners/v3"
	envoy_routes_v3 "github.com/kumahq/kuma/v2/pkg/xds/envoy/routes/v3"
)

const (
	httpLocalRateLimitFilterName      = "envoy.filters.http.local_ratelimit"
	tcpLocalRateLimitFilterName       = "envoy.extensions.filters.network.local_ratelimit.v3.LocalRateLimit"
	matchFilterStateNetworkFilterName = "envoy.filters.network.set_filter_state"
	matchSpiffeIDFilterStateKey       = "kuma.mesh_rate_limit.match.spiffe_id"
	matchSNIStateKey                  = "kuma.mesh_rate_limit.match.sni"
)

type ListenerConfigurer struct {
	Conf             api.Conf
	Rules            []*rules_inbound.Rule
	SkipCommonConfig bool
}

func (lc *ListenerConfigurer) ConfigureListener(listener *envoy_listener.Listener) error {
	if listener == nil {
		return nil
	}

	for _, filterChain := range listener.FilterChains {
		if !lc.SkipCommonConfig {
			if err := ConfigureFilterChain(lc.Conf, filterChain); err != nil {
				return err
			}
		}
		if err := ConfigureMatchedRoutesOnFilterChain(filterChain, lc.Conf, lc.Rules); err != nil {
			return err
		}
	}

	return nil
}

type Configurer struct {
	Element subsetutils.Element
	Rules   core_rules.Rules
	Conf    *api.Conf
}

func (c *Configurer) ConfigureFilterChain(filterChain *envoy_listener.FilterChain) error {
	conf := c.getConf(c.Element)
	if conf == nil || conf.Local == nil {
		return nil
	}

	if conf.Local.HTTP != nil {
		if err := configureHttpListener(filterChain, conf.Local.HTTP); err != nil {
			return err
		}
	}
	if conf.Local.TCP != nil {
		if err := configureTcpListener(filterChain, conf.Local.TCP); err != nil {
			return err
		}
	}
	return nil
}

func (c *Configurer) ConfigureRoute(route *envoy_route.RouteConfiguration) error {
	conf := c.getConf(c.Element)
	if route == nil || conf == nil || conf.Local == nil {
		return nil
	}

	rlConf := RateLimitConfigurationFromPolicy(conf.Local.HTTP)
	if rlConf == nil {
		return nil
	}

	rateLimit, err := envoy_routes_v3.NewRateLimitConfiguration(rlConf)
	if err != nil {
		return err
	}

	for _, vh := range route.VirtualHosts {
		for _, r := range vh.Routes {
			addRateLimitToRoute(r, rateLimit)
		}
	}

	return nil
}

func (c *Configurer) ConfigureGatewayRoute(route *envoy_route.RouteConfiguration) error {
	if route == nil {
		return nil
	}

	conf := c.getConf(c.Element)
	var defaultConf *envoy_routes_v3.RateLimitConfiguration
	if conf != nil && conf.Local != nil && conf.Local.HTTP != nil {
		defaultConf = RateLimitConfigurationFromPolicy(conf.Local.HTTP)
	}

	var err error
	var defaultRateLimit *anypb.Any
	if defaultConf != nil {
		defaultRateLimit, err = envoy_routes_v3.NewRateLimitConfiguration(defaultConf)
	}
	if err != nil {
		return err
	}

	for _, vh := range route.VirtualHosts {
		for _, r := range vh.Routes {
			conf := c.getConf(c.Element.WithKeyValue(core_rules.RuleMatchesHashTag, r.Name))
			var routeConf *envoy_routes_v3.RateLimitConfiguration
			var rateLimit *anypb.Any
			if conf != nil && conf.Local != nil && conf.Local.HTTP != nil {
				routeConf = RateLimitConfigurationFromPolicy(conf.Local.HTTP)
			}
			if routeConf != nil {
				rateLimit, err = envoy_routes_v3.NewRateLimitConfiguration(routeConf)
			}
			if err != nil {
				return err
			}
			if defaultConf == nil && routeConf == nil {
				continue
			}
			if routeConf == nil {
				rateLimit = defaultRateLimit
			}
			addRateLimitToRoute(r, rateLimit)
		}
	}

	return nil
}

func ConfigureFilterChain(conf api.Conf, filterChain *envoy_listener.FilterChain) error {
	if conf.Local == nil {
		return nil
	}
	if conf.Local.HTTP != nil {
		if err := configureHttpListener(filterChain, conf.Local.HTTP); err != nil {
			return err
		}
	}
	if conf.Local.TCP != nil {
		if err := configureTcpListener(filterChain, conf.Local.TCP); err != nil {
			return err
		}
	}
	return nil
}

func ConfigureRoute(conf api.Conf, route *envoy_route.RouteConfiguration) error {
	if route == nil || conf.Local == nil {
		return nil
	}

	rlConf := RateLimitConfigurationFromPolicy(conf.Local.HTTP)
	if rlConf == nil {
		return nil
	}

	rateLimit, err := envoy_routes_v3.NewRateLimitConfiguration(rlConf)
	if err != nil {
		return err
	}

	for _, vh := range route.VirtualHosts {
		for _, r := range vh.Routes {
			addRateLimitToRoute(r, rateLimit)
		}
	}

	return nil
}

func ConfigureMatchedRoutesOnFilterChain(filterChain *envoy_listener.FilterChain, baseConf api.Conf, inboundRules []*rules_inbound.Rule) error {
	if filterChain == nil {
		return nil
	}

	matchedRules := matchedRouteRateLimitRules(inboundRules)
	if len(matchedRules) == 0 {
		return nil
	}

	effectiveRules, err := effectiveMatchedRouteRules(baseConf, matchedRules)
	if err != nil {
		return err
	}
	if len(effectiveRules) == 0 {
		return nil
	}

	if err := ensureMatchFilterStateFilter(filterChain, effectiveRules); err != nil {
		return err
	}

	if err := listeners_v3.UpdateHTTPConnectionManager(filterChain, func(hcm *envoy_hcm.HttpConnectionManager) error {
		if hasMatchedRateLimit(effectiveRules) {
			httpFilter, err := newHTTPLocalRateLimitFilter()
			if err != nil {
				return err
			}
			upsertHTTPFilter(hcm, httpFilter)
		}
		return ConfigureMatchedRoutes(hcm.GetRouteConfig(), effectiveRules)
	}); err != nil && !errors.Is(err, &listeners_v3.UnexpectedFilterConfigTypeError{}) {
		return err
	}

	return nil
}

func ConfigureMatchedRoutes(routeConfiguration *envoy_route.RouteConfiguration, effectiveRules []effectiveMatchedRouteRule) error {
	if routeConfiguration == nil || len(effectiveRules) == 0 {
		return nil
	}

	for _, virtualHost := range routeConfiguration.VirtualHosts {
		originalRoutes := append([]*envoy_route.Route(nil), virtualHost.Routes...)
		virtualHost.Routes = make([]*envoy_route.Route, 0, len(originalRoutes)*(len(effectiveRules)+1))
		for _, route := range originalRoutes {
			virtualHost.Routes = append(virtualHost.Routes, duplicateMatchedRoutes(route, effectiveRules)...)
			virtualHost.Routes = append(virtualHost.Routes, route)
		}
	}

	return nil
}

func RateLimitConfigurationFromPolicy(rl *api.LocalHTTP) *envoy_routes_v3.RateLimitConfiguration {
	if pointer.Deref(rl.Disabled) || rl.RequestRate == nil {
		return nil
	}

	onRateLimit := &envoy_routes_v3.OnRateLimit{}
	if rl.OnRateLimit != nil {
		for _, h := range pointer.Deref(pointer.Deref(rl.OnRateLimit.Headers).Add) {
			onRateLimit.Headers = append(onRateLimit.Headers, &envoy_routes_v3.Headers{
				Key:    string(h.Name),
				Value:  string(h.Value),
				Append: true,
			})
		}
		for _, header := range pointer.Deref(pointer.Deref(rl.OnRateLimit.Headers).Set) {
			for val := range strings.SplitSeq(string(header.Value), ",") {
				onRateLimit.Headers = append(onRateLimit.Headers, &envoy_routes_v3.Headers{
					Key:    string(header.Name),
					Value:  val,
					Append: false,
				})
			}
		}
		onRateLimit.Status = pointer.Deref(rl.OnRateLimit.Status)
	}

	return &envoy_routes_v3.RateLimitConfiguration{
		Interval:    rl.RequestRate.Interval.Duration,
		Requests:    rl.RequestRate.Num,
		OnRateLimit: onRateLimit,
	}
}

func configureHttpListener(filterChain *envoy_listener.FilterChain, conf *api.LocalHTTP) error {
	rlConf := RateLimitConfigurationFromPolicy(conf)
	if rlConf == nil {
		return nil
	}

	rateLimit, err := envoy_routes_v3.NewRateLimitConfiguration(rlConf)
	if err != nil {
		return err
	}

	httpFilter, err := newHTTPLocalRateLimitFilter()
	if err != nil {
		return err
	}

	if err := listeners_v3.UpdateHTTPConnectionManager(filterChain, func(hcm *envoy_hcm.HttpConnectionManager) error {
		if hcm.GetRouteConfig() == nil {
			return nil
		}

		for _, vh := range hcm.GetRouteConfig().GetVirtualHosts() {
			for _, route := range vh.GetRoutes() {
				addRateLimitToRoute(route, rateLimit)
			}
		}

		upsertHTTPFilter(hcm, httpFilter)
		return nil
	}); err != nil && !errors.Is(err, &listeners_v3.UnexpectedFilterConfigTypeError{}) {
		return err
	}

	return nil
}

func newHTTPLocalRateLimitFilter() (*envoy_hcm.HttpFilter, error) {
	typedConfig, err := util_proto.MarshalAnyDeterministic(&envoy_extensions_filters_http_local_ratelimit_v3.LocalRateLimit{
		StatPrefix: "rate_limit",
	})
	if err != nil {
		return nil, err
	}

	return &envoy_hcm.HttpFilter{
		Name: httpLocalRateLimitFilterName,
		ConfigType: &envoy_hcm.HttpFilter_TypedConfig{
			TypedConfig: typedConfig,
		},
	}, nil
}

func upsertHTTPFilter(hcm *envoy_hcm.HttpConnectionManager, filter *envoy_hcm.HttpFilter) {
	for idx, existing := range hcm.GetHttpFilters() {
		if existing.GetName() == filter.GetName() {
			hcm.HttpFilters[idx] = filter
			return
		}
	}
	hcm.HttpFilters = append([]*envoy_hcm.HttpFilter{filter}, hcm.HttpFilters...)
}

func configureTcpListener(filterChain *envoy_listener.FilterChain, conf *api.LocalTCP) error {
	if pointer.Deref(conf.Disabled) || conf.ConnectionRate == nil {
		return nil
	}

	config := &envoy_extensions_filters_network_local_ratelimit_v3.LocalRateLimit{
		StatPrefix: "tcp_rate_limit",
		TokenBucket: &envoy_type_v3.TokenBucket{
			MaxTokens:     conf.ConnectionRate.Num,
			TokensPerFill: util_proto.UInt32(conf.ConnectionRate.Num),
			FillInterval:  util_proto.Duration(conf.ConnectionRate.Interval.Duration),
		},
	}
	typedConfig, err := util_proto.MarshalAnyDeterministic(config)
	if err != nil {
		return err
	}

	filter := &envoy_listener.Filter{
		Name: tcpLocalRateLimitFilterName,
		ConfigType: &envoy_listener.Filter_TypedConfig{
			TypedConfig: typedConfig,
		},
	}
	for idx, existing := range filterChain.GetFilters() {
		if existing.GetName() == filter.GetName() {
			filterChain.Filters[idx] = filter
			return nil
		}
	}
	filterChain.Filters = append([]*envoy_listener.Filter{filter}, filterChain.Filters...)
	return nil
}

func addRateLimitToRoute(route *envoy_route.Route, rateLimit *anypb.Any) {
	if route.TypedPerFilterConfig == nil {
		route.TypedPerFilterConfig = map[string]*anypb.Any{}
	}
	if route.TypedPerFilterConfig[httpLocalRateLimitFilterName] != nil {
		return
	}
	route.TypedPerFilterConfig[httpLocalRateLimitFilterName] = rateLimit
}

func setRateLimitOnRoute(route *envoy_route.Route, rateLimit *anypb.Any) {
	if route.TypedPerFilterConfig == nil {
		if rateLimit == nil {
			return
		}
		route.TypedPerFilterConfig = map[string]*anypb.Any{}
	}
	if rateLimit == nil {
		delete(route.TypedPerFilterConfig, httpLocalRateLimitFilterName)
		return
	}
	route.TypedPerFilterConfig[httpLocalRateLimitFilterName] = rateLimit
}

func (c *Configurer) getConf(element subsetutils.Element) *api.Conf {
	if c.Conf != nil {
		return c.Conf
	}
	if c.Rules == nil {
		return &api.Conf{}
	}
	return core_rules.ComputeConf[api.Conf](c.Rules, element)
}

type effectiveMatchedRouteRule struct {
	Match     *common_api.Match
	Conf      api.Conf
	RateLimit *anypb.Any
}

func matchedRouteRateLimitRules(inboundRules []*rules_inbound.Rule) []*rules_inbound.Rule {
	var matched []*rules_inbound.Rule
	for _, rule := range inboundRules {
		if !hasSpiffeIDMatch(rule.Match) {
			continue
		}
		conf, ok := rule.Conf.(api.Conf)
		if !ok || conf.Local == nil || conf.Local.HTTP == nil {
			continue
		}
		matched = append(matched, rule)
	}
	slices.SortStableFunc(matched, compareMatchedRouteRuleSpecificityDesc)
	return matched
}

func hasSpiffeIDMatch(match *common_api.Match) bool {
	return match != nil && match.SpiffeID != nil
}

func effectiveMatchedRouteRules(baseConf api.Conf, matchedRules []*rules_inbound.Rule) ([]effectiveMatchedRouteRule, error) {
	var effective []effectiveMatchedRouteRule
	for _, rule := range matchedRules {
		if containsEffectiveMatchedRoute(effective, rule.Match) {
			continue
		}

		conf, ok, err := mergeSubsumingMatchedRouteConfs(baseConf, matchedRules, rule.Match)
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}

		rateLimit, err := routeRateLimit(conf)
		if err != nil {
			return nil, err
		}

		effective = append(effective, effectiveMatchedRouteRule{
			Match:     rule.Match,
			Conf:      conf,
			RateLimit: rateLimit,
		})
	}
	return effective, nil
}

func containsEffectiveMatchedRoute(rules []effectiveMatchedRouteRule, match *common_api.Match) bool {
	for _, rule := range rules {
		if sameMatchedRoute(rule.Match, match) {
			return true
		}
	}
	return false
}

func sameMatchedRoute(a, b *common_api.Match) bool {
	if a == nil || b == nil {
		return a == b
	}

	return sameSpiffeIDMatch(a.SpiffeID, b.SpiffeID) &&
		sameSNIMatch(a.SNI, b.SNI)
}

func mergeSubsumingMatchedRouteConfs(baseConf api.Conf, matchedRules []*rules_inbound.Rule, target *common_api.Match) (api.Conf, bool, error) {
	var applicable []*rules_inbound.Rule
	for _, rule := range matchedRules {
		if matchedRouteSubsumes(rule.Match, target) {
			applicable = append(applicable, rule)
		}
	}
	if len(applicable) == 0 && (baseConf.Local == nil || baseConf.Local.HTTP == nil) {
		return api.Conf{}, false, nil
	}

	slices.SortStableFunc(applicable, compareMatchedRouteRuleSpecificityAsc)

	confs := make([]any, 0, len(applicable)+1)
	if baseConf.Local != nil && baseConf.Local.HTTP != nil {
		confs = append(confs, baseConf)
	}
	for _, rule := range applicable {
		conf, ok := rule.Conf.(api.Conf)
		if !ok {
			continue
		}
		confs = append(confs, conf)
	}
	if len(confs) == 0 {
		return api.Conf{}, false, nil
	}

	merged, err := merge.Confs(confs)
	if err != nil {
		return api.Conf{}, false, err
	}
	if len(merged) == 0 {
		return api.Conf{}, false, nil
	}

	conf, ok := merged[0].(api.Conf)
	if !ok {
		return api.Conf{}, false, errors.Errorf("unexpected merged matched route conf type: %T", merged[0])
	}
	return conf, true, nil
}

func routeRateLimit(conf api.Conf) (*anypb.Any, error) {
	if conf.Local == nil || conf.Local.HTTP == nil {
		return nil, nil
	}

	rlConf := RateLimitConfigurationFromPolicy(conf.Local.HTTP)
	if rlConf == nil {
		return nil, nil
	}

	return envoy_routes_v3.NewRateLimitConfiguration(rlConf)
}

func matchedRouteSubsumes(candidate, target *common_api.Match) bool {
	if candidate == nil || target == nil {
		return false
	}

	return spiffeIDMatchSubsumes(candidate.SpiffeID, target.SpiffeID) &&
		sniMatchSubsumes(candidate.SNI, target.SNI)
}

func compareMatchedRouteRuleSpecificityDesc(a, b *rules_inbound.Rule) int {
	return compareMatchedRouteSpecificity(a.Match, b.Match, true)
}

func compareMatchedRouteRuleSpecificityAsc(a, b *rules_inbound.Rule) int {
	return compareMatchedRouteSpecificity(a.Match, b.Match, false)
}

func compareMatchedRouteSpecificity(a, b *common_api.Match, descending bool) int {
	if c := rules_inbound.CompareByMatch(a, b); c != 0 {
		if descending {
			return c
		}
		return -c
	}

	aLen := matchedRoutePrefixLen(a)
	bLen := matchedRoutePrefixLen(b)
	if aLen == 0 || bLen == 0 {
		return 0
	}

	switch {
	case descending && aLen > bLen:
		return -1
	case descending && aLen < bLen:
		return 1
	case !descending && aLen < bLen:
		return -1
	case !descending && aLen > bLen:
		return 1
	default:
		return 0
	}
}

func matchedRoutePrefixLen(match *common_api.Match) int {
	if match == nil || match.SpiffeID == nil || match.SpiffeID.Type != common_api.PrefixMatchType {
		return 0
	}
	return len(match.SpiffeID.Value)
}

func sameSpiffeIDMatch(a, b *common_api.SpiffeIDMatch) bool {
	switch {
	case a == nil || b == nil:
		return a == b
	default:
		return a.Type == b.Type && a.Value == b.Value
	}
}

func sameSNIMatch(a, b *common_api.SNIMatch) bool {
	switch {
	case a == nil || b == nil:
		return a == b
	default:
		return a.Type == b.Type && a.Value == b.Value
	}
}

func spiffeIDMatchSubsumes(candidate, target *common_api.SpiffeIDMatch) bool {
	if candidate == nil || target == nil {
		return false
	}

	switch candidate.Type {
	case common_api.ExactMatchType:
		return target.Type == common_api.ExactMatchType && candidate.Value == target.Value
	case common_api.PrefixMatchType:
		return strings.HasPrefix(target.Value, candidate.Value)
	default:
		return false
	}
}

func sniMatchSubsumes(candidate, target *common_api.SNIMatch) bool {
	if candidate == nil {
		return true
	}
	if target == nil {
		return false
	}

	return candidate.Type == target.Type && candidate.Value == target.Value
}

func duplicateMatchedRoutes(route *envoy_route.Route, rules []effectiveMatchedRouteRule) []*envoy_route.Route {
	var duplicated []*envoy_route.Route
	for _, rule := range rules {
		clone := proto.Clone(route).(*envoy_route.Route)
		if clone.Match == nil || !matchedRouteSupportsRateLimit(clone.GetRoute()) {
			continue
		}
		clone.Match.FilterState = append(clone.Match.FilterState, routeFilterStateMatchers(rule.Match)...)
		setRateLimitOnRoute(clone, rule.RateLimit)
		duplicated = append(duplicated, clone)
	}
	return duplicated
}

func matchedRouteSupportsRateLimit(routeAction *envoy_route.RouteAction) bool {
	if routeAction == nil {
		return false
	}

	switch specifier := routeAction.ClusterSpecifier.(type) {
	case *envoy_route.RouteAction_Cluster:
		return specifier.Cluster != ""
	case *envoy_route.RouteAction_ClusterHeader:
		return specifier.ClusterHeader != ""
	case *envoy_route.RouteAction_WeightedClusters:
		if specifier.WeightedClusters == nil || len(specifier.WeightedClusters.Clusters) == 0 {
			return false
		}
		for _, cluster := range specifier.WeightedClusters.Clusters {
			if cluster.GetName() == "" && cluster.GetClusterHeader() == "" {
				return false
			}
		}
		return true
	default:
		return false
	}
}

func routeFilterStateMatchers(match *common_api.Match) []*envoy_matcher.FilterStateMatcher {
	if match == nil {
		return nil
	}

	var matchers []*envoy_matcher.FilterStateMatcher
	if match.SpiffeID != nil {
		matchers = append(matchers, &envoy_matcher.FilterStateMatcher{
			Key: matchSpiffeIDFilterStateKey,
			Matcher: &envoy_matcher.FilterStateMatcher_StringMatch{
				StringMatch: spiffeIDMatcher(match.SpiffeID),
			},
		})
	}
	if match.SNI != nil {
		matchers = append(matchers, &envoy_matcher.FilterStateMatcher{
			Key: matchSNIStateKey,
			Matcher: &envoy_matcher.FilterStateMatcher_StringMatch{
				StringMatch: &envoy_matcher.StringMatcher{
					MatchPattern: &envoy_matcher.StringMatcher_Exact{Exact: match.SNI.Value},
				},
			},
		})
	}
	return matchers
}

func spiffeIDMatcher(match *common_api.SpiffeIDMatch) *envoy_matcher.StringMatcher {
	switch match.Type {
	case common_api.PrefixMatchType:
		return &envoy_matcher.StringMatcher{
			MatchPattern: &envoy_matcher.StringMatcher_Prefix{Prefix: match.Value},
		}
	default:
		return &envoy_matcher.StringMatcher{
			MatchPattern: &envoy_matcher.StringMatcher_Exact{Exact: match.Value},
		}
	}
}

func hasMatchedRateLimit(effectiveRules []effectiveMatchedRouteRule) bool {
	for _, rule := range effectiveRules {
		if rule.RateLimit != nil {
			return true
		}
	}
	return false
}

func ensureMatchFilterStateFilter(filterChain *envoy_listener.FilterChain, effectiveRules []effectiveMatchedRouteRule) error {
	if filterChain == nil {
		return nil
	}

	var requiredValues []*common_set_filter_state.FilterStateValue
	if hasAnySpiffeIDMatch(effectiveRules) {
		requiredValues = append(requiredValues, newFilterStateValue(matchSpiffeIDFilterStateKey, "%DOWNSTREAM_PEER_URI_SAN%"))
	}
	if hasAnySNIMatch(effectiveRules) {
		requiredValues = append(requiredValues, newFilterStateValue(matchSNIStateKey, "%REQUESTED_SERVER_NAME%"))
	}
	if len(requiredValues) == 0 {
		return nil
	}

	for idx, filter := range filterChain.Filters {
		if filter.Name != matchFilterStateNetworkFilterName {
			continue
		}

		config := &envoy_set_filter_state.Config{}
		typedConfig, ok := filter.ConfigType.(*envoy_listener.Filter_TypedConfig)
		if !ok {
			return errors.Errorf("unexpected %s config type: %T", filter.Name, filter.ConfigType)
		}
		if err := util_proto.UnmarshalAnyTo(typedConfig.TypedConfig, config); err != nil {
			return err
		}
		if !appendMissingFilterStateValues(config, requiredValues) {
			return nil
		}

		marshaled, err := util_proto.MarshalAnyDeterministic(config)
		if err != nil {
			return err
		}
		filterChain.Filters[idx] = &envoy_listener.Filter{
			Name: matchFilterStateNetworkFilterName,
			ConfigType: &envoy_listener.Filter_TypedConfig{
				TypedConfig: marshaled,
			},
		}
		return nil
	}

	config := &envoy_set_filter_state.Config{
		OnDownstreamTlsHandshake: requiredValues,
	}
	typedConfig, err := util_proto.MarshalAnyDeterministic(config)
	if err != nil {
		return err
	}

	filterChain.Filters = append([]*envoy_listener.Filter{{
		Name: matchFilterStateNetworkFilterName,
		ConfigType: &envoy_listener.Filter_TypedConfig{
			TypedConfig: typedConfig,
		},
	}}, filterChain.Filters...)
	return nil
}

func hasAnySpiffeIDMatch(rules []effectiveMatchedRouteRule) bool {
	for _, rule := range rules {
		if rule.Match != nil && rule.Match.SpiffeID != nil {
			return true
		}
	}
	return false
}

func hasAnySNIMatch(rules []effectiveMatchedRouteRule) bool {
	for _, rule := range rules {
		if rule.Match != nil && rule.Match.SNI != nil {
			return true
		}
	}
	return false
}

func appendMissingFilterStateValues(config *envoy_set_filter_state.Config, requiredValues []*common_set_filter_state.FilterStateValue) bool {
	existing := map[string]struct{}{}
	for _, value := range config.OnDownstreamTlsHandshake {
		existing[value.GetObjectKey()] = struct{}{}
	}

	changed := false
	for _, value := range requiredValues {
		key := value.GetObjectKey()
		if _, ok := existing[key]; ok {
			continue
		}
		config.OnDownstreamTlsHandshake = append(config.OnDownstreamTlsHandshake, value)
		existing[key] = struct{}{}
		changed = true
	}

	return changed
}

func newFilterStateValue(key, format string) *common_set_filter_state.FilterStateValue {
	return &common_set_filter_state.FilterStateValue{
		Key:        &common_set_filter_state.FilterStateValue_ObjectKey{ObjectKey: key},
		FactoryKey: "envoy.string",
		Value: &common_set_filter_state.FilterStateValue_FormatString{
			FormatString: &envoy_core.SubstitutionFormatString{
				Format: &envoy_core.SubstitutionFormatString_TextFormatSource{
					TextFormatSource: &envoy_core.DataSource{
						Specifier: &envoy_core.DataSource_InlineString{
							InlineString: format,
						},
					},
				},
			},
		},
		ReadOnly:    true,
		SkipIfEmpty: true,
	}
}
