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
	// the L7 protocol configured for a filter chain, the first match wins
	l7ProtocolOnChain := map[l7ChainKey]core_meta.Protocol{}
	for _, match := range pointer.Deref(conf.AppendMatch) {
		port := pointer.DerefOr[uint32](match.Port, 0)
		protocol := core_meta.ParseProtocol(string(match.Protocol))
		matchType, isWildcardDomain := getMatchType(match, protocol)
		if slices.Contains(l7Protocols, protocol) {
			key := newL7ChainKey(port, matchType, match.Value)
			if configured, found := l7ProtocolOnChain[key]; found && configured != protocol {
				warnings.add(fmt.Sprintf(
					"ignoring match %q with protocol %s, protocol %s is already configured for %s on %s, only one of %v can be configured on the same port",
					match.Value, protocol, configured, key.describe(), describePort(port), l7Protocols,
				))
				continue
			}
			l7ProtocolOnChain[key] = protocol
		}
		matcher := Matcher{
			Protocol:  protocol,
			Port:      port,
			MatchType: matchType,
		}
		if _, found := portProtocols[port]; !found {
			portProtocols[port] = map[core_meta.Protocol]bool{protocol: true}
		} else {
			portProtocols[port][protocol] = true
		}
		switch protocol {
		case core_meta.ProtocolHTTP, core_meta.ProtocolHTTP2, core_meta.ProtocolGRPC:
			// when there are domains we want to create VirtualHosts with Domain match
			if matchType == Domain {
				if isWildcardDomain {
					matchType = WildcardDomain
				}
				route := Route{
					Value:     match.Value,
					MatchType: matchType,
				}
				if _, found := matcherWithRoutes[matcher]; found {
					matcherWithRoutes[matcher][route] = true
				} else {
					matcherWithRoutes[matcher] = map[Route]bool{
						route: true,
					}
				}
			} else {
				matcher.Value = match.Value
				// there should be no existing matcher if there is ip/cidr
				matcherWithRoutes[matcher] = map[Route]bool{}
			}
		default:
			matcher.Value = match.Value
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
				// a match without a port must not override the protocol explicitly configured
				// for this port, otherwise we generate two filter chains with the same matcher
				key := newL7ChainKey(port, matcher.MatchType, matcher.Value)
				if configured, found := l7ProtocolOnChain[key]; found && configured != matcher.Protocol && slices.Contains(l7Protocols, matcher.Protocol) {
					warnings.add(fmt.Sprintf(
						"matches with protocol %s and no port are not applied to %s on port %d, protocol %s is already configured there",
						matcher.Protocol, key.describe(), port, configured,
					))
					continue
				}
				additionalMatcher := Matcher{
					Protocol:  matcher.Protocol,
					Port:      port,
					MatchType: matcher.MatchType,
					Value:     matcher.Value,
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

// l7ChainKey identifies the filter chain an L7 match ends up in. All domains of a given
// protocol and port share a single filter chain without an address matcher, IPs and CIDRs
// get their own filter chain matched on the destination address.
type l7ChainKey struct {
	port    uint32
	address string
}

func newL7ChainKey(port uint32, matchType MatchType, value string) l7ChainKey {
	key := l7ChainKey{port: port}
	// the address is normalized to the prefix range Envoy matches on, an IP and a CIDR
	// covering only that IP end up in the same filter chain
	switch matchType {
	case IP:
		key.address = fmt.Sprintf("%s/32", value)
	case IPV6:
		key.address = fmt.Sprintf("%s/128", value)
	case CIDR, CIDRV6:
		ip, prefixLen := getIpAndMask(value)
		key.address = fmt.Sprintf("%s/%d", ip, prefixLen)
	}
	return key
}

func (k l7ChainKey) describe() string {
	if k.address == "" {
		return "domains"
	}
	return k.address
}

// warnings deduplicates messages and returns them in a stable order, matchers are kept
// in maps so the same conflict can be reported multiple times and in a random order
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
