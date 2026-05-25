package xds

import (
	"slices"
	"strings"
	"time"

	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	common_set_filter_state "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/common/set_filter_state/v3"
	envoy_hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	envoy_set_filter_state "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/set_filter_state/v3"
	envoy_tcp "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/tcp_proxy/v3"
	envoy_upstream_http "github.com/envoyproxy/go-control-plane/envoy/extensions/upstreams/http/v3"
	envoy_matcher "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	common_api "github.com/kumahq/kuma/v2/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	core_meta "github.com/kumahq/kuma/v2/pkg/core/metadata"
	policies_defaults "github.com/kumahq/kuma/v2/pkg/plugins/policies/core/defaults"
	"github.com/kumahq/kuma/v2/pkg/plugins/policies/core/rules"
	rules_inbound "github.com/kumahq/kuma/v2/pkg/plugins/policies/core/rules/inbound"
	"github.com/kumahq/kuma/v2/pkg/plugins/policies/core/rules/merge"
	"github.com/kumahq/kuma/v2/pkg/plugins/policies/core/rules/subsetutils"
	api "github.com/kumahq/kuma/v2/pkg/plugins/policies/meshtimeout/api/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/util/pointer"
	util_proto "github.com/kumahq/kuma/v2/pkg/util/proto"
	clusters_v3 "github.com/kumahq/kuma/v2/pkg/xds/envoy/clusters/v3"
	listeners_v3 "github.com/kumahq/kuma/v2/pkg/xds/envoy/listeners/v3"
)

const (
	matchFilterStateNetworkFilterName = "envoy.filters.network.set_filter_state"
	matchSpiffeIDFilterStateKey       = "kuma.mesh_timeout.match.spiffe_id"
	matchSNIStateKey                  = "kuma.mesh_timeout.match.sni"
)

// DeprecatedListenerConfigurer should be only used for configuring old MeshService outbounds.
// It should be removed after we stop using kuma.io/service tag, and move fully to new MeshService
// Deprecated
type DeprecatedListenerConfigurer struct {
	Rules    rules.Rules
	Protocol core_meta.Protocol
	Element  subsetutils.Element
}

func (c *DeprecatedListenerConfigurer) ConfigureListener(listener *envoy_listener.Listener) error {
	if listener == nil {
		return nil
	}

	httpTimeouts := func(hcm *envoy_hcm.HttpConnectionManager) error {
		c.configureRequestTimeout(hcm.GetRouteConfig())
		c.configureRequestHeadersTimeout(hcm)
		// old Timeout policy configures idleTimeout on listener while MeshTimeout sets this in cluster
		if hcm.CommonHttpProtocolOptions == nil {
			hcm.CommonHttpProtocolOptions = &envoy_core.HttpProtocolOptions{}
		}

		hcm.CommonHttpProtocolOptions.IdleTimeout = util_proto.Duration(0)
		return nil
	}
	tcpTimeouts := func(proxy *envoy_tcp.TcpProxy) error {
		if conf := c.getConf(c.Element); conf != nil {
			proxy.IdleTimeout = toProtoDurationOrDefault(conf.IdleTimeout, policies_defaults.DefaultIdleTimeout)
		}
		return nil
	}
	for _, filterChain := range listener.FilterChains {
		switch c.Protocol {
		case core_meta.ProtocolHTTP, core_meta.ProtocolHTTP2, core_meta.ProtocolGRPC:
			if err := listeners_v3.UpdateHTTPConnectionManager(filterChain, httpTimeouts); err != nil && !errors.Is(err, &listeners_v3.UnexpectedFilterConfigTypeError{}) {
				return err
			}
		case core_meta.ProtocolUnknown, core_meta.ProtocolTCP, core_meta.ProtocolKafka:
			if err := listeners_v3.UpdateTCPProxy(filterChain, tcpTimeouts); err != nil && !errors.Is(err, &listeners_v3.UnexpectedFilterConfigTypeError{}) {
				return err
			}
		}
	}

	return nil
}

func (c *DeprecatedListenerConfigurer) configureRequestTimeout(routeConfiguration *envoy_route.RouteConfiguration) {
	if routeConfiguration != nil {
		for _, vh := range routeConfiguration.VirtualHosts {
			for _, route := range vh.Routes {
				conf := c.getConf(c.Element.WithKeyValue(rules.RuleMatchesHashTag, route.Name))
				if conf == nil {
					conf = c.getConf(c.Element)
				}
				if conf == nil {
					continue
				}
				ConfigureRouteAction(
					route.GetRoute(),
					pointer.Deref(conf.Http).RequestTimeout,
					pointer.Deref(conf.Http).StreamIdleTimeout,
				)
			}
		}
	}
}

func (c *DeprecatedListenerConfigurer) configureRequestHeadersTimeout(hcm *envoy_hcm.HttpConnectionManager) {
	if conf := c.getConf(c.Element); conf != nil {
		hcm.RequestHeadersTimeout = toProtoDurationOrDefault(
			pointer.Deref(conf.Http).RequestHeadersTimeout,
			policies_defaults.DefaultRequestHeadersTimeout,
		)
	}
}

func (c *DeprecatedListenerConfigurer) getConf(element subsetutils.Element) *api.Conf {
	if c.Rules == nil {
		return &api.Conf{}
	}
	return rules.ComputeConf[api.Conf](c.Rules, element)
}

type ClusterConfigurer struct {
	ConnectionTimeout         *kube_meta.Duration
	IdleTimeout               *kube_meta.Duration
	HTTPMaxStreamDuration     *kube_meta.Duration
	HTTPMaxConnectionDuration *kube_meta.Duration
	Protocol                  core_meta.Protocol
}

func ClusterConfigurerFromConf(conf api.Conf, protocol core_meta.Protocol) ClusterConfigurer {
	return ClusterConfigurer{
		ConnectionTimeout:         conf.ConnectionTimeout,
		IdleTimeout:               conf.IdleTimeout,
		HTTPMaxStreamDuration:     pointer.Deref(conf.Http).MaxStreamDuration,
		HTTPMaxConnectionDuration: pointer.Deref(conf.Http).MaxConnectionDuration,
		Protocol:                  protocol,
	}
}

func (c *ClusterConfigurer) Configure(cluster *envoy_cluster.Cluster) error {
	cluster.ConnectTimeout = toProtoDurationOrDefault(c.ConnectionTimeout, policies_defaults.DefaultConnectTimeout)
	switch c.Protocol {
	case core_meta.ProtocolHTTP, core_meta.ProtocolHTTP2:
		err := clusters_v3.UpdateCommonHttpProtocolOptions(cluster, func(options *envoy_upstream_http.HttpProtocolOptions) {
			if options.CommonHttpProtocolOptions == nil {
				options.CommonHttpProtocolOptions = &envoy_core.HttpProtocolOptions{}
			}
			commonHttp := options.CommonHttpProtocolOptions
			commonHttp.IdleTimeout = toProtoDurationOrDefault(c.IdleTimeout, policies_defaults.DefaultIdleTimeout)
			commonHttp.MaxStreamDuration = toProtoDurationOrDefault(c.HTTPMaxStreamDuration, policies_defaults.DefaultMaxStreamDuration)
			commonHttp.MaxConnectionDuration = toProtoDurationOrDefault(c.HTTPMaxConnectionDuration, policies_defaults.DefaultMaxConnectionDuration)
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func ConfigureRouteAction(
	routeAction *envoy_route.RouteAction,
	httpRequestTimeout *kube_meta.Duration,
	httpStreamIdleTimeout *kube_meta.Duration,
) {
	if routeAction == nil {
		return
	}
	routeAction.Timeout = toProtoDurationOrDefault(httpRequestTimeout, policies_defaults.DefaultRequestTimeout)
	if httpStreamIdleTimeout != nil {
		routeAction.IdleTimeout = toProtoDurationOrDefault(httpStreamIdleTimeout, policies_defaults.DefaultStreamIdleTimeout)
	} else if routeAction.IdleTimeout == nil {
		routeAction.IdleTimeout = util_proto.Duration(policies_defaults.DefaultStreamIdleTimeout)
	}
}

func ConfigureGatewayListener(conf api.Conf, protocol mesh_proto.MeshGateway_Listener_Protocol, listener *envoy_listener.Listener) error {
	if listener == nil {
		return nil
	}

	httpTimeouts := func(hcm *envoy_hcm.HttpConnectionManager) error {
		if hcm.CommonHttpProtocolOptions == nil {
			hcm.CommonHttpProtocolOptions = &envoy_core.HttpProtocolOptions{}
		}
		hcm.CommonHttpProtocolOptions.IdleTimeout = toProtoDurationOrDefault(
			conf.IdleTimeout,
			policies_defaults.DefaultGatewayIdleTimeout,
		)
		hcm.RequestHeadersTimeout = toProtoDurationOrDefault(
			pointer.Deref(conf.Http).RequestHeadersTimeout,
			policies_defaults.DefaultGatewayRequestHeadersTimeout,
		)
		hcm.StreamIdleTimeout = toProtoDurationOrDefault(
			pointer.Deref(conf.Http).StreamIdleTimeout,
			policies_defaults.DefaultGatewayStreamIdleTimeout,
		)
		if httpConf := pointer.Deref(conf.Http); httpConf.RequestTimeout != nil {
			hcm.RequestTimeout = util_proto.Duration(httpConf.RequestTimeout.Duration)
		}
		return nil
	}
	tcpTimeouts := func(proxy *envoy_tcp.TcpProxy) error {
		proxy.IdleTimeout = toProtoDurationOrDefault(conf.IdleTimeout, policies_defaults.DefaultGatewayIdleTimeout)
		return nil
	}
	for _, filterChain := range listener.FilterChains {
		switch protocol {
		case mesh_proto.MeshGateway_Listener_HTTP, mesh_proto.MeshGateway_Listener_HTTPS:
			if err := listeners_v3.UpdateHTTPConnectionManager(filterChain, httpTimeouts); err != nil && !errors.Is(err, &listeners_v3.UnexpectedFilterConfigTypeError{}) {
				return err
			}
		case mesh_proto.MeshGateway_Listener_TCP, mesh_proto.MeshGateway_Listener_TLS:
			if err := listeners_v3.UpdateTCPProxy(filterChain, tcpTimeouts); err != nil && !errors.Is(err, &listeners_v3.UnexpectedFilterConfigTypeError{}) {
				return err
			}
		}
	}

	return nil
}

func ConfigureFilterChain(conf api.Conf, filterChain *envoy_listener.FilterChain) error {
	if filterChain == nil {
		return nil
	}

	httpTimeouts := func(hcm *envoy_hcm.HttpConnectionManager) error {
		configureRouteConfigurationTimeouts(hcm.GetRouteConfig(), conf)
		configureRequestHeadersTimeout(conf, hcm)
		// old Timeout policy configures idleTimeout on listener while MeshTimeout sets this in cluster
		if hcm.CommonHttpProtocolOptions == nil {
			hcm.CommonHttpProtocolOptions = &envoy_core.HttpProtocolOptions{}
		}
		hcm.CommonHttpProtocolOptions.IdleTimeout = util_proto.Duration(0)
		return nil
	}
	tcpTimeouts := func(proxy *envoy_tcp.TcpProxy) error {
		proxy.IdleTimeout = toProtoDurationOrDefault(conf.IdleTimeout, policies_defaults.DefaultIdleTimeout)
		return nil
	}

	if err := listeners_v3.UpdateHTTPConnectionManager(filterChain, httpTimeouts); err != nil && !errors.Is(err, &listeners_v3.UnexpectedFilterConfigTypeError{}) {
		return err
	}
	if err := listeners_v3.UpdateTCPProxy(filterChain, tcpTimeouts); err != nil && !errors.Is(err, &listeners_v3.UnexpectedFilterConfigTypeError{}) {
		return err
	}

	return nil
}

func ConfigureMatchedRoutesOnFilterChain(filterChain *envoy_listener.FilterChain, inboundRules []*rules_inbound.Rule) error {
	if filterChain == nil {
		return nil
	}

	matchedRules := matchedRouteTimeoutRules(inboundRules)
	if len(matchedRules) == 0 {
		return nil
	}

	if err := ensureMatchFilterStateFilter(filterChain, matchedRules); err != nil {
		return err
	}

	if err := listeners_v3.UpdateHTTPConnectionManager(filterChain, func(hcm *envoy_hcm.HttpConnectionManager) error {
		return ConfigureMatchedRoutes(hcm.GetRouteConfig(), matchedRules)
	}); err != nil && !errors.Is(err, &listeners_v3.UnexpectedFilterConfigTypeError{}) {
		return err
	}

	return nil
}

func EnsureMatchFilterState(listener *envoy_listener.Listener, inboundRules []*rules_inbound.Rule) error {
	if listener == nil {
		return nil
	}

	matchedRules := matchedRouteTimeoutRules(inboundRules)
	if len(matchedRules) == 0 {
		return nil
	}

	for _, filterChain := range listener.FilterChains {
		if err := ensureMatchFilterStateFilter(filterChain, matchedRules); err != nil {
			return err
		}
	}

	return nil
}

func ConfigureMatchedRoutes(routeConfiguration *envoy_route.RouteConfiguration, inboundRules []*rules_inbound.Rule) error {
	if routeConfiguration == nil {
		return nil
	}

	matchedRules := matchedRouteTimeoutRules(inboundRules)
	if len(matchedRules) == 0 {
		return nil
	}
	effectiveRules, err := effectiveMatchedRouteRules(matchedRules)
	if err != nil {
		return err
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

func toProtoDurationOrDefault(d *kube_meta.Duration, defaultDuration time.Duration) *durationpb.Duration {
	if d == nil {
		return util_proto.Duration(defaultDuration)
	}
	return util_proto.Duration(d.Duration)
}

type ListenerConfigurer struct {
	Conf             api.Conf
	Rules            []*rules_inbound.Rule
	SkipCommonConfig bool
}

func (rc *ListenerConfigurer) ConfigureListener(listener *envoy_listener.Listener) error {
	if listener == nil {
		return nil
	}

	for _, filterChain := range listener.FilterChains {
		if !rc.SkipCommonConfig {
			if err := ConfigureFilterChain(rc.Conf, filterChain); err != nil {
				return err
			}
		}
		if err := ConfigureMatchedRoutesOnFilterChain(filterChain, rc.Rules); err != nil {
			return err
		}
	}

	return nil
}

func configureRequestHeadersTimeout(conf api.Conf, hcm *envoy_hcm.HttpConnectionManager) {
	hcm.RequestHeadersTimeout = toProtoDurationOrDefault(
		pointer.Deref(conf.Http).RequestHeadersTimeout,
		policies_defaults.DefaultRequestHeadersTimeout,
	)
}

func configureRouteConfigurationTimeouts(routeConfiguration *envoy_route.RouteConfiguration, conf api.Conf) {
	if routeConfiguration == nil {
		return
	}
	for _, vh := range routeConfiguration.VirtualHosts {
		for _, route := range vh.Routes {
			ConfigureRouteAction(
				route.GetRoute(),
				pointer.Deref(conf.Http).RequestTimeout,
				pointer.Deref(conf.Http).StreamIdleTimeout,
			)
		}
	}
}

type effectiveMatchedRouteRule struct {
	Match *common_api.Match
	Conf  api.Conf
}

func matchedRouteTimeoutRules(inboundRules []*rules_inbound.Rule) []*rules_inbound.Rule {
	var matched []*rules_inbound.Rule
	for _, rule := range inboundRules {
		if !hasSpiffeIDMatch(rule.Match) {
			continue
		}
		conf, ok := rule.Conf.(api.Conf)
		if !ok {
			continue
		}
		http := pointer.Deref(conf.Http)
		if http.RequestTimeout == nil && http.StreamIdleTimeout == nil {
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

func effectiveMatchedRouteRules(matchedRules []*rules_inbound.Rule) ([]effectiveMatchedRouteRule, error) {
	var effective []effectiveMatchedRouteRule
	for _, rule := range matchedRules {
		if containsEffectiveMatchedRoute(effective, rule.Match) {
			continue
		}

		conf, ok, err := mergeSubsumingMatchedRouteConfs(matchedRules, rule.Match)
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}
		effective = append(effective, effectiveMatchedRouteRule{
			Match: rule.Match,
			Conf:  conf,
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
	switch {
	case a == nil || b == nil:
		return a == b
	case a.SpiffeID == nil || b.SpiffeID == nil:
		return a.SpiffeID == b.SpiffeID
	default:
		return a.SpiffeID.Type == b.SpiffeID.Type && a.SpiffeID.Value == b.SpiffeID.Value
	}
}

func mergeSubsumingMatchedRouteConfs(matchedRules []*rules_inbound.Rule, target *common_api.Match) (api.Conf, bool, error) {
	var applicable []*rules_inbound.Rule
	for _, rule := range matchedRules {
		if matchedRouteSubsumes(rule.Match, target) {
			applicable = append(applicable, rule)
		}
	}
	if len(applicable) == 0 {
		return api.Conf{}, false, nil
	}

	slices.SortStableFunc(applicable, compareMatchedRouteRuleSpecificityAsc)

	confs := make([]any, 0, len(applicable))
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

func matchedRouteSubsumes(candidate, target *common_api.Match) bool {
	if candidate == nil || target == nil || candidate.SpiffeID == nil || target.SpiffeID == nil {
		return false
	}

	switch candidate.SpiffeID.Type {
	case common_api.ExactMatchType:
		return target.SpiffeID.Type == common_api.ExactMatchType && candidate.SpiffeID.Value == target.SpiffeID.Value
	case common_api.PrefixMatchType:
		return strings.HasPrefix(target.SpiffeID.Value, candidate.SpiffeID.Value)
	default:
		return false
	}
}

func compareMatchedRouteRuleSpecificityDesc(a, b *rules_inbound.Rule) int {
	return compareMatchedRouteSpecificity(a.Match, b.Match, true)
}

func compareMatchedRouteRuleSpecificityAsc(a, b *rules_inbound.Rule) int {
	return compareMatchedRouteSpecificity(a.Match, b.Match, false)
}

// Keep this MeshTimeout-local instead of reusing inbound.CompareMatch.
// Route precedence needs both ascending and descending order, and for Prefix
// matches it also needs a longer-prefix tie-break that the generic inbound
// sorter intentionally does not provide.
func compareMatchedRouteSpecificity(a, b *common_api.Match, descending bool) int {
	aRank, aLen := matchedRouteSpecificity(a)
	bRank, bLen := matchedRouteSpecificity(b)

	switch {
	case descending && aRank > bRank:
		return -1
	case descending && aRank < bRank:
		return 1
	case !descending && aRank < bRank:
		return -1
	case !descending && aRank > bRank:
		return 1
	}

	if aRank != 1 || bRank != 1 {
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

func matchedRouteSpecificity(match *common_api.Match) (int, int) {
	if match == nil || match.SpiffeID == nil {
		return 0, 0
	}

	switch match.SpiffeID.Type {
	case common_api.ExactMatchType:
		return 2, len(match.SpiffeID.Value)
	case common_api.PrefixMatchType:
		return 1, len(match.SpiffeID.Value)
	default:
		return 0, len(match.SpiffeID.Value)
	}
}

func duplicateMatchedRoutes(route *envoy_route.Route, inboundRules []effectiveMatchedRouteRule) []*envoy_route.Route {
	var duplicated []*envoy_route.Route
	for _, rule := range inboundRules {
		clone := proto.Clone(route).(*envoy_route.Route)
		if clone.Match == nil || clone.GetRoute() == nil {
			continue
		}
		clone.Match.FilterState = append(clone.Match.FilterState, routeFilterStateMatchers(rule.Match)...)
		ConfigureMatchedRouteAction(
			clone.GetRoute(),
			pointer.Deref(rule.Conf.Http).RequestTimeout,
			pointer.Deref(rule.Conf.Http).StreamIdleTimeout,
		)
		duplicated = append(duplicated, clone)
	}
	return duplicated
}

func ConfigureMatchedRouteAction(
	routeAction *envoy_route.RouteAction,
	httpRequestTimeout *kube_meta.Duration,
	httpStreamIdleTimeout *kube_meta.Duration,
) {
	if routeAction == nil {
		return
	}
	if httpRequestTimeout != nil {
		routeAction.Timeout = util_proto.Duration(httpRequestTimeout.Duration)
	}
	if httpStreamIdleTimeout != nil {
		routeAction.IdleTimeout = util_proto.Duration(httpStreamIdleTimeout.Duration)
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

func ensureMatchFilterStateFilter(filterChain *envoy_listener.FilterChain, inboundRules []*rules_inbound.Rule) error {
	if filterChain == nil || len(matchedRouteTimeoutRules(inboundRules)) == 0 {
		return nil
	}
	for _, filter := range filterChain.Filters {
		if filter.Name == matchFilterStateNetworkFilterName {
			return nil
		}
	}

	config := &envoy_set_filter_state.Config{}
	if hasAnySpiffeIDMatch(inboundRules) {
		config.OnDownstreamTlsHandshake = append(config.OnDownstreamTlsHandshake, newFilterStateValue(matchSpiffeIDFilterStateKey, "%DOWNSTREAM_PEER_URI_SAN%"))
	}
	if hasAnySNIMatch(inboundRules) {
		config.OnDownstreamTlsHandshake = append(config.OnDownstreamTlsHandshake, newFilterStateValue(matchSNIStateKey, "%REQUESTED_SERVER_NAME%"))
	}

	typedConfig, err := util_proto.MarshalAnyDeterministic(config)
	if err != nil {
		return err
	}

	filterChain.Filters = append([]*envoy_listener.Filter{{
		Name:       matchFilterStateNetworkFilterName,
		ConfigType: &envoy_listener.Filter_TypedConfig{TypedConfig: typedConfig},
	}}, filterChain.Filters...)
	return nil
}

func hasAnySpiffeIDMatch(inboundRules []*rules_inbound.Rule) bool {
	for _, rule := range inboundRules {
		if hasSpiffeIDMatch(rule.Match) {
			return true
		}
	}
	return false
}

func hasAnySNIMatch(inboundRules []*rules_inbound.Rule) bool {
	for _, rule := range inboundRules {
		if rule.Match != nil && rule.Match.SNI != nil {
			return true
		}
	}
	return false
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
