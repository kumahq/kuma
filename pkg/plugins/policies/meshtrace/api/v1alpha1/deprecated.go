package v1alpha1

import (
	"github.com/kumahq/kuma/v2/pkg/plugins/policies/core/jsonpatch/validators"
)

func (r *MeshTraceResource) Deprecations() []string {
	return validators.TopLevelTargetRefDeprecations(r.Spec.TargetRef)
}
