// Generated by tools/resource-gen.
// Run "make generate" to update this file.

// nolint:whitespace
package v1alpha1

import (
	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
)

func (x *MeshHTTPRoute) GetTargetRef() common_api.TargetRef {
	return x.TargetRef
}

func (x *To) GetTargetRef() common_api.TargetRef {
	return x.TargetRef
}

func (x *MeshHTTPRoute) GetToList() []core_xds.PolicyItem {
	var result []core_xds.PolicyItem
	for i := range x.To {
		item := x.To[i]
		result = append(result, &item)
	}
	return result
}
