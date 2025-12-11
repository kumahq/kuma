package v1alpha1

import (
	"github.com/kumahq/kuma/v2/pkg/core/resources/model"
	"github.com/kumahq/kuma/v2/pkg/plugins/policies/core/jsonpatch/validators"
	"github.com/kumahq/kuma/v2/pkg/util/pointer"
)

func deprecations(r *model.Res[*MeshTimeout]) []string {
	var deprecations []string
	if len(pointer.Deref(r.Spec.From)) > 0 {
		deprecations = append(deprecations, "'from' field is deprecated, use 'rules' instead")
	}
	return append(deprecations, validators.TopLevelTargetRefDeprecations(r.Spec.TargetRef)...)
}
