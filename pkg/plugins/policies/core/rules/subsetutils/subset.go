package subsetutils

import (
	"maps"
	"sort"
	"strings"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	util_maps "github.com/kumahq/kuma/pkg/util/maps"
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

// ContainsElement returns true if there exists a key in 'other' that matches the current set,
// and the corresponding k-v pair must match the set rule. Also, the left set rules of the current set can't make an impact.
// Empty set is a superset for all elements.
//
// For example if you have a Subset with Tags: [{key: zone, value: east, not: true}, {key: service, value: frontend, not: false}]
// an Element with k-v pairs: 1) service: frontend  2) version: zone1
// there's a k-v pair 'service: frontend' in Element that matches the set rule {key: service, value: frontend, not: false}
// the left set rule of Subset {key: zone, value: east, not: true} won't make an impact because of 'not: true'
func (ss Subset) ContainsElement(other Element) bool {
	// 1. find the overlaps of element and current subset
	// 2. verify the overlaps
	// 3. verify the left of current subset
	// 4. if no overlaps, verify if all the Subset rules are negative

	if len(ss) == 0 {
		return true
	}
	if len(other) == 0 {
		return false
	}

	hasOverlapKey := false
	for _, tag := range ss {
		otherVal, ok := other[tag.Key]
		if ok {
			hasOverlapKey = true

			// contradict
			if tag.Value == otherVal && tag.Not {
				return false
			}
			// intersect
			if tag.Value == otherVal && !tag.Not {
				continue
			}
			// intersect
			if tag.Value != otherVal && tag.Not {
				continue
			}
			// contradict
			if tag.Value != otherVal && !tag.Not {
				return false
			}
		} else if !tag.Not {
			// For those items that don't exist in element should not make an impact.
			// For example, the DP with tag {"service: frontend"} doesn't match
			// the policy with matching tags [{"service: frontend"}, {"zone": "east"}]
			return false
		}
	}

	// if the current Subset owns all of negative rules and no overlapped keys in Element,
	// we can also regard the Subset contains Element
	if !hasOverlapKey && ss.NumPositive() == 0 {
		return true
	}

	return hasOverlapKey
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

// Intersect returns true if there exists an element that belongs both to 'other' and current set.
// Empty set intersects with all sets.
//
// We're using this function to check if 2 'from' rules of MeshTrafficPermission can be applied to the same client DPP.
// For example:
//
//	from:
//	  - targetRef:
//	      kind: MeshSubset
//	      tags:
//	        team: team-a
//	  - targetRef:
//	      kind: MeshSubset
//	      tags:
//	        zone: east
//
// there is a DPP with tags 'team: team-a' and 'zone: east' that's subjected to both these rules.
// So 'from[0]' and 'from[1]' have an intersection.
// However, in another example:
//
//	from:
//	 - targetRef:
//	     kind: MeshSubset
//	     tags:
//	       team: team-a
//	 - targetRef:
//	     kind: MeshSubset
//	     tags:
//	       team: team-b
//	       zone: east
//
// there is no DPP that'd hit both 'from[0]' and 'from[1]'. So in this case they don't have an intersection.
func (ss Subset) Intersect(other Subset) bool {
	if len(ss) == 0 || len(other) == 0 {
		return true
	}
	otherByKeysOnlyPositive := map[string][]Tag{}
	for _, t := range other {
		if t.Not {
			continue
		}
		otherByKeysOnlyPositive[t.Key] = append(otherByKeysOnlyPositive[t.Key], t)
	}
	for _, tag := range ss {
		if tag.Not {
			continue
		}
		oTags, ok := otherByKeysOnlyPositive[tag.Key]
		if !ok {
			continue
		}
		for _, otherTag := range oTags {
			if otherTag != tag {
				return false
			}
		}
	}
	return true
}

func (ss Subset) WithTag(key, value string, not bool) Subset {
	return append(ss, Tag{Key: key, Value: value, Not: not})
}

func MeshSubset() Subset {
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

func KumaServiceTagElement(name string) Element {
	return Element{mesh_proto.ServiceTag: name}
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

type SubsetIter struct {
	dims     []chooser
	idx      []int
	finished bool
}

type chooser struct {
	key  string
	vals []string // unique values for this key in stable order
}

func NewSubsetIter(tags []Tag) *SubsetIter {
	seenKey := make(map[string]int)
	seenVal := make(map[string]map[string]bool)
	var dims []chooser

	for _, t := range tags {
		i, ok := seenKey[t.Key]
		if !ok {
			i = len(dims)
			seenKey[t.Key] = i
			seenVal[t.Key] = make(map[string]bool)
			dims = append(dims, chooser{key: t.Key})
		}
		if !seenVal[t.Key][t.Value] {
			dims[i].vals = append(dims[i].vals, t.Value)
			seenVal[t.Key][t.Value] = true
		}
	}

	return &SubsetIter{
		dims: dims,
		idx:  make([]int, len(dims)),
	}
}

// Next returns the next subset of the partition. When reaches the end Next returns 'nil'
func (it *SubsetIter) Next() Subset {
	if it.finished {
		return nil
	}

	for i := 0; i < len(it.idx); i++ {
		radix := len(it.dims[i].vals) + 1
		it.idx[i]++
		if it.idx[i] < radix {
			return it.currentSubset()
		}
		it.idx[i] = 0
	}

	it.finished = true
	return it.currentSubset()
}

// Build the subset for the current counter state.
// Rule per key:
//   idx == len(vals)  -> include all negatives for that key
//   else              -> include exactly one positive for that key
func (it *SubsetIter) currentSubset() Subset {
	out := Subset{}
	for i, d := range it.dims {
		ch := it.idx[i]
		if ch == len(d.vals) {
			for _, v := range d.vals {
				out = append(out, Tag{Key: d.key, Value: v, Not: true})
			}
		} else {
			backIdx := len(d.vals) - ch - 1
			out = append(out, Tag{Key: d.key, Value: d.vals[backIdx], Not: false})
		}
	}
	return out
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
