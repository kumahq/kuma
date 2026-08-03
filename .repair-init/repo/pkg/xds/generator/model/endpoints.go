package model

import (
	"cmp"
	"maps"
	"slices"
)

type Endpoint struct {
	Address string
	Port    uint32
}

type Endpoints []Endpoint

func EndpointsFromMap[T any](endpointsMap map[Endpoint]T) Endpoints {
	// sort for consistent envoy config
	return slices.SortedStableFunc(
		maps.Keys(endpointsMap),
		func(a Endpoint, b Endpoint) int {
			switch {
			case a.Address != b.Address:
				return cmp.Compare(a.Address, b.Address)
			default:
				return cmp.Compare(a.Port, b.Port)
			}
		},
	)
}
