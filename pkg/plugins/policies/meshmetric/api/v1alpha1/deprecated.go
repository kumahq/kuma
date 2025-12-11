package v1alpha1

import (
	"github.com/kumahq/kuma/v2/pkg/core/resources/model"
	"github.com/kumahq/kuma/v2/pkg/plugins/policies/core/jsonpatch/validators"
)

func deprecations(r *model.Res[*MeshMetric]) []string {
	return validators.TopLevelTargetRefDeprecations(r.Spec.TargetRef)
}
