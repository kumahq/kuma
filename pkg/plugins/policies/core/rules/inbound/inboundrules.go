package inbound

import (
	common_api "github.com/kumahq/kuma/v3/api/common/v1alpha1"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
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

type PolicyWithRules interface {
	core_model.Policy
	GetRules() []RuleEntry
}

func BuildRules(policies core_model.ResourceList) ([]*Rule, error) {
	entries := getEntries(policies)
	return buildRules(entries)
}

func getEntries(resources core_model.ResourceList) []common.WithPolicyAttributes[RuleEntry] {
	policies, ok := common.Cast[PolicyWithRules](resources.GetItems())
	if !ok {
		return nil
	}

	entries := []common.WithPolicyAttributes[RuleEntry]{}

	for i, policy := range policies {
		for j, rule := range policy.GetRules() {
			entries = append(entries, common.WithPolicyAttributes[RuleEntry]{
				Entry:     rule,
				Meta:      resources.GetItems()[i].GetMeta(),
				TopLevel:  policy.GetTargetRef(),
				RuleIndex: j,
			})
		}
	}

	return entries
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
	return ok && len(pr.GetRules()) > 0
}
