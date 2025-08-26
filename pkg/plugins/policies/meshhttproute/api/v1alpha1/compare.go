package v1alpha1

import (
	"cmp"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/rules/resolve"
	"github.com/kumahq/kuma/pkg/util/pointer"
)

func comparePath(a, b *PathMatch) int {
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

func compareMethod(a, b *Method) int {
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

func compareHeaders(a, b []common_api.HeaderMatch) int {
	// Note this is intentionally "flipped" because more header matches
	// means a lesser match
	return cmp.Compare(len(b), len(a))
}

func compareQueryParams(a, b []QueryParamsMatch) int {
	// Note this is intentionally "flipped" because more query params matches
	// means a lesser match
	return cmp.Compare(len(b), len(a))
}

// CompareMatch orders the rules according to Gateway API precedence:
// https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1.HTTPRouteRule
// We treat RegularExpression matches, which are implementation-specific, the
// same as prefix matches, the longer length match has priority.
func CompareMatch(a, b Match) int {
	if p := comparePath(a.Path, b.Path); p != 0 {
		return p
	}

	if p := compareMethod(a.Method, b.Method); p != 0 {
		return p
	}

	if p := compareHeaders(pointer.Deref(a.Headers), pointer.Deref(b.Headers)); p != 0 {
		return p
	}

	if p := compareQueryParams(pointer.Deref(a.QueryParams), pointer.Deref(b.QueryParams)); p != 0 {
		return p
	}

	return 0
}

type Route struct {
	Name        string
	Match       Match
	Filters     []Filter
	BackendRefs []resolve.ResolvedBackendRef
}
