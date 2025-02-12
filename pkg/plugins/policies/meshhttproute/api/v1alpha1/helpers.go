package v1alpha1

import (
	"github.com/kumahq/kuma/pkg/util/pointer"
	"slices"
)

type PolicyDefault struct {
	Rules     []Rule   `json:"rules,omitempty" policyMerge:"mergeValuesByKey"`
	Hostnames []string `json:"hostnames,omitempty" policyMerge:"mergeValues"`
}

func (x *To) GetDefault() interface{} {
	reversed := slices.Clone(x.Rules)
	slices.Reverse(reversed)
	return PolicyDefault{
		Rules:     reversed,
		Hostnames: pointer.Deref(x.Hostnames),
	}
}

func (policy *PolicyDefault) Transform() {
	slices.Reverse(policy.Rules)
}
