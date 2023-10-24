package v1alpha1

import (
	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/validators"
)

func (r *MeshTCPRouteResource) validate() error {
	var verr validators.ValidationError

	path := validators.RootedAt("spec")

	verr.AddErrorAt(path.Field("targetRef"), validateTop(r.Spec.TargetRef))
	verr.AddErrorAt(path, validateTo(r.Spec.To))

	return verr.OrNil()
}

func validateTop(targetRef common_api.TargetRef) validators.ValidationError {
	return mesh.ValidateTargetRef(
		targetRef,
		&mesh.ValidateTargetRefOpts{
			SupportedKinds: []common_api.TargetRefKind{
				common_api.Mesh,
				common_api.MeshSubset,
				common_api.MeshService,
				common_api.MeshServiceSubset,
			},
		},
	)
}

func validateTo(to []To) validators.ValidationError {
	var verr validators.ValidationError

	for idx, toItem := range to {
		path := validators.RootedAt("to").Index(idx)

		verr.AddErrorAt(
			path.Field("targetRef"),
			mesh.ValidateTargetRef(toItem.TargetRef,
				&mesh.ValidateTargetRefOpts{
					SupportedKinds: []common_api.TargetRefKind{
						common_api.MeshService,
					},
				},
			),
		)
		verr.AddErrorAt(path.Field("rules"), validateRules(toItem.Rules))
	}

	return verr
}

func validateRules(rules []Rule) validators.ValidationError {
	var verr validators.ValidationError

	for i, rule := range rules {
		path := validators.Root().Index(i)

		verr.AddErrorAt(path.Field("default").Field("backendRefs"),
			validateBackendRefs(rule.Default.BackendRefs),
		)
	}

	return verr
}

func validateBackendRefs(backendRefs []common_api.BackendRef) validators.ValidationError {
	var verr validators.ValidationError

	if backendRefs == nil {
		return verr
	}

	for i, backendRef := range backendRefs {
		verr.AddErrorAt(
			validators.Root().Index(i),
			mesh.ValidateTargetRef(
				backendRef.TargetRef,
				&mesh.ValidateTargetRefOpts{
					SupportedKinds: []common_api.TargetRefKind{
						common_api.MeshService,
						common_api.MeshServiceSubset,
					},
				},
			),
		)
	}

	return verr
}
