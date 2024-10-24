package v1alpha1

import (
	"github.com/kumahq/kuma/api/common/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/jsonpatch/validators"
)

func (t *MeshFaultInjectionResource) Deprecations() []string {
	deprecations := validators.TopLevelTargetRefDeprecations(t.Spec.TargetRef)
	for _, f := range t.Spec.From {
		if f.GetTargetRef().Kind == v1alpha1.MeshService {
			deprecations = append(deprecations, "MeshService value for 'from[].targetRef.kind' is deprecated, use MeshSubset with 'kuma.io/service' instead")
		}
	}
	return deprecations
}
