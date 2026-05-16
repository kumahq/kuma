package inbound

import (
	common_api "github.com/kumahq/kuma/v2/api/common/v1alpha1"
	core_model "github.com/kumahq/kuma/v2/pkg/core/resources/model"
	"github.com/kumahq/kuma/v2/pkg/core/resources/registry"
	"github.com/kumahq/kuma/v2/pkg/plugins/policies/core/rules/common"
	"github.com/kumahq/kuma/v2/pkg/plugins/policies/core/rules/merge"
	util_slices "github.com/kumahq/kuma/v2/pkg/util/slices"
)

type Rule struct {
	Matches []common_api.Match `json:"matches,omitempty"`
	Conf    any                `json:"conf"`
	Origin  common.Origin      `json:"origin"`
}

func MatchesAllIncomingTraffic[T any](rules []*Rule) T {
	var result T
	if len(rules) > 0 {
		confs := util_slices.Map(rules, func(a *Rule) any {
			return effectiveConf(a.Conf)
		})
		conf, err := merge.Confs(confs)
		if err != nil {
			return result
		}
		if len(conf) > 0 {
			result = conf[0].(T)
		}
	}
	return result
}

func effectiveConf(conf any) any {
	if entry, ok := conf.(common.BaseEntry); ok {
		return entry.GetDefault()
	}
	return conf
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
	return nil
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

	policies, ok := common.Cast[interface {
		PolicyWithRules
		core_model.PolicyWithFromList
	}](resources.GetItems())
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
		case desc.IsFromAsRules && len(policy.GetFromList()) > 0:
			for j, fromEntry := range policy.GetFromList() {
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

	// Build rules one entry at a time so the origin stays tied to the policy that produced the config.
	rules := make([]*Rule, 0, len(list))
	for _, entry := range list {
		rules = append(rules, &Rule{
			Matches: entry.GetEntry().GetMatches(),
			Conf:    entry.GetEntry().GetDefault(),
			Origin: common.Origin{
				Resource:  entry.GetResourceMeta(),
				RuleIndex: entry.GetRuleIndex(),
			},
		})
	}

	// Split multi-match rules so sorting and xDS generation can handle each match independently.
	var expanded []*Rule
	for _, rule := range rules {
		if len(rule.Matches) == 0 {
			expanded = append(expanded, rule)
			continue
		}
		for _, match := range rule.Matches {
			expanded = append(expanded, &Rule{
				Matches: []common_api.Match{match},
				Conf:    rule.Conf,
				Origin:  rule.Origin,
			})
		}
	}
	SortRules(expanded)

	return expanded, nil
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
