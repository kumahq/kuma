package subsetutils

import (
	"testing"
)

// IsSubset and Intersect were rewritten to scan 'other' directly instead of
// indexing it into a map[string][]Tag, which removed an allocation from the hot
// path of rule building. Both are used to decide which MeshTrafficPermission
// entries apply to a client, so the rewrite must not change a single verdict.
//
// The functions below are the original map-based implementations, kept verbatim as
// a reference. The tests assert that the current implementations agree with them on
// every pair of subsets built from a small tag alphabet.

func isSubsetReference(ss Subset, other Subset) bool {
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

func intersectReference(ss Subset, other Subset) bool {
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

// alphabet covers both keys and values colliding or differing, and both polarities,
// which is everything isSubset distinguishes between.
func alphabet() []Tag {
	var tags []Tag
	for _, key := range []string{"k1", "k2"} {
		for _, value := range []string{"v1", "v2"} {
			for _, not := range []bool{false, true} {
				tags = append(tags, Tag{Key: key, Value: value, Not: not})
			}
		}
	}
	return tags
}

// allSubsets enumerates every ordered tag sequence up to maxLen, including
// sequences that repeat a key, which is the case the map-based versions grouped.
func allSubsets(maxLen int) []Subset {
	tags := alphabet()
	subsets := []Subset{{}}
	current := []Subset{{}}

	for length := 1; length <= maxLen; length++ {
		var next []Subset
		for _, prefix := range current {
			for _, tag := range tags {
				extended := make(Subset, len(prefix), len(prefix)+1)
				copy(extended, prefix)
				next = append(next, append(extended, tag))
			}
		}
		subsets = append(subsets, next...)
		current = next
	}
	return subsets
}

func TestIsSubsetMatchesReference(t *testing.T) {
	subsets := allSubsets(2)
	for _, ss := range subsets {
		for _, other := range subsets {
			got := ss.IsSubset(other)
			want := isSubsetReference(ss, other)
			if got != want {
				t.Fatalf("IsSubset(%v, %v) = %v, reference = %v", ss, other, got, want)
			}
		}
	}
}

func TestIntersectMatchesReference(t *testing.T) {
	subsets := allSubsets(2)
	for _, ss := range subsets {
		for _, other := range subsets {
			got := ss.Intersect(other)
			want := intersectReference(ss, other)
			if got != want {
				t.Fatalf("Intersect(%v, %v) = %v, reference = %v", ss, other, got, want)
			}
		}
	}
}

// The length-3 sweep is the same check over a larger space. It is a separate test so
// the cheap one still runs when this is skipped in short mode.
func TestIsSubsetAndIntersectMatchReferenceDeep(t *testing.T) {
	if testing.Short() {
		t.Skip("exhaustive length-3 sweep")
	}
	subsets := allSubsets(3)
	for _, ss := range subsets {
		for _, other := range subsets {
			if got, want := ss.IsSubset(other), isSubsetReference(ss, other); got != want {
				t.Fatalf("IsSubset(%v, %v) = %v, reference = %v", ss, other, got, want)
			}
			if got, want := ss.Intersect(other), intersectReference(ss, other); got != want {
				t.Fatalf("Intersect(%v, %v) = %v, reference = %v", ss, other, got, want)
			}
		}
	}
}

func TestIsSubsetAndIntersectDoNotAllocate(t *testing.T) {
	ss := Subset{{Key: "kuma.io/service", Value: "caller-1"}}
	other := Subset{{Key: "kuma.io/service", Value: "caller-1"}, {Key: "version", Value: "v1"}}

	if allocs := testing.AllocsPerRun(100, func() { _ = ss.IsSubset(other) }); allocs != 0 {
		t.Errorf("IsSubset allocated %v times per run, want 0", allocs)
	}
	if allocs := testing.AllocsPerRun(100, func() { _ = ss.Intersect(other) }); allocs != 0 {
		t.Errorf("Intersect allocated %v times per run, want 0", allocs)
	}
}
