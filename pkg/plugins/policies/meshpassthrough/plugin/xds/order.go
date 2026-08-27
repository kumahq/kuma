package xds

import (
	"fmt"
	"maps"
	"net"
	"slices"
	"sort"
	"strings"

	"github.com/asaskevich/govalidator"

	core_meta "github.com/kumahq/kuma/v3/pkg/core/metadata"
	api "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshpassthrough/api/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/util/pointer"
)

type MatchType int

const (
	WildcardDomain MatchType = iota + 1
	Domain
	CIDR
	CIDRV6
	IP
	IPV6
)

// l7Protocols produce identical Envoy filter chain matchers, so only one of them
// can be configured on a given port. More than one makes Envoy reject the listener.
var l7Protocols = []core_meta.Protocol{core_meta.ProtocolHTTP, core_meta.ProtocolHTTP2, core_meta.ProtocolGRPC}

var protocolOrder = map[core_meta.Protocol]int{
	core_meta.ProtocolTLS:   0,
	core_meta.ProtocolTCP:   1,
	core_meta.ProtocolHTTP:  2,
	core_meta.ProtocolHTTP2: 3,
	core_meta.ProtocolGRPC:  4,
}

type Route struct {
	Value     string
	MatchType MatchType
}

type Matcher struct {
	Protocol  core_meta.Protocol
	Port      uint32
	MatchType MatchType
	Value     string
}

type FilterChainMatch struct {
	Protocol  core_meta.Protocol
	Port      uint32
	MatchType MatchType
	Value     string
	Routes    []Route
}

// GetOrderedMatchers builds filter chain matchers for the given configuration. Matches
// that would result in a listener rejected by Envoy are dropped and returned as warnings,
// so a single incorrect match doesn't invalidate the whole passthrough listener.
func GetOrderedMatchers(conf api.Conf) ([]FilterChainMatch, []string) {
	warnings := newWarnings()
	matcherWithRoutes := map[Matcher]map[Route]bool{}
	portProtocols := map[uint32]map[core_meta.Protocol]bool{}
	chainOwners := map[chainKey]chainOwner{}
	for _, match := range pointer.Deref(conf.AppendMatch) {
		port := pointer.DerefOr[uint32](match.Port, 0)
		protocol := core_meta.ParseProtocol(string(match.Protocol))
		matchType, isWildcardDomain := getMatchType(match, protocol)
		matcher := Matcher{
			Protocol:  protocol,
			Port:      port,
			MatchType: matchType,
		}
		// L7 domains share one filter chain per port and become virtual host routes
		isL7Domain := slices.Contains(l7Protocols, protocol) && matchType == Domain
		if !isL7Domain {
			matcher.Value = match.Value
		}
		key := newChainKey(port, protocol, matchType, match.Value)
		if owner, found := chainOwners[key]; found {
			if owner.protocol != protocol {
				warnings.add(fmt.Sprintf(
					"ignoring match %q with protocol %s, protocol %s is already configured for %s on %s, %s",
					match.Value, protocol, owner.protocol, key.describe(), describePort(port), key.conflictReason(),
				))
				continue
			}
			if owner.matcher != matcher {
				warnings.add(fmt.Sprintf(
					"ignoring match %q with protocol %s, match %q already defines a filter chain for %s on %s",
					match.Value, protocol, owner.value, key.describe(), describePort(port),
				))
				continue
			}
		} else {
			chainOwners[key] = chainOwner{protocol: protocol, matcher: matcher, value: match.Value}
		}
		if _, found := portProtocols[port]; !found {
			portProtocols[port] = map[core_meta.Protocol]bool{protocol: true}
		} else {
			portProtocols[port][protocol] = true
		}
		if isL7Domain {
			routeMatchType := Domain
			if isWildcardDomain {
				routeMatchType = WildcardDomain
			}
			route := Route{
				Value:     match.Value,
				MatchType: routeMatchType,
			}
			if _, found := matcherWithRoutes[matcher]; found {
				matcherWithRoutes[matcher][route] = true
			} else {
				matcherWithRoutes[matcher] = map[Route]bool{
					route: true,
				}
			}
		} else if _, found := matcherWithRoutes[matcher]; !found {
			matcherWithRoutes[matcher] = map[Route]bool{}
		}
	}
	// Envoy first checks the port when performing matching. If there is a matcher for a specific port
	// and one rule to match all ports alongside another for a specific port,
	// it might select the matcher for the specific port but fail to find a corresponding filter chain.
	// To avoid this issue, we also generate specific port matchers for rules intended to match all ports.
	matcherWithRoutesAndAdditionalPorts := map[Matcher]map[Route]bool{}
	for matcher, routes := range matcherWithRoutes {
		if _, found := matcherWithRoutesAndAdditionalPorts[matcher]; found {
			for route := range routes {
				matcherWithRoutesAndAdditionalPorts[matcher][Route{
					Value:     route.Value,
					MatchType: route.MatchType,
				}] = true
			}
		} else {
			matcherWithRoutesAndAdditionalPorts[matcher] = maps.Clone(routes)
		}
		if matcher.Port == 0 {
			for port := range portProtocols {
				if port == 0 {
					continue
				}
				additionalMatcher := Matcher{
					Protocol:  matcher.Protocol,
					Port:      port,
					MatchType: matcher.MatchType,
					Value:     matcher.Value,
				}
				key := newChainKey(port, matcher.Protocol, matcher.MatchType, matcher.Value)
				if owner, found := chainOwners[key]; found {
					if owner.protocol != matcher.Protocol {
						warnings.add(fmt.Sprintf(
							"matches with protocol %s and no port are not applied to %s on port %d, protocol %s is already configured there",
							matcher.Protocol, key.describe(), port, owner.protocol,
						))
						continue
					}
					if owner.matcher != additionalMatcher {
						warnings.add(fmt.Sprintf(
							"matches with protocol %s and no port are not applied to %s on port %d, match %q already defines a filter chain there",
							matcher.Protocol, key.describe(), port, owner.value,
						))
						continue
					}
				}
				if _, found := matcherWithRoutesAndAdditionalPorts[additionalMatcher]; found {
					for route := range routes {
						matcherWithRoutesAndAdditionalPorts[additionalMatcher][Route{
							Value:     route.Value,
							MatchType: route.MatchType,
						}] = true
					}
				} else {
					matcherWithRoutesAndAdditionalPorts[additionalMatcher] = maps.Clone(routes)
				}
			}
		}
	}
	filterChainMatchers := []FilterChainMatch{}
	for matcher, routes := range matcherWithRoutesAndAdditionalPorts {
		filterChainMatchers = append(filterChainMatchers,
			FilterChainMatch{
				Protocol:  matcher.Protocol,
				Port:      matcher.Port,
				MatchType: matcher.MatchType,
				Value:     matcher.Value,
				Routes:    getOrderedRoutes(routes),
			})
	}
	orderMatchers(filterChainMatchers)
	return filterChainMatchers, warnings.sorted()
}

// chainClass groups protocols whose filter chains can collide: chains of different
// classes always differ in the transport or application protocols they match on.
type chainClass int

const (
	tcpChain  chainClass = iota // tcp and mysql: raw_buffer with optional address and port
	tlsChain                    // tls: tls transport with SNI or address and optional port
	httpChain                   // http, http2 and grpc: raw_buffer and http/1.1,h2c with optional address and port
)

func protocolClass(protocol core_meta.Protocol) chainClass {
	switch protocol {
	case core_meta.ProtocolTLS:
		return tlsChain
	case core_meta.ProtocolHTTP, core_meta.ProtocolHTTP2, core_meta.ProtocolGRPC:
		return httpChain
	default:
		return tcpChain
	}
}

// chainKey identifies the filter chain a match ends up in: all domains of an L7
// protocol and port share one chain, TLS domains get a chain per SNI, IPs and CIDRs
// a chain per destination prefix range.
type chainKey struct {
	class   chainClass
	port    uint32
	address string
	sni     string
}

// chainOwner tells a conflicting matcher apart from one that merges into the chain
type chainOwner struct {
	protocol core_meta.Protocol
	matcher  Matcher
	value    string
}

func newChainKey(port uint32, protocol core_meta.Protocol, matchType MatchType, value string) chainKey {
	key := chainKey{class: protocolClass(protocol), port: port}
	switch matchType {
	case IP, IPV6, CIDR, CIDRV6:
		key.address = normalizePrefix(matchType, value)
	case Domain, WildcardDomain:
		if key.class == tlsChain {
			key.sni = value
		}
	}
	return key
}

// normalizePrefix converts an address to the prefix range Envoy matches on, so an IP,
// a CIDR covering only that IP and a non-canonical spelling all resolve to the same chain
func normalizePrefix(matchType MatchType, value string) string {
	switch matchType {
	case IP:
		return canonicalIP(value) + "/32"
	case IPV6:
		return canonicalIP(value) + "/128"
	case CIDR, CIDRV6:
		ip, prefixLen := getIpAndMask(value)
		return fmt.Sprintf("%s/%d", ip, prefixLen)
	}
	return ""
}

func canonicalIP(value string) string {
	if ip := net.ParseIP(value); ip != nil {
		return ip.String()
	}
	return value
}

func (k chainKey) conflictReason() string {
	if k.class == httpChain {
		return fmt.Sprintf("only one of %v can be configured on the same port", l7Protocols)
	}
	return "both would produce the same filter chain matcher"
}

func (k chainKey) describe() string {
	if k.sni != "" {
		return fmt.Sprintf("domain %q", k.sni)
	}
	if k.address == "" {
		return "domains"
	}
	return k.address
}

// warnings deduplicates messages, map iteration surfaces the same conflict multiple
// times and in random order
type warnings struct {
	messages map[string]struct{}
}

func newWarnings() *warnings {
	return &warnings{messages: map[string]struct{}{}}
}

func (w *warnings) add(message string) {
	w.messages[message] = struct{}{}
}

func (w *warnings) sorted() []string {
	messages := slices.Collect(maps.Keys(w.messages))
	sort.Strings(messages)
	return messages
}

func describePort(port uint32) string {
	if port == 0 {
		return "all ports"
	}
	return fmt.Sprintf("port %d", port)
}

func getMatchType(match api.Match, protocol core_meta.Protocol) (MatchType, bool) {
	var matchType MatchType
	isWildcardDomain := false
	switch match.Type {
	case api.MatchType("Domain"):
		matchType = Domain
		// for L7 protocol we want to aggregate routes
		if strings.HasPrefix(match.Value, "*") && slices.Contains([]core_meta.Protocol{core_meta.ProtocolGRPC, core_meta.ProtocolHTTP, core_meta.ProtocolHTTP2}, protocol) {
			matchType = Domain
			isWildcardDomain = true
		} else if strings.HasPrefix(match.Value, "*") {
			matchType = WildcardDomain
		}
	case api.MatchType("IP"):
		if govalidator.IsIPv6(match.Value) {
			matchType = IPV6
		} else {
			matchType = IP
		}
	case api.MatchType("CIDR"):
		split := strings.Split(match.Value, "/")
		if govalidator.IsIPv6(split[0]) {
			matchType = CIDRV6
		} else {
			matchType = CIDR
		}
	}
	return matchType, isWildcardDomain
}

func getOrderedRoutes(routesMap map[Route]bool) []Route {
	routes := []Route{}
	for route := range routesMap {
		routes = append(routes, route)
	}
	sort.SliceStable(routes, func(i, j int) bool {
		if routes[i].MatchType != routes[j].MatchType {
			return routes[i].MatchType > routes[j].MatchType
		}
		if routes[i].MatchType == Domain || routes[i].MatchType == WildcardDomain {
			return sortDomains(routes[i].Value, routes[j].Value)
		}

		return routes[i].MatchType < routes[j].MatchType
	})
	return routes
}

func orderMatchers(matchers []FilterChainMatch) {
	sort.SliceStable(matchers, func(i, j int) bool {
		if protocolOrder[matchers[i].Protocol] != protocolOrder[matchers[j].Protocol] {
			return protocolOrder[matchers[i].Protocol] < protocolOrder[matchers[j].Protocol]
		}
		if matchers[i].MatchType != matchers[j].MatchType {
			return matchers[i].MatchType > matchers[j].MatchType
		}
		if matchers[i].Port != matchers[j].Port {
			return matchers[i].Port > matchers[j].Port
		}
		if matchers[i].MatchType == Domain || matchers[i].MatchType == WildcardDomain {
			return sortDomains(matchers[i].Value, matchers[j].Value)
		}
		if matchers[i].MatchType == CIDR || matchers[i].MatchType == CIDRV6 {
			ipI, prefixI := getIpAndMask(matchers[i].Value)
			ipJ, prefixJ := getIpAndMask(matchers[j].Value)
			if prefixI == prefixJ {
				return ipI > ipJ
			}
			return prefixI > prefixJ
		}

		if matchers[i].MatchType == IP || matchers[i].MatchType == IPV6 {
			return matchers[i].Value > matchers[j].Value
		}

		return len(matchers[i].Routes) > len(matchers[j].Routes)
	})
}

func sortDomains(i string, j string) bool {
	splitI := strings.Split(i, ".")
	splitJ := strings.Split(j, ".")

	lenI := len(splitI)
	lenJ := len(splitJ)

	if lenI != lenJ {
		return lenI > lenJ
	}

	return i < j
}

func getIpAndMask(cidr string) (string, uint32) {
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return "", 0
	}
	ip := ipNet.IP.String()
	mask, _ := ipNet.Mask.Size()
	return ip, uint32(mask)
}
