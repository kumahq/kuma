package v1alpha1

import (
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/jsonpatch/validators"
)

func (t *MeshTraceResource) Deprecations() []string {
	return validators.TopLevelTargetRefDeprecations(t.Spec.TargetRef)
}
