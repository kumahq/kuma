package xds

import (
	"net"
	"slices"
	"sort"
	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/pkg/errors"
	"go.uber.org/multierr"

	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshpassthrough/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/util/pointer"
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

var protocolOrder = map[core_mesh.Protocol]int{
	core_mesh.ProtocolTLS:   0,
	core_mesh.ProtocolTCP:   1,
	core_mesh.ProtocolHTTP:  2,
	core_mesh.ProtocolHTTP2: 3,
	core_mesh.ProtocolGRPC:  4,
}

type Route struct {
	Value     string
	MatchType MatchType
}

type Matcher struct {
	Protocol  core_mesh.Protocol
	Port      uint32
	MatchType MatchType
	Value     string
}

type FilterChainMatch struct {
	Protocol  core_mesh.Protocol
	Port      uint32
	MatchType MatchType
	Value     string
	Routes    []Route
}

func GetOrderedMatchers(conf api.Conf) ([]FilterChainMatch, error) {
	matcherWithRoutes := map[Matcher]map[Route]bool{}
	portProtocols := map[uint32]map[core_mesh.Protocol]bool{}
	for _, match := range conf.AppendMatch {
		port := pointer.DerefOr[uint32](match.Port, 0)
		protocol := core_mesh.ParseProtocol(string(match.Protocol))
		matchType, isWildcardDomain := getMatchType(match, protocol)
		matcher := Matcher{
			Protocol:  protocol,
			Port:      port,
			MatchType: matchType,
		}
		if _, found := portProtocols[port]; !found {
			portProtocols[port] = map[core_mesh.Protocol]bool{protocol: true}
		} else {
			portProtocols[port][protocol] = true
		}
		switch protocol {
		case core_mesh.ProtocolHTTP, core_mesh.ProtocolHTTP2, core_mesh.ProtocolGRPC:
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
	// we cannot differentiate between HTTP, HTTP/2, and gRPC on the same port.
	if err := validatePortAndProtocol(portProtocols); err != nil {
		return nil, err
	}
	// Envoy first checks the port when performing matching. If there is a matcher for a specific port
	// and one rule to match all ports alongside another for a specific port,
	// it might select the matcher for the specific port but fail to find a corresponding filter chain.
	// To avoid this issue, we also generate specific port matchers for rules intended to match all ports.
	matcherWithRoutesAndAdditionalPorts := map[Matcher]map[Route]bool{}
	for matcher, routes := range matcherWithRoutes {
		matcherWithRoutesAndAdditionalPorts[matcher] = routes
		if matcher.Port == 0 {
			for port := range portProtocols {
				additionalMatcher := Matcher{
					Protocol:  matcher.Protocol,
					Port:      port,
					MatchType: matcher.MatchType,
					Value:     matcher.Value,
				}
				if routes, found := matcherWithRoutesAndAdditionalPorts[additionalMatcher]; found {
					for route := range routes {
						matcherWithRoutesAndAdditionalPorts[additionalMatcher][route] = true
					}
				} else {
					matcherWithRoutesAndAdditionalPorts[additionalMatcher] = routes
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
	return filterChainMatchers, nil
}

func getMatchType(match api.Match, protocol core_mesh.Protocol) (MatchType, bool) {
	var matchType MatchType
	isWildcardDomain := false
	switch match.Type {
	case api.MatchType("Domain"):
		matchType = Domain
		// for L7 protocol we want to aggregate routes
		if strings.HasPrefix(match.Value, "*") && slices.Contains([]core_mesh.Protocol{core_mesh.ProtocolGRPC, core_mesh.ProtocolHTTP, core_mesh.ProtocolHTTP2}, protocol) {
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

func validatePortAndProtocol(portProtocols map[uint32]map[core_mesh.Protocol]bool) error {
	var errs error
	for port, protocols := range portProtocols {
		var counter int
		if _, found := protocols[core_mesh.ProtocolHTTP]; found {
			counter++
		}
		if _, found := protocols[core_mesh.ProtocolHTTP2]; found {
			counter++
		}
		if _, found := protocols[core_mesh.ProtocolGRPC]; found {
			counter++
		}
		if counter > 1 {
			errs = multierr.Append(errs, errors.Errorf("you cannot configure http, http2, grpc on the same port %d", port))
		}
	}
	return errs
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
			_, prefixI := getIpAndMask(matchers[i].Value)
			_, prefixJ := getIpAndMask(matchers[j].Value)
			return prefixI > prefixJ
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
