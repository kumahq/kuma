package xds

import (
	"sort"
	"strconv"
	"strings"

	"github.com/asaskevich/govalidator"

	api "github.com/kumahq/kuma/pkg/plugins/policies/meshpassthrough/api/v1alpha1"
)

type MatchType int

const (
	Domain MatchType = iota + 1
	WildcardDomain
	IP
	CIDR
	IPV6
	CIDRV6
)

type (
	MatchersPerType map[MatchType][]api.Match
	MatchersPerPort map[int]MatchersPerType
)

func GetOrderedMatchers(conf api.Conf) (MatchersPerPort, MatchersPerPort) {
	// validate port and protocol conflict 
	rawBuffer := MatchersPerPort{}
	tls := MatchersPerPort{}
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

		switch match.Protocol {
		case "tls":
			port := 0
			if match.Port != nil {
				port = *match.Port
			}
			if _, ok := tls[port]; !ok {
				tls[port] = MatchersPerType{}
			}
			if _, ok := tls[port][matchType]; !ok {
				tls[port][matchType] = []api.Match{}
			}
			tls[port][matchType] = append(tls[port][matchType], match)
		default:
			port := 0
			if match.Port != nil {
				port = *match.Port
			}
			if _, ok := rawBuffer[port]; !ok {
				rawBuffer[port] = MatchersPerType{}
			}
			if _, ok := rawBuffer[port][matchType]; !ok {
				rawBuffer[port][matchType] = []api.Match{}
			}
			rawBuffer[port][matchType] = append(rawBuffer[port][matchType], match)
		}
	}
	orderValues(tls)
	orderValues(rawBuffer)
	return tls, rawBuffer
}

func orderValues(matchersPerPort MatchersPerPort) {
	for _, matchers := range matchersPerPort {
		for matchType, values := range matchers {
			switch matchType {
			case Domain:
				sort.SliceStable(values, func(i, j int) bool {
					return sortDomains(values[i].Value, values[j].Value)
				})
			case WildcardDomain:
				sort.SliceStable(values, func(i, j int) bool {
					return sortDomains(values[i].Value, values[j].Value)
				})
			case CIDR, CIDRV6:
				sort.SliceStable(values, func(i, j int) bool {
					return sortCIDR(values[i].Value, values[j].Value)
				})
			}
		}
	}
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

func sortCIDR(i string, j string) bool {
	// Compare CIDR prefix lengths
	lenI := GetCIDRPrefixLength(i)
	lenJ := GetCIDRPrefixLength(j)

	return lenI > lenJ
}

func GetCIDRPrefixLength(cidr string) uint32 {
	// Split CIDR into address and prefix
	parts := strings.Split(cidr, "/")
	if len(parts) != 2 {
		return 0
	}

	// Convert prefix length to integer
	prefixLength, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0
	}
	return uint32(prefixLength)
}
