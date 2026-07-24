package xds

import (
	"fmt"
	"sort"

	xds_core "github.com/cncf/xds/go/xds/core/v3"
	matcher_config "github.com/cncf/xds/go/xds/type/matcher/v3"
	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	ipv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/matching/input_matchers/ip/v3"

	core_meta "github.com/kumahq/kuma/v3/pkg/core/metadata"
	bldrs_matchers "github.com/kumahq/kuma/v3/pkg/envoy/builders/xds/matchers"
	api "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshpassthrough/api/v1alpha1"
	util_proto "github.com/kumahq/kuma/v3/pkg/util/proto"
)

// BuildFilterChainMatcher builds an xds.type.matcher.v3.Matcher tree that selects
// filter chains by name based on network properties (transport protocol, SNI, port, IP).
// This replaces per-filter-chain FilterChainMatch fields with a single matcher tree on the Listener.
func BuildFilterChainMatcher(matches []FilterChainMatch, isIPv6 bool) (*matcher_config.Matcher, error) {
	tlsOnMatch, err := buildTLSMatcher(matches, isIPv6)
	if err != nil {
		return nil, err
	}
	rawBufferOnMatch, err := buildRawBufferMatcher(matches, isIPv6)
	if err != nil {
		return nil, err
	}

	entries := map[string]*matcher_config.Matcher_OnMatch{}
	if tlsOnMatch != nil {
		entries["tls"] = tlsOnMatch
	}
	if rawBufferOnMatch != nil {
		entries["raw_buffer"] = rawBufferOnMatch
	}
	if len(entries) == 0 {
		return nil, nil
	}

	m := &matcher_config.Matcher{}
	if err := bldrs_matchers.MatcherTreeExactMap(bldrs_matchers.NetworkInputTransportProtocol(), entries)(m); err != nil {
		return nil, err
	}
	return m, nil
}

// serverNameMatcherExtensionName is the registered name of Envoy's domain-aware SNI custom
// matcher (config proto xds.type.matcher.v3.ServerNameMatcher). Unlike a prefix_match_map on
// ServerNameInput — which does a literal character-prefix match and never matches a real
// subdomain — this matcher treats "*.example.com" as a domain suffix and evaluates
// longest-suffix-first.
const serverNameMatcherExtensionName = "envoy.matching.custom_matchers.domain_matcher"

// buildTLSMatcher builds the "tls" transport protocol branch.
// Matches SNI (exact + wildcard domains) via a ServerNameMatcher, then destination port, with an
// IP/CIDR fallback when the SNI does not match any configured domain.
func buildTLSMatcher(matches []FilterChainMatch, isIPv6 bool) (*matcher_config.Matcher_OnMatch, error) {
	var tlsMatches []FilterChainMatch
	for _, m := range matches {
		if m.Protocol == core_meta.ProtocolTLS && matchesIPVersion(m, isIPv6) {
			tlsMatches = append(tlsMatches, m)
		}
	}
	if len(tlsMatches) == 0 {
		return nil, nil
	}

	// Group SNI (exact + wildcard) matches by domain value; IP/CIDR matches become the fallback.
	sniMatches := map[string][]FilterChainMatch{}
	var sniOrder []string
	var ipCIDRMatches []FilterChainMatch
	for _, m := range tlsMatches {
		switch m.MatchType {
		case Domain, WildcardDomain:
			if _, ok := sniMatches[m.Value]; !ok {
				sniOrder = append(sniOrder, m.Value)
			}
			sniMatches[m.Value] = append(sniMatches[m.Value], m)
		case IP, IPV6, CIDR, CIDRV6:
			ipCIDRMatches = append(ipCIDRMatches, m)
		}
	}

	var ipCIDRFallback *matcher_config.Matcher_OnMatch
	if len(ipCIDRMatches) > 0 {
		ipCIDRFallback = buildIPCIDRMatcherList(ipCIDRMatches, core_meta.ProtocolTLS)
	}

	if len(sniMatches) == 0 {
		// Only IP/CIDR TLS matches (no SNI): fall straight through to the IP matcher list.
		return bldrs_matchers.OnMatchMatcher(&matcher_config.Matcher{OnNoMatch: ipCIDRFallback}), nil
	}

	// One DomainMatcher per SNI value. ServerNameMatcher does domain-aware longest-suffix
	// matching, so exact domains automatically win over wildcards regardless of list order.
	sort.Strings(sniOrder)
	var domainMatchers []*matcher_config.ServerNameMatcher_DomainMatcher
	for _, sni := range sniOrder {
		onMatch, err := buildPortMatcher(sniMatches[sni])
		if err != nil {
			return nil, err
		}
		domainMatchers = append(domainMatchers, &matcher_config.ServerNameMatcher_DomainMatcher{
			Domains: []string{sni},
			OnMatch: onMatch,
		})
	}

	sniMatcher, err := buildServerNameMatcher(domainMatchers, ipCIDRFallback)
	if err != nil {
		return nil, err
	}
	return bldrs_matchers.OnMatchMatcher(sniMatcher), nil
}

// buildServerNameMatcher wraps the domain matchers in a ServerNameMatcher custom_match. The
// domain matcher matches against whatever string the MatcherTree input provides, so the input
// MUST be ServerNameInput (the raw SNI). IP/CIDR matches (if any) are the on_no_match fallback.
func buildServerNameMatcher(
	domainMatchers []*matcher_config.ServerNameMatcher_DomainMatcher,
	ipCIDRFallback *matcher_config.Matcher_OnMatch,
) (*matcher_config.Matcher, error) {
	serverNameMatcher := &matcher_config.ServerNameMatcher{DomainMatchers: domainMatchers}
	typedConfig, err := util_proto.MarshalAnyDeterministic(serverNameMatcher)
	if err != nil {
		return nil, err
	}
	return &matcher_config.Matcher{
		MatcherType: &matcher_config.Matcher_MatcherTree_{
			MatcherTree: &matcher_config.Matcher_MatcherTree{
				Input: bldrs_matchers.NetworkInputServerName(),
				TreeType: &matcher_config.Matcher_MatcherTree_CustomMatch{
					CustomMatch: &xds_core.TypedExtensionConfig{
						Name:        serverNameMatcherExtensionName,
						TypedConfig: typedConfig,
					},
				},
			},
		},
		OnNoMatch: ipCIDRFallback,
	}, nil
}

// buildRawBufferMatcher builds the "raw_buffer" transport protocol branch.
// Domain/WildcardDomain HTTP matches use port-based selection (filter chain names derived from protocol+port).
// IP/CIDR matches for any protocol use destination IP matching (filter chain names derived from IP value).
// ApplicationProtocolInput is intentionally avoided: it does not reliably populate in Envoy 1.37.x
// when used in the filter_chain_matcher context (listener stat no_filter_chain_match fires instead).
func buildRawBufferMatcher(matches []FilterChainMatch, isIPv6 bool) (*matcher_config.Matcher_OnMatch, error) {
	var httpDomainMatches []FilterChainMatch // Domain/WildcardDomain + HTTP/HTTP2/GRPC
	var ipBasedMatches []FilterChainMatch    // IP/CIDR + any non-TLS protocol

	for _, m := range matches {
		if !matchesIPVersion(m, isIPv6) {
			continue
		}
		switch m.Protocol {
		case core_meta.ProtocolHTTP, core_meta.ProtocolHTTP2, core_meta.ProtocolGRPC:
			switch m.MatchType {
			case Domain, WildcardDomain:
				httpDomainMatches = append(httpDomainMatches, m)
			case IP, IPV6, CIDR, CIDRV6:
				ipBasedMatches = append(ipBasedMatches, m)
			}
		case core_meta.ProtocolTCP, core_meta.Protocol(api.MysqlProtocol):
			switch m.MatchType {
			case IP, IPV6, CIDR, CIDRV6:
				ipBasedMatches = append(ipBasedMatches, m)
			}
		}
	}

	if len(httpDomainMatches) == 0 && len(ipBasedMatches) == 0 {
		return nil, nil
	}

	// Domain HTTP port matcher acts as the fallback when no IP/CIDR matches.
	var domainHTTPOnMatch *matcher_config.Matcher_OnMatch
	if len(httpDomainMatches) > 0 {
		var err error
		domainHTTPOnMatch, err = buildHTTPPortMatcher(httpDomainMatches)
		if err != nil {
			return nil, err
		}
	}

	if len(ipBasedMatches) == 0 {
		return domainHTTPOnMatch, nil
	}

	return buildMixedIPCIDRMatcherList(ipBasedMatches, domainHTTPOnMatch), nil
}

// buildHTTPPortMatcher builds port-based matchers for Domain/WildcardDomain HTTP filter chains.
// Filter chain names are derived from protocol+port (e.g. meshpassthrough_http_80).
func buildHTTPPortMatcher(httpDomainMatches []FilterChainMatch) (*matcher_config.Matcher_OnMatch, error) {
	portEntries := map[string]*matcher_config.Matcher_OnMatch{}
	var wildcard *matcher_config.Matcher_OnMatch

	for _, m := range httpDomainMatches {
		chainName := FilterChainName(string(m.Protocol), m.Protocol, m.Port)
		action := bldrs_matchers.FilterChainNameOnMatch(chainName)
		if m.Port == 0 {
			wildcard = action
		} else {
			portEntries[fmt.Sprintf("%d", m.Port)] = action
		}
	}

	if len(portEntries) == 0 {
		return wildcard, nil
	}

	m := &matcher_config.Matcher{OnNoMatch: wildcard}
	if err := bldrs_matchers.MatcherTreeExactMap(bldrs_matchers.NetworkInputDestinationPort(), portEntries)(m); err != nil {
		return nil, err
	}
	return bldrs_matchers.OnMatchMatcher(m), nil
}

// buildMixedIPCIDRMatcherList builds a MatcherList for IP/CIDR destination matching across
// multiple protocols (HTTP, TCP, MySQL). All port variants for the same IP/CIDR are collapsed
// into a single FieldMatcher to avoid the MatcherList "first predicate match wins" problem:
// if two FieldMatchers share the same IP predicate, only the first is ever evaluated.
// Domain HTTP matches are used as the OnNoMatch fallback.
func buildMixedIPCIDRMatcherList(ipMatches []FilterChainMatch, fallback *matcher_config.Matcher_OnMatch) *matcher_config.Matcher_OnMatch {
	type ipGroup struct {
		cidr  *corev3.CidrRange
		ports []portChain
	}

	groups := map[string]*ipGroup{}
	var orderedKeys []string

	for _, m := range ipMatches {
		var cidr *corev3.CidrRange
		switch m.MatchType {
		case IP, IPV6:
			cidr = ipToCIDR(m.Value, m.MatchType == IPV6)
		case CIDR, CIDRV6:
			ip, mask := getIpAndMask(m.Value)
			cidr = &corev3.CidrRange{AddressPrefix: ip, PrefixLen: util_proto.UInt32(mask)}
		default:
			continue
		}

		// MySQL filter chains are created by configureTcpFilterChain which uses ProtocolTCP for naming.
		chainProtocol := m.Protocol
		if m.Protocol == core_meta.Protocol(api.MysqlProtocol) {
			chainProtocol = core_meta.ProtocolTCP
		}
		chainName := FilterChainName(m.Value, chainProtocol, m.Port)

		if _, exists := groups[m.Value]; !exists {
			groups[m.Value] = &ipGroup{cidr: cidr}
			orderedKeys = append(orderedKeys, m.Value)
		}
		groups[m.Value].ports = append(groups[m.Value].ports, portChain{m.Port, chainName})
	}

	if len(groups) == 0 {
		return fallback
	}

	var fieldMatchers []*matcher_config.Matcher_MatcherList_FieldMatcher
	for _, key := range orderedKeys {
		group := groups[key]
		onMatch := buildPortOnMatchForGroup(group.ports)
		fieldMatchers = append(fieldMatchers, buildIPFieldMatcher([]*corev3.CidrRange{group.cidr}, key, onMatch))
	}

	nested := &matcher_config.Matcher{
		MatcherType: &matcher_config.Matcher_MatcherList_{
			MatcherList: &matcher_config.Matcher_MatcherList{Matchers: fieldMatchers},
		},
		OnNoMatch: fallback,
	}
	return bldrs_matchers.OnMatchMatcher(nested)
}

type portChain struct {
	port      uint32
	chainName string
}

// buildPortOnMatchForGroup builds a port ExactMap (with optional wildcard OnNoMatch) from a set
// of (port, chainName) entries sharing the same IP/CIDR predicate.
func buildPortOnMatchForGroup(ports []portChain) *matcher_config.Matcher_OnMatch {
	portMap := map[string]*matcher_config.Matcher_OnMatch{}
	var wildcardAction *matcher_config.Matcher_OnMatch
	for _, p := range ports {
		action := bldrs_matchers.FilterChainNameOnMatch(p.chainName)
		if p.port == 0 {
			wildcardAction = action
		} else {
			portMap[fmt.Sprintf("%d", p.port)] = action
		}
	}
	if len(portMap) == 0 {
		return wildcardAction
	}
	m := &matcher_config.Matcher{OnNoMatch: wildcardAction}
	if err := bldrs_matchers.MatcherTreeExactMap(bldrs_matchers.NetworkInputDestinationPort(), portMap)(m); err != nil {
		panic(err)
	}
	return bldrs_matchers.OnMatchMatcher(m)
}

// buildPortMatcher builds a destination-port matcher for filter chains sharing the same SNI/address.
func buildPortMatcher(chainMatches []FilterChainMatch) (*matcher_config.Matcher_OnMatch, error) {
	portEntries := map[string]*matcher_config.Matcher_OnMatch{}
	var wildcard *matcher_config.Matcher_OnMatch

	for _, m := range chainMatches {
		chainName := FilterChainName(m.Value, m.Protocol, m.Port)
		action := bldrs_matchers.FilterChainNameOnMatch(chainName)
		if m.Port == 0 {
			wildcard = action
		} else {
			portEntries[fmt.Sprintf("%d", m.Port)] = action
		}
	}

	if len(portEntries) == 0 {
		return wildcard, nil
	}

	m := &matcher_config.Matcher{OnNoMatch: wildcard}
	if err := bldrs_matchers.MatcherTreeExactMap(bldrs_matchers.NetworkInputDestinationPort(), portEntries)(m); err != nil {
		return nil, err
	}
	return bldrs_matchers.OnMatchMatcher(m), nil
}

// buildIPCIDRMatcherList builds a MatcherList for IP/CIDR-based filter chain selection (TLS branch).
// All port variants for the same IP/CIDR are collapsed into a single FieldMatcher — see
// buildMixedIPCIDRMatcherList for the rationale.
func buildIPCIDRMatcherList(ipMatches []FilterChainMatch, protocol core_meta.Protocol) *matcher_config.Matcher_OnMatch {
	type ipGroup struct {
		cidr  *corev3.CidrRange
		ports []portChain
	}

	groups := map[string]*ipGroup{}
	var orderedKeys []string

	for _, m := range ipMatches {
		var cidr *corev3.CidrRange
		switch m.MatchType {
		case IP, IPV6:
			cidr = ipToCIDR(m.Value, m.MatchType == IPV6)
		case CIDR, CIDRV6:
			ip, mask := getIpAndMask(m.Value)
			cidr = &corev3.CidrRange{AddressPrefix: ip, PrefixLen: util_proto.UInt32(mask)}
		default:
			continue
		}

		chainName := FilterChainName(m.Value, protocol, m.Port)

		if _, exists := groups[m.Value]; !exists {
			groups[m.Value] = &ipGroup{cidr: cidr}
			orderedKeys = append(orderedKeys, m.Value)
		}
		groups[m.Value].ports = append(groups[m.Value].ports, portChain{m.Port, chainName})
	}

	if len(groups) == 0 {
		return nil
	}

	var fieldMatchers []*matcher_config.Matcher_MatcherList_FieldMatcher
	for _, key := range orderedKeys {
		group := groups[key]
		onMatch := buildPortOnMatchForGroup(group.ports)
		fieldMatchers = append(fieldMatchers, buildIPFieldMatcher([]*corev3.CidrRange{group.cidr}, key, onMatch))
	}

	nested := &matcher_config.Matcher{
		MatcherType: &matcher_config.Matcher_MatcherList_{
			MatcherList: &matcher_config.Matcher_MatcherList{Matchers: fieldMatchers},
		},
	}
	return bldrs_matchers.OnMatchMatcher(nested)
}

func buildIPFieldMatcher(cidrRanges []*corev3.CidrRange, statPrefix string, action *matcher_config.Matcher_OnMatch) *matcher_config.Matcher_MatcherList_FieldMatcher {
	ipMatcherProto, err := util_proto.MarshalAnyDeterministic(&ipv3.Ip{
		CidrRanges: cidrRanges,
		StatPrefix: statPrefix,
	})
	if err != nil {
		panic(err)
	}
	return &matcher_config.Matcher_MatcherList_FieldMatcher{
		Predicate: &matcher_config.Matcher_MatcherList_Predicate{
			MatchType: &matcher_config.Matcher_MatcherList_Predicate_SinglePredicate_{
				SinglePredicate: &matcher_config.Matcher_MatcherList_Predicate_SinglePredicate{
					Input: bldrs_matchers.NetworkInputDestinationIP(),
					Matcher: &matcher_config.Matcher_MatcherList_Predicate_SinglePredicate_CustomMatch{
						CustomMatch: &xds_core.TypedExtensionConfig{
							Name:        "envoy.matching.matchers.ip",
							TypedConfig: ipMatcherProto,
						},
					},
				},
			},
		},
		OnMatch: action,
	}
}

func ipToCIDR(ip string, isIPv6 bool) *corev3.CidrRange {
	prefixLen := uint32(32)
	if isIPv6 {
		prefixLen = 128
	}
	return &corev3.CidrRange{
		AddressPrefix: ip,
		PrefixLen:     util_proto.UInt32(prefixLen),
	}
}

func matchesIPVersion(m FilterChainMatch, isIPv6 bool) bool {
	if isIPv6 {
		return m.MatchType == Domain || m.MatchType == WildcardDomain || m.MatchType == CIDRV6 || m.MatchType == IPV6
	}
	return m.MatchType == Domain || m.MatchType == WildcardDomain || m.MatchType == CIDR || m.MatchType == IP
}
