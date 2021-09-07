package merge

import (
	"sort"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
)

type oldestFirst []*core_mesh.TrafficRouteResource

func (x oldestFirst) Less(i, j int) bool {
	return x[i].GetMeta().GetCreationTime().Before(
		x[j].GetMeta().GetCreationTime(),
	)
}

func (x oldestFirst) Len() int {
	return len(x)
}

func (x oldestFirst) Swap(i, j int) {
	x[i], x[j] = x[j], x[i]
}

// compareStringMatchers compares two StringMatcher objects in a manner
// similar to strings.Compare. A return value of -1 means lhs is less than
// rhs, 0 means they are equal, and 1 means that lhs is greater than rhs.
//
// The order we want is from most to least specific, so we return -1 if
// we consider lhs to be a longer (i.e. more specific) match.
func compareStringMatchers(
	lhs *mesh_proto.TrafficRoute_Http_Match_StringMatcher,
	rhs *mesh_proto.TrafficRoute_Http_Match_StringMatcher,
) int {
	switch {
	case lhs == nil && rhs == nil:
		return 0
	case lhs == nil:
		// Nil is less than anything else.
		return -1
	case rhs == nil:
		// Anything else is less than nil.
		return 1
	}

	// Compare strings by length (longest is lesser).
	longest := func(lhs string, rhs string) int {
		if lhs == rhs {
			return 0
		}

		if len(lhs) > len(rhs) {
			return -1
		}

		return 1
	}

	// Compare in order of specificity - exact, prefix, regex.
	switch lhs := lhs.MatcherType.(type) {
	case *mesh_proto.TrafficRoute_Http_Match_StringMatcher_Exact:
		switch rhs := rhs.MatcherType.(type) {
		case *mesh_proto.TrafficRoute_Http_Match_StringMatcher_Exact:
			return longest(lhs.Exact, rhs.Exact)
		case *mesh_proto.TrafficRoute_Http_Match_StringMatcher_Prefix:
			return -1
		case *mesh_proto.TrafficRoute_Http_Match_StringMatcher_Regex:
			return -1
		}
	case *mesh_proto.TrafficRoute_Http_Match_StringMatcher_Prefix:
		switch rhs := rhs.MatcherType.(type) {
		case *mesh_proto.TrafficRoute_Http_Match_StringMatcher_Exact:
			return 1
		case *mesh_proto.TrafficRoute_Http_Match_StringMatcher_Prefix:
			return longest(lhs.Prefix, rhs.Prefix)
		case *mesh_proto.TrafficRoute_Http_Match_StringMatcher_Regex:
			return -1
		}
	case *mesh_proto.TrafficRoute_Http_Match_StringMatcher_Regex:
		switch rhs := rhs.MatcherType.(type) {
		case *mesh_proto.TrafficRoute_Http_Match_StringMatcher_Exact:
			return 1
		case *mesh_proto.TrafficRoute_Http_Match_StringMatcher_Prefix:
			return 1
		case *mesh_proto.TrafficRoute_Http_Match_StringMatcher_Regex:
			// String compare on a regex is pretty arbitrary. Compare by
			// length (as a proxy for complexity).
			return longest(lhs.Regex, rhs.Regex)
		}
	}

	return 0 // Probably equal.
}

type longestMatchFirst []*mesh_proto.TrafficRoute_Http

func (x longestMatchFirst) Less(i, j int) bool {
	lhs := x[i].GetMatch()
	rhs := x[j].GetMatch()

	// Empty match (i.e. matches '/') should always be last.
	switch {
	case lhs == nil:
		return false
	case rhs == nil:
		return true
	}

	switch compareStringMatchers(lhs.GetPath(), rhs.GetPath()) {
	case -1:
		return true
	case 1:
		return false
	}

	// Path compared equal, so we check the method.
	switch compareStringMatchers(lhs.GetMethod(), rhs.GetMethod()) {
	case -1:
		return true
	case 1:
		return false
	}

	lhsHeaders := lhs.GetHeaders()
	rhsHeaders := rhs.GetHeaders()

	keysOf := func(h map[string]*mesh_proto.TrafficRoute_Http_Match_StringMatcher) []string {
		var keys []string
		for k := range h {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		return keys
	}

	switch {
	case len(lhsHeaders) > len(rhsHeaders):
		return true
	case len(lhsHeaders) < len(rhsHeaders):
		return false
	default:
		// Now sort the headers and compare pairwise.
		lhsKeys := keysOf(lhsHeaders)
		rhsKeys := keysOf(rhsHeaders)

		for i, k := range lhsKeys {
			if k == rhsKeys[i] {
				continue
			}

			return k < rhsKeys[i]
		}
	}

	// Probably equal, or at any rate, not less than.
	return false
}

func (x longestMatchFirst) Len() int {
	return len(x)
}

func (x longestMatchFirst) Swap(i, j int) {
	x[i], x[j] = x[j], x[i]
}

// TrafficRoute merges a slice of TrafficRoute policies. Singular fields
// will preserve the value of the oldest route.
func TrafficRoute(routes ...*core_mesh.TrafficRouteResource) *core_mesh.TrafficRouteResource {
	if len(routes) == 0 {
		return nil
	}

	// Sort oldest first so that we take the default route from the
	// oldest resource.
	sort.Stable(oldestFirst(routes))

	out := core_mesh.NewTrafficRouteResource()
	out.Spec.Conf = &mesh_proto.TrafficRoute_Conf{}

	for _, r := range routes {
		// If we haven't set the default route yet, attempt to
		// take it from the next route.
		if out.Spec.GetConf().GetSplit() == nil &&
			out.Spec.GetConf().GetDestination() == nil {
			out.Spec.GetConf().Split = r.Spec.GetConf().GetSplit()
			out.Spec.GetConf().Destination = r.Spec.GetConf().GetDestination()
		}

		out.Spec.GetConf().Http = append(out.Spec.GetConf().Http, r.Spec.GetConf().GetHttp()...)
	}

	sort.Stable(longestMatchFirst(out.Spec.GetConf().GetHttp()))

	return out
}
