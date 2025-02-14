package v1alpha1

import (
	"github.com/kumahq/kuma/pkg/plugins/policies/core/jsonpatch/validators"
)

func (t *MeshTimeoutResource) Deprecations() []string {
	var deprecations []string
	if len(t.Spec.From) > 0 {
		deprecations = append(deprecations, "'from' field is deprecated, use 'rules' instead")
	}
	return append(deprecations, validators.TopLevelTargetRefDeprecations(t.Spec.TargetRef)...)
}
