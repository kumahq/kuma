package inbound

import (
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/rules/common"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/rules/merge"
)

type Rule struct {
	Conf   interface{}     `json:"conf"`
	Origin []common.Origin `json:"origin"`
}

type RuleEntry interface {
	common.BaseEntry
}

// ruleEntryAdapter is a helper struct that allows using any BaseEntry as RuleEntry. For example, this is needed to
// provide backward compatibility for legacy FromEntries and use them as RuleEntries. Currently, RuleEntry and BaseEntry
// are the same, so this adapter is not needed, but in the future RuleEntry is expected to have additional methods
// like GetMatches() and GetTargetRef() that are not present in BaseEntry.
type ruleEntryAdapter[T common.BaseEntry] struct {
	BaseEntry T
}

func newRuleEntryAdapter[T common.BaseEntry](base T) *ruleEntryAdapter[T] {
	return &ruleEntryAdapter[T]{BaseEntry: base}
}

func (r *ruleEntryAdapter[T]) GetDefault() interface{} {
	return r.BaseEntry.GetDefault()
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
		case desc.InterpretFromEntriesAsRules && len(policy.GetFromList()) > 0:
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

	merged, err := merge.Entries(list)
	if err != nil {
		return nil, err
	}

	ruleOrigin, _ := common.Origins(list, true)
	return []*Rule{{
		Conf:   merged,
		Origin: ruleOrigin,
	}}, nil
}
