package v1alpha1

import (
	"github.com/kumahq/kuma/pkg/plugins/policies/core/jsonpatch/validators"
	"github.com/kumahq/kuma/pkg/util/pointer"
)

func (t *MeshAccessLogResource) Deprecations() []string {
	var deprecations []string
	if len(pointer.Deref(t.Spec.From)) > 0 {
		deprecations = append(deprecations, "'from' field is deprecated, use 'rules' instead")
	}
	return append(deprecations, validators.TopLevelTargetRefDeprecations(t.Spec.TargetRef)...)
}
