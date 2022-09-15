package route

import "sort"

type Sorter []Entry

var _ sort.Interface = &Sorter{}

func (s Sorter) Len() int {
	return len(s)
}

func (s Sorter) Less(i, j int) bool {
	return isMoreSpecific(&s[i].Match, &s[j].Match)
}

func (s Sorter) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func isMoreSpecific(lhs *Match, rhs *Match) bool {
	switch {
	case lhs.ExactPath != "":
		// Exact match is more specific that prefix or regex.
		if rhs.ExactPath == "" {
			return true
		}
		// Longer path is more specific.
		if len(lhs.ExactPath) > len(rhs.ExactPath) {
			return true
		}
		// Shorter path is not more specific.
		if len(lhs.ExactPath) < len(rhs.ExactPath) {
			return false
		}
	case rhs.ExactPath != "":
		return false
	case lhs.PrefixPath != "":
		// Prefix match is more specific than regex.
		if rhs.PrefixPath == "" {
			return true
		}
		// Longer prefix is more specific.
		if len(lhs.PrefixPath) > len(rhs.PrefixPath) {
			return true
		}
		// Shorter path is not more specific.
		if len(lhs.PrefixPath) < len(rhs.PrefixPath) {
			return false
		}
	case rhs.PrefixPath != "":
		return false
	default:
		// Regex match is more specific than no path match.
		if rhs.RegexPath == "" {
			return true
		}
		// Longer regex might be more specific.
		if len(lhs.RegexPath) > len(rhs.RegexPath) {
			return true
		}
		// Shorter path might not more specific.
		if len(lhs.RegexPath) < len(rhs.RegexPath) {
			return false
		}
	}

	if lhs.Method != "" && rhs.Method == "" {
		return true
	}

	if (len(lhs.ExactHeader) + len(lhs.ExactQuery)) >
		(len(rhs.ExactHeader) + len(rhs.ExactQuery)) {
		return true
	}

	if (len(lhs.RegexHeader) + len(lhs.RegexQuery)) >
		(len(rhs.RegexHeader) + len(rhs.RegexQuery)) {
		return true
	}

	// NOTE: this is a partial ordering, since we don't (yet?) order on
	// the contents of the non-path match criteria. This means there
	// is still a chance of churning Envoy config with changes that
	// differ only in route order.

	// Sorting on equal-specificity paths helps to mitigate YAML
	// element ordering errors in tests.
	return lhs.ExactPath+lhs.PrefixPath+lhs.RegexPath <
		rhs.ExactPath+rhs.PrefixPath+rhs.RegexPath
}
