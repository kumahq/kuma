package v1alpha1

import (
	"github.com/kumahq/kuma/pkg/plugins/policies/core/jsonpatch/validators"
	"github.com/kumahq/kuma/pkg/util/pointer"
)

func (t *MeshHealthCheckResource) Deprecations() []string {
	deprecations := validators.TopLevelTargetRefDeprecations(t.Spec.TargetRef)
	for _, to := range pointer.Deref(t.Spec.To) {
		if to.Default.HealthyPanicThreshold != nil {
			deprecations = append(deprecations, "healthyPanicThreshold for 'to[].default' is deprecated, use MeshCircuitBreaker instead")
		}
	}
	return deprecations
}
