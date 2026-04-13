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
// Splits on application protocol: HTTP → HTTP filter chains, fallback → TCP filter chains.
func buildRawBufferMatcher(matches []FilterChainMatch, isIPv6 bool) (*matcher_config.Matcher_OnMatch, error) {
	var httpMatches []FilterChainMatch
	var tcpMatches []FilterChainMatch

	for _, m := range matches {
		if !matchesIPVersion(m, isIPv6) {
			continue
		}
		switch m.Protocol {
		case core_meta.ProtocolHTTP, core_meta.ProtocolHTTP2, core_meta.ProtocolGRPC:
			httpMatches = append(httpMatches, m)
		case core_meta.ProtocolTCP, core_meta.Protocol(api.MysqlProtocol):
			tcpMatches = append(tcpMatches, m)
		}
	}

	if len(httpMatches) == 0 && len(tcpMatches) == 0 {
		return nil, nil
	}

	var tcpOnMatch *matcher_config.Matcher_OnMatch
	if len(tcpMatches) > 0 {
		tcpOnMatch = buildIPCIDRMatcherList(tcpMatches, core_meta.ProtocolTCP)
	}

	if len(httpMatches) == 0 {
		return tcpOnMatch, nil
	}

	return buildHTTPAppProtocolMatcher(httpMatches, tcpOnMatch)
}

// buildHTTPAppProtocolMatcher maps "http/1.1" and "h2c" to HTTP filter chain port matchers.
func buildHTTPAppProtocolMatcher(httpMatches []FilterChainMatch, tcpFallback *matcher_config.Matcher_OnMatch) (*matcher_config.Matcher_OnMatch, error) {
	httpPortOnMatch, err := buildHTTPPortMatcher(httpMatches)
	if err != nil {
		return nil, err
	}

	entries := map[string]*matcher_config.Matcher_OnMatch{
		"http/1.1": httpPortOnMatch,
		"h2c":      httpPortOnMatch,
	}

	m := &matcher_config.Matcher{OnNoMatch: tcpFallback}
	if err := bldrs_matchers.MatcherTreeExactMap(bldrs_matchers.NetworkInputApplicationProtocol(), entries)(m); err != nil {
		return nil, err
	}
	return bldrs_matchers.OnMatchMatcher(m), nil
}

// buildHTTPPortMatcher builds port-based matchers for HTTP filter chains.
// For HTTP protocols the filter chain name is derived from protocol+port (not value).
func buildHTTPPortMatcher(httpMatches []FilterChainMatch) (*matcher_config.Matcher_OnMatch, error) {
	portEntries := map[string]*matcher_config.Matcher_OnMatch{}
	var wildcard *matcher_config.Matcher_OnMatch

	for _, m := range httpMatches {
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

// buildIPCIDRMatcherList builds a MatcherList for IP/CIDR-based filter chain selection.
func buildIPCIDRMatcherList(ipMatches []FilterChainMatch, protocol core_meta.Protocol) *matcher_config.Matcher_OnMatch {
	var fieldMatchers []*matcher_config.Matcher_MatcherList_FieldMatcher

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
		action := bldrs_matchers.FilterChainNameOnMatch(chainName)
		portedAction := wrapWithPortMatch(action, m.Port)

		fieldMatchers = append(fieldMatchers, buildIPFieldMatcher([]*corev3.CidrRange{cidr}, chainName, portedAction))
	}

	if len(fieldMatchers) == 0 {
		return nil
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

// wrapWithPortMatch wraps an action in a port-exact matcher when port != 0.
func wrapWithPortMatch(action *matcher_config.Matcher_OnMatch, port uint32) *matcher_config.Matcher_OnMatch {
	if port == 0 {
		return action
	}
	portMatcher := &matcher_config.Matcher{}
	entries := map[string]*matcher_config.Matcher_OnMatch{
		fmt.Sprintf("%d", port): action,
	}
	if err := bldrs_matchers.MatcherTreeExactMap(bldrs_matchers.NetworkInputDestinationPort(), entries)(portMatcher); err != nil {
		panic(err)
	}
	return bldrs_matchers.OnMatchMatcher(portMatcher)
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
