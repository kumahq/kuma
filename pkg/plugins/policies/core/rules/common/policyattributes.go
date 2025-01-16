package common

import (
	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
)

type PolicyAttributes interface {
	GetTopLevel() common_api.TargetRef
	GetResourceMeta() core_model.ResourceMeta
	GetRuleIndex() int
}

// Entry is a piece of configuration that is part of a policy. Outbound policies, for example, have a list of entries called 'to'.
// Inbound policies at this moment have a list of entries called 'from'. Entries in 'from' and 'to' have the same type:
//
//	type PolicyItem interface {
//	    GetTargetRef() common_api.TargetRef
//	    GetDefault() interface{}
//	}
//
// But 'from' list of entries is going to be replaced with 'rules' according to docs/madr/decisions/069-inbound-policies.md.
// Entries in 'to' and entries in 'rules' won't have the same type anymore, they're going to have 'ToEntry' and 'RuleEntry'
// types respectively. So, we need to make 'Entry' generic to be able to use it in shared packages like 'sort', 'merge' and 'common'.
type Entry[T BaseEntry] interface {
	GetEntry() T
}

// BaseEntry is a base interface for all entries in policies. Regardless of the type of the entry,
// it should always contain a piece of configuration.
type BaseEntry interface {
	GetDefault() interface{}
}

type WithPolicyAttributes[T any] struct {
	Entry T

	TopLevel  common_api.TargetRef
	Meta      core_model.ResourceMeta
	RuleIndex int
}

func (p WithPolicyAttributes[T]) GetTopLevel() common_api.TargetRef {
	return p.TopLevel
}

func (p WithPolicyAttributes[T]) GetResourceMeta() core_model.ResourceMeta {
	return p.Meta
}

func (p WithPolicyAttributes[T]) GetRuleIndex() int {
	return p.RuleIndex
}

func (p WithPolicyAttributes[T]) GetEntry() T {
	return p.Entry
}
