package subsetutils

import (
	"maps"
	"sort"
	"strings"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	util_maps "github.com/kumahq/kuma/v3/pkg/util/maps"
)

// Tag is a key-value pair. If Not is true then Key != Value
type Tag struct {
	Key   string
	Value string
	Not   bool
}

// Subset represents a group of proxies
type Subset []Tag

func NewSubset(m map[string]string) Subset {
	var s Subset
	for _, k := range util_maps.SortedKeys(m) {
		s = append(s, Tag{Key: k, Value: m[k]})
	}
	return s
}

// IsSubset returns true if 'other' is a subset of the current set.
// Empty set is a superset for all subsets.
func (ss Subset) IsSubset(other Subset) bool {
	if len(ss) == 0 {
		return true
	}
	otherByKeys := map[string][]Tag{}
	for _, t := range other {
		otherByKeys[t.Key] = append(otherByKeys[t.Key], t)
	}
	for _, tag := range ss {
		oTags, ok := otherByKeys[tag.Key]
		if !ok {
			return false
		}
		for _, otherTag := range oTags {
			if !isSubset(tag, otherTag) {
				return false
			}
		}
	}
	return true
}

func isSubset(t1, t2 Tag) bool {
	switch {
	// t2={y: b} can't be a subset of t1={x: a} because point {y: b, x: c} belongs to t2, but doesn't belong to t1
	case t1.Key != t2.Key:
		return false

	// t2={y: !a} is a subset of t1={y: !b} if and only if a == b
	case t1.Not == t2.Not:
		return t1.Value == t2.Value

	// t2={y: a} is a subset of t1={y: !b} if and only if a != b
	case t1.Not:
		return t1.Value != t2.Value

	// t2={y: !a} can't be a subset of t1={y: b} because point {y: c} belongs to t2, but doesn't belong to t1
	case t2.Not:
		return false

	default:
		panic("impossible")
	}
}

func (ss Subset) WithTag(key, value string, not bool) Subset {
	return append(ss, Tag{Key: key, Value: value, Not: not})
}

// MatchAll returns an empty Subset, meaning it matches every element.
func MatchAll() Subset {
	return Subset{}
}

func MeshService(name string) Subset {
	return Subset{{
		Key: mesh_proto.ServiceTag, Value: name,
	}}
}

func MeshExternalService(name string) Subset {
	return Subset{{
		Key: mesh_proto.ServiceTag, Value: name,
	}}
}

func SubsetFromTags(tags map[string]string) Subset {
	subset := Subset{}
	for k, v := range tags {
		subset = append(subset, Tag{Key: k, Value: v})
	}
	return subset
}

type Element map[string]string

func (e Element) WithKeyValue(key, value string) Element {
	c := maps.Clone(e)
	if c == nil {
		c = Element{}
	}

	c[key] = value
	return c
}

func MeshElement() Element {
	return Element{}
}

// NumPositive returns a number of tags without negation
func (ss Subset) NumPositive() int {
	pos := 0
	for _, t := range ss {
		if !t.Not {
			pos++
		}
	}
	return pos
}

func (ss Subset) Sorted() {
	sort.SliceStable(ss, func(i, j int) bool {
		if ss[i].Key != ss[j].Key {
			return ss[i].Key < ss[j].Key
		}
		if ss[i].Value != ss[j].Value {
			return ss[i].Value < ss[j].Value
		}
		return !ss[i].Not && ss[j].Not
	})
}

func (ss Subset) IndexOfPositive() int {
	for i, t := range ss {
		if !t.Not {
			return i
		}
	}
	return -1
}

// Deduplicate returns a new slice of subsetutils.Subset with duplicates removed.
func Deduplicate(subsets []Subset) []Subset {
	seen := make(map[string]struct{})
	result := make([]Subset, 0, len(subsets))

	for _, s := range subsets {
		key := canonicalSubset(s)
		if _, exists := seen[key]; !exists {
			seen[key] = struct{}{}
			result = append(result, s)
		}
	}
	return result
}

// canonicalSubset returns a canonical string representation for a subset.
// It assumes that a subset is a slice of subsetutils.Tag with fields Key, Value, and Not.
func canonicalSubset(s Subset) string {
	if len(s) == 0 {
		return ""
	}
	s.Sorted()
	var sb strings.Builder
	for i, t := range s {
		if i > 0 {
			sb.WriteByte('|') // Separator
		}
		sb.WriteString(t.Key)
		sb.WriteByte(':')
		sb.WriteString(t.Value)
		sb.WriteByte(':')
		if t.Not {
			sb.WriteByte('1')
		} else {
			sb.WriteByte('0')
		}
	}
	return sb.String()
}
