// Generated by tools/resource-gen.
// Run "make generate" to update this file.

// nolint:whitespace
package v1alpha1

import (
	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
)

func (x *MeshProxyPatch) GetTargetRef() common_api.TargetRef {
	return x.TargetRef
}

func (x *MeshProxyPatch) GetDefault() interface{} {
	return x.Default
}

func (x *MeshProxyPatch) GetPolicyItem() core_xds.PolicyItem {
	return &policyItem{
		MeshProxyPatch: x,
	}
}

// policyItem is an auxiliary struct with the implementation of the GetTargetRef() to always return empty result
type policyItem struct {
	*MeshProxyPatch
}

var _ core_xds.PolicyItem = &policyItem{}

func (p *policyItem) GetTargetRef() common_api.TargetRef {
	return common_api.TargetRef{Kind: common_api.Mesh}
}
