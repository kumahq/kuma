// Generated by tools/policy-gen.
// Run "make generate" to update this file.

// nolint:whitespace
package v1alpha1

import (
	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/util/pointer"
)

func (x *MeshTrafficPermission) GetTargetRef() common_api.TargetRef {
	return pointer.DerefOr(x.TargetRef, common_api.TargetRef{Kind: common_api.Mesh, UsesSyntacticSugar: true})
}

func (x *From) GetTargetRef() common_api.TargetRef {
	return pointer.Deref(x.TargetRef)
}

func (x *From) GetDefault() interface{} {
	return x.Default
}

func (x *MeshTrafficPermission) GetFromList() []core_model.PolicyItem {
	var result []core_model.PolicyItem
	for _, itm := range pointer.Deref(x.From) {
		item := itm
		result = append(result, &item)
	}
	return result
}
