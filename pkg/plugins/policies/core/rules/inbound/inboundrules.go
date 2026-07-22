package inbound

import (
	common_api "github.com/kumahq/kuma/v3/api/common/v1alpha1"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/core/resources/registry"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules/common"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules/merge"
	util_slices "github.com/kumahq/kuma/v3/pkg/util/slices"
)

type Rule struct {
	Match  *common_api.Match `json:"match"`
	Conf   any               `json:"conf"`
	Origin common.Origin     `json:"origin"`
}

func MatchesAllIncomingTraffic[T any](rules []*Rule) T {
	var result T
	catchAll := util_slices.Filter(rules, func(r *Rule) bool { return r.Match == nil })
	if len(catchAll) == 0 {
		return result
	}
	confs := util_slices.Map(catchAll, func(a *Rule) any { return a.Conf })
	conf, err := merge.Confs(confs)
	if err != nil {
		return result
	}
	if len(conf) > 0 {
		result = conf[0].(T)
	}
	return result
}

type RuleEntry interface {
	common.BaseEntry
	GetMatches() []common_api.Match
}

// ruleEntryAdapter is a helper struct that allows using any BaseEntry as RuleEntry. For example, this is needed to
// provide backward compatibility for legacy FromEntries and use them as RuleEntries without match conditions.
type ruleEntryAdapter[T common.BaseEntry] struct {
	BaseEntry T
}

func newRuleEntryAdapter[T common.BaseEntry](base T) *ruleEntryAdapter[T] {
	return &ruleEntryAdapter[T]{BaseEntry: base}
}

func (r *ruleEntryAdapter[T]) GetDefault() any {
	return r.BaseEntry.GetDefault()
}

func (r *ruleEntryAdapter[T]) GetMatches() []common_api.Match {
	return []common_api.Match{}
}

type PolicyWithRules interface {
	core_model.Policy
	GetRules() []RuleEntry
}

func BuildRules(policies core_model.ResourceList) ([]*Rule, error) {
	entries, err := getEntries(policies)
	if err != nil {
		return []*Rule{}, err
	}
	return buildRules(entries)
}

func getEntries(resources core_model.ResourceList) ([]common.WithPolicyAttributes[RuleEntry], error) {
	desc, err := registry.Global().DescriptorFor(resources.GetItemType())
	if err != nil {
		return nil, err
	}

	policies, ok := common.Cast[PolicyWithRules](resources.GetItems())
	if !ok {
		return nil, nil
	}

	entries := []common.WithPolicyAttributes[RuleEntry]{}

	for i, policy := range policies {
		switch {
		case len(policy.GetRules()) > 0:
			for j, rule := range policy.GetRules() {
				entries = append(entries, common.WithPolicyAttributes[RuleEntry]{
					Entry:     rule,
					Meta:      resources.GetItems()[i].GetMeta(),
					TopLevel:  policy.GetTargetRef(),
					RuleIndex: j,
				})
			}
		case desc.IsFromAsRules:
			policyWithFrom, ok := policy.(core_model.PolicyWithFromList)
			if !ok {
				continue
			}
			for j, fromEntry := range policyWithFrom.GetFromList() {
				entries = append(entries, common.WithPolicyAttributes[RuleEntry]{
					Entry:     newRuleEntryAdapter(fromEntry),
					Meta:      resources.GetItems()[i].GetMeta(),
					TopLevel:  policy.GetTargetRef(),
					RuleIndex: j,
				})
			}
		}
	}

	return entries, nil
}

func buildRules[T interface {
	common.PolicyAttributes
	common.Entry[RuleEntry]
}](list []T) ([]*Rule, error) {
	if len(list) == 0 {
		return []*Rule{}, nil
	}

	Sort(list)

	// Build rules, expanding multi-match entries so each Rule carries exactly one match.
	var rules []*Rule
	for _, entry := range list {
		origin := common.Origin{
			Resource:  entry.GetResourceMeta(),
			RuleIndex: entry.GetRuleIndex(),
		}
		matches := entry.GetEntry().GetMatches()
		if len(matches) == 0 {
			rules = append(rules, &Rule{
				Conf:   entry.GetEntry().GetDefault(),
				Origin: origin,
			})
		} else {
			for k := range matches {
				rules = append(rules, &Rule{
					Match:  &matches[k],
					Conf:   entry.GetEntry().GetDefault(),
					Origin: origin,
				})
			}
		}
	}

	SortRules(rules)
	return rules, nil
}

func AffectsInbounds(p core_model.Policy) bool {
	pr, ok := p.(PolicyWithRules)
	if ok && len(pr.GetRules()) > 0 {
		return true
	}
	pf, ok := p.(core_model.PolicyWithFromList)
	if ok && len(pf.GetFromList()) > 0 {
		return true
	}
	return false
}
