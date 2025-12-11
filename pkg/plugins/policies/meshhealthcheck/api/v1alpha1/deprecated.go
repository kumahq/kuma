package v1alpha1

import (
	"github.com/kumahq/kuma/v2/pkg/core/resources/model"
	"github.com/kumahq/kuma/v2/pkg/plugins/policies/core/jsonpatch/validators"
	"github.com/kumahq/kuma/v2/pkg/util/pointer"
)

func deprecations(r *model.Res[*MeshHealthCheck]) []string {
	deprecations := validators.TopLevelTargetRefDeprecations(r.Spec.TargetRef)
	for _, to := range pointer.Deref(r.Spec.To) {
		if to.Default.HealthyPanicThreshold != nil {
			deprecations = append(deprecations, "healthyPanicThreshold for 'to[].default' is deprecated. "+
				"The setting has been moved to MeshCircuitBreaker policy, please use MeshCircuitBreaker policy instead.")
		}
	}
	return deprecations
}
