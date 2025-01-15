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

type Entry[T BaseEntry] interface {
	GetEntry() T
}

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
