package xds

import (
	"fmt"

	xds_core "github.com/cncf/xds/go/xds/core/v3"
	matcher_config "github.com/cncf/xds/go/xds/type/matcher/v3"
	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	ipv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/matching/input_matchers/ip/v3"

	core_meta "github.com/kumahq/kuma/v2/pkg/core/metadata"
	bldrs_matchers "github.com/kumahq/kuma/v2/pkg/envoy/builders/xds/matchers"
	api "github.com/kumahq/kuma/v2/pkg/plugins/policies/meshpassthrough/api/v1alpha1"
	util_proto "github.com/kumahq/kuma/v2/pkg/util/proto"
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

// buildTLSMatcher builds the "tls" transport protocol branch.
// Matches SNI (exact/wildcard domains), then destination port, with IP/CIDR fallback.
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

	exactSNI := map[string][]FilterChainMatch{}
	wildcardSNI := map[string][]FilterChainMatch{}
	var ipCIDRMatches []FilterChainMatch

	for _, m := range tlsMatches {
		switch m.MatchType {
		case Domain:
			exactSNI[m.Value] = append(exactSNI[m.Value], m)
		case WildcardDomain:
			prefix := wildcardDomainToPrefix(m.Value)
			wildcardSNI[prefix] = append(wildcardSNI[prefix], m)
		case IP, IPV6, CIDR, CIDRV6:
			ipCIDRMatches = append(ipCIDRMatches, m)
		}
	}

	exactEntries := map[string]*matcher_config.Matcher_OnMatch{}
	for sni, chainMatches := range exactSNI {
		onMatch, err := buildPortMatcher(chainMatches)
		if err != nil {
			return nil, err
		}
		exactEntries[sni] = onMatch
	}

	prefixEntries := map[string]*matcher_config.Matcher_OnMatch{}
	for prefix, chainMatches := range wildcardSNI {
		onMatch, err := buildPortMatcher(chainMatches)
		if err != nil {
			return nil, err
		}
		prefixEntries[prefix] = onMatch
	}

	var ipCIDRFallback *matcher_config.Matcher_OnMatch
	if len(ipCIDRMatches) > 0 {
		ipCIDRFallback = buildIPCIDRMatcherList(ipCIDRMatches, core_meta.ProtocolTLS)
	}

	sniMatcher, err := buildSNITree(exactEntries, prefixEntries, ipCIDRFallback)
	if err != nil {
		return nil, err
	}
	return bldrs_matchers.OnMatchMatcher(sniMatcher), nil
}

func buildSNITree(
	exactEntries map[string]*matcher_config.Matcher_OnMatch,
	prefixEntries map[string]*matcher_config.Matcher_OnMatch,
	ipCIDRFallback *matcher_config.Matcher_OnMatch,
) (*matcher_config.Matcher, error) {
	// Build prefix matcher (wildcard domains) with IP/CIDR as its fallback.
	var prefixMatcher *matcher_config.Matcher
	if len(prefixEntries) > 0 {
		prefixMatcher = &matcher_config.Matcher{OnNoMatch: ipCIDRFallback}
		if err := bldrs_matchers.MatcherTreePrefixMap(bldrs_matchers.NetworkInputServerName(), prefixEntries)(prefixMatcher); err != nil {
			return nil, err
		}
	}

	// Build exact matcher with prefix matcher (or IP/CIDR) as fallback.
	var onNoMatch *matcher_config.Matcher_OnMatch
	if prefixMatcher != nil {
		onNoMatch = bldrs_matchers.OnMatchMatcher(prefixMatcher)
	} else {
		onNoMatch = ipCIDRFallback
	}

	if len(exactEntries) == 0 {
		// Only prefix or IP/CIDR matchers. Return the prefix matcher if exists, otherwise a
		// passthrough matcher that immediately falls back to IP/CIDR.
		if prefixMatcher != nil {
			return prefixMatcher, nil
		}
		return &matcher_config.Matcher{OnNoMatch: ipCIDRFallback}, nil
	}

	m := &matcher_config.Matcher{OnNoMatch: onNoMatch}
	if err := bldrs_matchers.MatcherTreeExactMap(bldrs_matchers.NetworkInputServerName(), exactEntries)(m); err != nil {
		return nil, err
	}
	return m, nil
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

// wildcardDomainToPrefix converts *.example.com to .example.com for prefix matching.
func wildcardDomainToPrefix(domain string) string {
	if len(domain) > 1 && domain[0] == '*' {
		return domain[1:]
	}
	return domain
}

func matchesIPVersion(m FilterChainMatch, isIPv6 bool) bool {
	if isIPv6 {
		return m.MatchType == Domain || m.MatchType == WildcardDomain || m.MatchType == CIDRV6 || m.MatchType == IPV6
	}
	return m.MatchType == Domain || m.MatchType == WildcardDomain || m.MatchType == CIDR || m.MatchType == IP
}
