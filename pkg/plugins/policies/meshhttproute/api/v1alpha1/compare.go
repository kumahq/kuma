package v1alpha1

import (
	"cmp"
	"slices"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
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
	case a.Type == b.Type:
		switch a.Type {
		case Exact:
			return 0
		case PathPrefix, RegularExpression:
			// Note this is intentionally "flipped" because a longer prefix means a
			// lesser match
			return cmp.Compare(len(b.Value), len(a.Value))
		}
	case a.Type == Exact:
		return -1
	case b.Type == Exact:
		return 1
	case a.Type == PathPrefix:
		return -1
	case b.Type == PathPrefix:
		return 1
	case a.Type == RegularExpression:
		return -1
	case b.Type == RegularExpression:
		return 1
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
	// Note this is intentionally "flipped" because more header matches
	// means a lesser match
	return cmp.Compare(len(b), len(a))
}

func compareQueryParams(a []QueryParamsMatch, b []QueryParamsMatch) int {
	// Note this is intentionally "flipped" because more query params matches
	// means a lesser match
	return cmp.Compare(len(b), len(a))
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
	BackendRefs []core_model.ResolvedBackendRef
	Hash        common_api.MatchesHash
}

// SortRules orders the rules according to Gateway API precedence:
// https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1.HTTPRouteRule
// We treat RegularExpression matches, which are implementation-specific, the
// same as prefix matches, the longer length match has priority.
func SortRules(
	rules []Rule,
	backendRefToOrigin map[common_api.MatchesHash]core_model.ResourceMeta,
	labelResolver core_model.LabelResourceIdentifierResolver,
) []Route {
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
		var backendRefs []core_model.ResolvedBackendRef
		hashMatches := HashMatches(key.rule.Matches)
		for _, br := range pointer.Deref(key.rule.Default.BackendRefs) {
			if origin, ok := backendRefToOrigin[hashMatches]; ok {
				if resolved := core_model.ResolveBackendRef(origin, br, labelResolver); resolved != nil {
					backendRefs = append(backendRefs, *resolved)
				}
			}
		}
		out = append(out, Route{
			Hash:        hashMatches,
			Match:       key.sortKey,
			BackendRefs: backendRefs,
			Filters:     pointer.Deref(key.rule.Default.Filters),
		})
	}
	return out
}
