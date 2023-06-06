// Generated by tools/resource-gen.
// Run "make generate" to update this file.

// nolint:whitespace
package v1alpha1

import (
	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
)

func (x *MeshTrace) GetTargetRef() common_api.TargetRef {
	return x.TargetRef
}

func (x *MeshTrace) GetDefault() interface{} {
	return x.Default
}

func (x *MeshTrace) GetPolicyItem() core_model.PolicyItem {
	return &policyItem{
		MeshTrace: x,
	}
}

// policyItem is an auxiliary struct with the implementation of the GetTargetRef() to always return empty result
type policyItem struct {
	*MeshTrace
}

var _ core_model.PolicyItem = &policyItem{}

func (p *policyItem) GetTargetRef() common_api.TargetRef {
	return common_api.TargetRef{Kind: common_api.Mesh}
}
