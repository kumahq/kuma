package common

import (
	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	meshhttproute_api "github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/api/v1alpha1"
)

type Origin struct {
	Resource core_model.ResourceMeta
	// RuleIndex is an index in the 'to[]' array, so we could unambiguously detect what to-item contributed to the final conf.
	// Especially useful when to-item uses `targetRef.Labels`, because there is no obvious matching between the specific resource
	// in `ResourceRule.Resource` and to-item.
	RuleIndex int
}

type keyType struct {
	core_model.ResourceKey
	ruleIndex int
}

func getKey(policyItem PolicyAttributes, withRuleIndex bool) keyType {
	k := keyType{
		ResourceKey: core_model.MetaToResourceKey(policyItem.GetResourceMeta()),
	}
	if withRuleIndex {
		k.ruleIndex = policyItem.GetRuleIndex()
	}
	return k
}

func Origins[B BaseEntry, T interface {
	PolicyAttributes
	Entry[B]
}](items []T, withRuleIndex bool) []Origin {
	var rv []Origin

	set := map[keyType]struct{}{}
	for _, item := range items {
		if _, ok := set[getKey(item, withRuleIndex)]; !ok {
			set[getKey(item, withRuleIndex)] = struct{}{}

			rv = append(rv, Origin{
				Resource:  item.GetResourceMeta(),
				RuleIndex: item.GetRuleIndex(),
			})
		}
	}
	return rv
}

func OriginByMatches[B BaseEntry, T interface {
	PolicyAttributes
	Entry[B]
}](items []T) map[common_api.MatchesHash]Origin {
	rv := map[common_api.MatchesHash]Origin{}

	set := map[keyType]struct{}{}
	for _, item := range items {
		if _, ok := set[getKey(item, true)]; !ok {
			set[getKey(item, true)] = struct{}{}

			origin := Origin{
				Resource:  item.GetResourceMeta(),
				RuleIndex: item.GetRuleIndex(),
			}

			if conf, ok := item.GetEntry().GetDefault().(meshhttproute_api.PolicyDefault); ok {
				for _, rule := range conf.Rules {
					rv[meshhttproute_api.HashMatches(rule.Matches)] = origin
				}
			}
		}
	}

	return rv
}
