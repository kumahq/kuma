package v1alpha1

import "slices"

type PolicyDefault struct {
	Rules []Rule `policyMerge:"mergeValuesByKey"`
}

func (x *To) GetDefault() interface{} {
	reversed := slices.Clone(x.Rules)
	slices.Reverse(reversed)
	return PolicyDefault{
		Rules: reversed,
	}
}

func (policy *PolicyDefault) Transform() {
	slices.Reverse(policy.Rules)
}
