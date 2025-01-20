package common

import (
	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	meshhttproute_api "github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	meshtcproute_api "github.com/kumahq/kuma/pkg/plugins/policies/meshtcproute/api/v1alpha1"
)

type Origin struct {
	Resource core_model.ResourceMeta
	// RuleIndex is an index in the 'to[]' array, so we could unambiguously detect what to-item contributed to the final conf.
	// Especially useful when to-item uses `targetRef.Labels`, because there is no obvious matching between the specific resource
	// in `ResourceRule.Resource` and to-item.
	RuleIndex int
}

type BackendRefOriginIndex map[common_api.MatchesHash]int

var EmptyMatches common_api.MatchesHash = ""

func (originIndex BackendRefOriginIndex) Update(conf interface{}, newIndex int) {
	switch conf := conf.(type) {
	case meshtcproute_api.Rule:
		if conf.Default.BackendRefs != nil {
			originIndex[EmptyMatches] = newIndex
		}
	case meshhttproute_api.PolicyDefault:
		for _, rule := range conf.Rules {
			if rule.Default.BackendRefs != nil {
				hash := meshhttproute_api.HashMatches(rule.Matches)
				originIndex[hash] = newIndex
			}
		}
	default:
		return
	}
}

func Origins[B BaseEntry, T interface {
	PolicyAttributes
	Entry[B]
}](items []T, withRuleIndex bool) ([]Origin, BackendRefOriginIndex) {
	var rv []Origin

	type keyType struct {
		core_model.ResourceKey
		ruleIndex int
	}
	key := func(policyItem T) keyType {
		k := keyType{
			ResourceKey: core_model.MetaToResourceKey(policyItem.GetResourceMeta()),
		}
		if withRuleIndex {
			k.ruleIndex = policyItem.GetRuleIndex()
		}
		return k
	}
	set := map[keyType]struct{}{}
	originIndex := BackendRefOriginIndex{}
	for _, item := range items {
		if _, ok := set[key(item)]; !ok {
			originIndex.Update(item.GetEntry().GetDefault(), len(rv))
			rv = append(rv, Origin{Resource: item.GetResourceMeta(), RuleIndex: item.GetRuleIndex()})
			set[key(item)] = struct{}{}
		}
	}
	return rv, originIndex
}
