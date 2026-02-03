package v1alpha1

import (
	"github.com/kumahq/kuma/v2/pkg/plugins/policies/core/jsonpatch/validators"
)

func (t *MeshHealthCheckResource) Deprecations() []string {
	return validators.TopLevelTargetRefDeprecations(t.Spec.TargetRef)
}
