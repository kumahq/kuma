package xds

import (
	"net"
	"sort"
	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/pkg/errors"
	"go.uber.org/multierr"

	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshpassthrough/api/v1alpha1"
)

type MatchType int

const (
	Domain MatchType = iota + 1
	WildcardDomain
	IP
	IPV6
	CIDR
	CIDRV6
)

type Route struct {
	Value     string
	MatchType MatchType
}

type Matcher struct {
	Protocol core_mesh.Protocol
	Port     uint32
}

type FilterChainMatcher struct {
	Protocol core_mesh.Protocol
	Port     uint32
	Routes   []Route
}

func GetOrderedMatchers(conf api.Conf) ([]FilterChainMatcher, error) {
	matchers := map[Matcher]map[Route]bool{}
	portProtocols := map[int]map[core_mesh.Protocol]bool{}
	for _, match := range conf.AppendMatch {
		var matchType MatchType
		switch match.Type {
		case api.MatchType("Domain"):
			if strings.HasPrefix(match.Value, "*") {
				matchType = WildcardDomain
			} else {
				matchType = Domain
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
		port := 0
		if match.Port != nil {
			port = *match.Port
		}
		protocol := core_mesh.ParseProtocol(string(match.Protocol))
		if _, found := portProtocols[port]; !found {
			portProtocols[port] = map[core_mesh.Protocol]bool{protocol: true}
		} else {
			portProtocols[port][protocol] = true
		}
		matcher := Matcher{
			Protocol: protocol,
			Port:     uint32(port),
		}
		route := Route{
			Value:     match.Value,
			MatchType: matchType,
		}
		if _, found := matchers[matcher]; found {
			matchers[matcher][route] = true
		} else {
			matchers[matcher] = map[Route]bool{
				route: true,
			}
		}
	}
	filterChainMatchers := []FilterChainMatcher{}
	for matcher, routesMap := range matchers {
		routes := []Route{}
		for route := range routesMap {
			routes = append(routes, route)
		}
		orderRoutes(routes)
		filterChainMatcher := FilterChainMatcher{
			Protocol: matcher.Protocol,
			Port:     matcher.Port,
			Routes:   routes,
		}
		filterChainMatchers = append(filterChainMatchers, filterChainMatcher)
	}
	if err := validatePortAndProtocol(portProtocols); err != nil {
		return nil, err
	}
	orderValues(filterChainMatchers)
	return filterChainMatchers, nil
}

func validatePortAndProtocol(portProtocols map[int]map[core_mesh.Protocol]bool) error {
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

func orderRoutes(routes []Route) {
	sort.SliceStable(routes, func(i, j int) bool {
		if routes[i].MatchType != routes[j].MatchType {
			return routes[i].MatchType < routes[j].MatchType
		}
		if routes[i].MatchType == Domain || routes[i].MatchType == WildcardDomain {
			return sortDomains(routes[i].Value, routes[j].Value)
		}
		if routes[i].MatchType == CIDR || routes[i].MatchType == CIDRV6 {
			_, prefixI := getIpAndMask(routes[i].Value)
			_, prefixJ := getIpAndMask(routes[j].Value)
			return prefixI > prefixJ
		}

		return routes[i].MatchType < routes[j].MatchType
	})
}

func orderValues(matchers []FilterChainMatcher) {
	sort.SliceStable(matchers, func(i, j int) bool {
		if matchers[i].Protocol != matchers[j].Protocol {
			return matchers[i].Protocol > matchers[j].Protocol
		}
		return matchers[i].Port > matchers[j].Port
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
