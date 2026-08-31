package xds

import (
	"maps"
	"net"
	"slices"
	"sort"
	"strconv"
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
	core_meta.ProtocolMysql: 1,
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

// GetOrderedMatchers builds the filter chain matchers for the configuration. Matches
// Envoy would reject the listener for are dropped, so a single incorrect match doesn't
// invalidate the whole passthrough listener.
func GetOrderedMatchers(conf api.Conf) []FilterChainMatch {
	conflicts := api.FindConflicts(conf)
	matcherWithRoutes := map[Matcher]map[Route]bool{}
	matcherFilterChains := map[Matcher]api.FilterChainMatcher{}
	portProtocols := map[uint32]map[core_meta.Protocol]bool{}
	for i, match := range pointer.Deref(conf.AppendMatch) {
		if conflicts.IsDropped(i) {
			continue
		}
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
		matcherFilterChains[matcher] = match.FilterChainMatcher()
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
				if conflicts.IsSuppressed(matcherFilterChains[matcher].WithPort(port)) {
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
	return filterChainMatchers
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
		// classify by the canonical form, so an IPv4-mapped IPv6 CIDR lands on
		// the IPv4 listener the canonical prefix is generated for
		split := strings.Split(api.CanonicalCIDR(match.Value), "/")
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

// getIpAndMask returns the canonical prefix Envoy matches on, so an IPv4-mapped
// IPv6 CIDR produces its plain IPv4 form instead of a prefix length Envoy rejects
func getIpAndMask(cidr string) (string, uint32) {
	canonical := api.CanonicalCIDR(cidr)
	if _, _, err := net.ParseCIDR(canonical); err != nil {
		return "", 0
	}
	ip, maskText, _ := strings.Cut(canonical, "/")
	mask, _ := strconv.ParseUint(maskText, 10, 32)
	return ip, uint32(mask)
}
