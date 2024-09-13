package v1alpha1

import (
	"github.com/kumahq/kuma/pkg/plugins/policies/core/jsonpatch/validators"
)

func (t *MeshRetryResource) Deprecations() []string {
	return validators.TopLevelTargetRefDeprecations(t.Spec.TargetRef)
}
