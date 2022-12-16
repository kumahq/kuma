package xds

import (
	"sort"

	"github.com/pkg/errors"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
)

// Tag is a key-value pair. If Not is true then Key != Value
type Tag struct {
	Key   string
	Value string
	Not   bool
}

// Subset represents a group of proxies
type Subset []Tag

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
			if otherTag.Value == tag.Value && otherTag.Not != tag.Not {
				return false
			}
			if otherTag.Value != tag.Value && !otherTag.Not {
				return false
			}
		}
	}
	return true
}

func MeshSubset() Subset {
	return Subset{}
}

func MeshService(name string) Subset {
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

func (ss Subset) IndexOfPositive() int {
	for i, t := range ss {
		if !t.Not {
			return i
		}
	}
	return -1
}

// Rule contains a configuration for the given Subset. When rule is an inbound rule (from),
// then Subset represents a group of clients. When rule is an outbound (to) then Subset
// represents destinations.
type Rule struct {
	Subset Subset
	Conf   interface{}
	Origin []core_model.ResourceMeta
}

type Rules []*Rule

// Compute returns configuration for the given subset.
func (rs Rules) Compute(sub Subset) *Rule {
	for _, rule := range rs {
		if rule.Subset.IsSubset(sub) {
			return rule
		}
	}
	return nil
}

// BuildRules creates a list of rules with negations sorted by the number of positive tags.
// If rules with negative tags are filtered out then the order becomes 'most specific to less specific'.
// Filtering out of negative rules could be useful for XDS generators that don't have a way to configure negations.
//
// See the detailed algorithm description in docs/madr/decisions/007-mesh-traffic-permission.md
func BuildRules(list []PolicyItemWithMeta) (Rules, error) {
	rules := Rules{}

	// 1. Each targetRef should be represented as a list of tags
	tagSet := map[Tag]bool{}
	for _, item := range list {
		ss, err := asSubset(item.GetTargetRef())
		if err != nil {
			return nil, err
		}
		for _, t := range ss {
			tagSet[t] = true
		}
	}
	tags := []Tag{}
	for tag := range tagSet {
		tags = append(tags, tag)
	}

	sort.Slice(tags, func(i, j int) bool {
		if tags[i].Key != tags[j].Key {
			return tags[i].Key < tags[j].Key
		}
		return tags[i].Value < tags[j].Value
	})

	// 2. Iterate over all possible combinations with negations
	iter := NewSubsetIter(tags)
	for {
		ss := iter.Next()
		if ss == nil {
			break
		}
		// 3. For each combination determine a configuration
		confs := []interface{}{}
		// confs := []PolicyConf{}
		distinctOrigins := map[core_model.ResourceKey]core_model.ResourceMeta{}
		for i := 0; i < len(list); i++ {
			item := list[i]
			itemSubset, err := asSubset(item.GetTargetRef())
			if err != nil {
				return nil, err
			}
			if itemSubset.IsSubset(ss) {
				confs = append(confs, item.GetDefault())
				distinctOrigins[core_model.MetaToResourceKey(item.ResourceMeta)] = item.ResourceMeta
			}
		}
		merged, err := MergeConfs(confs)
		if err != nil {
			return nil, err
		}
		if merged != nil {
			var origins []core_model.ResourceMeta
			for _, origin := range distinctOrigins {
				origins = append(origins, origin)
			}
			sort.Slice(origins, func(i, j int) bool {
				return origins[i].GetName() < origins[j].GetName()
			})
			rules = append(rules, &Rule{
				Subset: ss,
				Conf:   merged,
				Origin: origins,
			})
		}
	}

	sort.SliceStable(rules, func(i, j int) bool {
		return rules[i].Subset.NumPositive() > rules[j].Subset.NumPositive()
	})

	return rules, nil
}

func asSubset(tr common_api.TargetRef) (Subset, error) {
	switch tr.Kind {
	case common_api.Mesh:
		return Subset{}, nil
	case common_api.MeshSubset:
		ss := Subset{}
		for k, v := range tr.Tags {
			ss = append(ss, Tag{Key: k, Value: v})
		}
		return ss, nil
	case common_api.MeshService:
		return Subset{{Key: mesh_proto.ServiceTag, Value: tr.Name}}, nil
	case common_api.MeshServiceSubset:
		ss := Subset{{Key: mesh_proto.ServiceTag, Value: tr.Name}}
		for k, v := range tr.Tags {
			ss = append(ss, Tag{Key: k, Value: v})
		}
		return ss, nil
	default:
		return nil, errors.Errorf("can't represent %s as tags", tr.Kind)
	}
}

type SubsetIter struct {
	current  []Tag
	finished bool
}

func NewSubsetIter(tags []Tag) *SubsetIter {
	return &SubsetIter{
		current: tags,
	}
}

// Next returns the next subset of the partition. When reaches the end Next returns 'nil'
func (c *SubsetIter) Next() Subset {
	if c.finished {
		return nil
	}
	for {
		hasNext := c.next()
		if !hasNext {
			c.finished = true
			return c.simplified()
		}
		if result := c.simplified(); result != nil {
			return result
		}
	}
}

func (c *SubsetIter) next() bool {
	for idx := 0; idx < len(c.current); idx++ {
		if c.current[idx].Not {
			c.current[idx].Not = false
		} else {
			c.current[idx].Not = true
			return true
		}
	}
	return false
}

// simplified returns copy of c.current and deletes redundant tags, for example:
//   - env: dev
//   - env: !prod
//
// could be simplified to:
//   - env: dev
//
// If tags are contradicted (same keys have different positive value) then the function
// returns nil.
func (c *SubsetIter) simplified() Subset {
	result := Subset{}

	ssByKey := map[string]Subset{}
	keyOrder := []string{}
	for _, t := range c.current {
		if _, ok := ssByKey[t.Key]; !ok {
			keyOrder = append(keyOrder, t.Key)
		}
		ssByKey[t.Key] = append(ssByKey[t.Key], Tag{Key: t.Key, Value: t.Value, Not: t.Not})
	}

	for _, key := range keyOrder {
		ss := ssByKey[key]
		positive := ss.NumPositive()
		switch {
		case positive == 0:
			result = append(result, ss...)
		case positive == 1:
			result = append(result, ss[ss.IndexOfPositive()])
		case positive >= 2:
			// contradicted, at least 2 positive values for the same key, i.e 'key1: value1' and 'key1: value2'
			return nil
		}
	}

	return result
}
