// Generated by tools/resource-gen.
// Run "make generate" to update this file.

// nolint:whitespace
package v1alpha1

import (
	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
)

func (x *MeshRetry) GetTargetRef() common_api.TargetRef {
	return x.TargetRef
}

func (x *To) GetTargetRef() common_api.TargetRef {
	return x.TargetRef
}

func (x *To) GetDefault() interface{} {
	return x.Default
}

func (x *MeshRetry) GetToList() []core_model.PolicyItem {
	var result []core_model.PolicyItem
	for i := range x.To {
		item := x.To[i]
		result = append(result, &item)
	}
	return result
}
