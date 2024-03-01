package v1alpha1

import (
	"slices"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	"github.com/kumahq/kuma/pkg/util/pointer"
)

func comparePath(a *PathMatch, b *PathMatch) int {
	switch {
	case a != nil && b == nil:
		return -1
	case a == nil && b != nil:
		return 1
	case a == nil && b == nil:
		return 0
	}

	switch {
	case a.Type == Exact && b.Type == Exact:
	case a.Type == Exact && b.Type != Exact:
		return -1
	case a.Type != Exact && b.Type == Exact:
		return 1
	}

	switch {
	case a.Type == PathPrefix && b.Type != PathPrefix:
		return -1
	case a.Type != PathPrefix && b.Type == PathPrefix:
		return 1
	case a.Type == PathPrefix && b.Type == PathPrefix:
		switch {
		case len(a.Value) > len(b.Value):
			return -1
		case len(a.Value) < len(b.Value):
			return 1
		}
	}

	switch {
	case a.Type == RegularExpression && b.Type != RegularExpression:
		return -1
	case a.Type != RegularExpression && b.Type == RegularExpression:
		return 1
	case a.Type == RegularExpression && b.Type == RegularExpression:
		switch {
		case len(a.Value) > len(b.Value):
			return -1
		case len(a.Value) < len(b.Value):
			return 1
		}
	}

	return 0
}

func compareMethod(a *Method, b *Method) int {
	switch {
	case a != nil && b == nil:
		return -1
	case a == nil && b != nil:
		return 1
	case a == nil && b == nil:
		return 0
	}

	return 0
}

func compareHeaders(a []common_api.HeaderMatch, b []common_api.HeaderMatch) int {
	switch {
	case len(a) > len(b):
		return -1
	case len(b) > len(a):
		return 1
	}

	return 0
}

func compareQueryParams(a []QueryParamsMatch, b []QueryParamsMatch) int {
	switch {
	case len(a) > len(b):
		return -1
	case len(b) > len(a):
		return 1
	}

	return 0
}

func CompareMatch(a Match, b Match) int {
	if p := comparePath(a.Path, b.Path); p != 0 {
		return p
	}

	if p := compareMethod(a.Method, b.Method); p != 0 {
		return p
	}

	if p := compareHeaders(a.Headers, b.Headers); p != 0 {
		return p
	}

	if p := compareQueryParams(a.QueryParams, b.QueryParams); p != 0 {
		return p
	}

	return 0
}

type Route struct {
	Match       Match
	Filters     []Filter
	BackendRefs []common_api.BackendRef
	Hash        string
}

// SortRules orders the rules according to Gateway API precedence:
// https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1.HTTPRouteRule
func SortRules(rules []Rule) []Route {
	type keyed struct {
		sortKey Match
		rule    Rule
	}
	var keys []keyed
	for _, rule := range rules {
		for _, match := range rule.Matches {
			keys = append(keys, keyed{
				sortKey: match,
				rule:    rule,
			})
		}
	}
	slices.SortStableFunc(keys, func(i, j keyed) int {
		return CompareMatch(i.sortKey, j.sortKey)
	})
	var out []Route
	for _, key := range keys {
		out = append(out, Route{
			Hash:        HashMatches(key.rule.Matches),
			Match:       key.sortKey,
			BackendRefs: pointer.Deref(key.rule.Default.BackendRefs),
			Filters:     pointer.Deref(key.rule.Default.Filters),
		})
	}
	return out
}
