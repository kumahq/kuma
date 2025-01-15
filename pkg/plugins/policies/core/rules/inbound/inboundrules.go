package inbound

import (
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/rules/common"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/rules/merge"
)

type Rule struct {
	Conf interface{}
}

type RuleEntry interface {
	common.BaseEntry
}

type PolicyWithRules interface {
	core_model.Policy
	GetRules() []RuleEntry
}

func BuildRules(policies []core_model.Resource) ([]*Rule, error) {
	return buildRules(getEntries(policies))
}

func getEntries(policies []core_model.Resource) []common.WithPolicyAttributes[RuleEntry] {
	policiesWithTo, ok := common.Cast[PolicyWithRules](policies)
	if !ok {
		return nil
	}

	entries := []common.WithPolicyAttributes[RuleEntry]{}

	for i, pwt := range policiesWithTo {
		for j, rule := range pwt.GetRules() {
			entries = append(entries, common.WithPolicyAttributes[RuleEntry]{
				Entry:     rule,
				Meta:      policies[i].GetMeta(),
				TopLevel:  pwt.GetTargetRef(),
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
		return nil, nil
	}

	Sort(list)

	merged, err := merge.Entries(list)
	if err != nil {
		return nil, err
	}

	return []*Rule{
		{Conf: merged},
	}, nil
}
