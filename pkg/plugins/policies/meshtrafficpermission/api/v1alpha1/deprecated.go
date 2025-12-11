package v1alpha1

import (
	"github.com/kumahq/kuma/v2/api/common/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/core/resources/model"
	"github.com/kumahq/kuma/v2/pkg/plugins/policies/core/jsonpatch/validators"
	"github.com/kumahq/kuma/v2/pkg/util/pointer"
)

func deprecations(r *model.Res[*MeshTrafficPermission]) []string {
	deprecations := validators.TopLevelTargetRefDeprecations(r.Spec.TargetRef)
	for _, f := range pointer.Deref(r.Spec.From) {
		if f.GetTargetRef().Kind == v1alpha1.MeshService {
			deprecations = append(deprecations, "MeshService value for 'from[].targetRef.kind' is deprecated, use MeshSubset with 'kuma.io/service' instead")
		}
	}
	return deprecations
}
