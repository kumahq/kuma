package v1alpha1

import (
	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/validators"
	"github.com/kumahq/kuma/pkg/util/pointer"
)

func (r *MeshTCPRouteResource) validate() error {
	var verr validators.ValidationError

	path := validators.RootedAt("spec")

	verr.AddErrorAt(path.Field("targetRef"), r.validateTop(r.Spec.TargetRef))
	verr.AddErrorAt(path, validateTo(pointer.DerefOr(r.Spec.TargetRef, common_api.TargetRef{Kind: common_api.Mesh}), r.Spec.To))

	return verr.OrNil()
}

func (r *MeshTCPRouteResource) validateTop(targetRef *common_api.TargetRef) validators.ValidationError {
	if targetRef == nil {
		return validators.ValidationError{}
	}
	switch core_model.PolicyRole(r.GetMeta()) {
	case mesh_proto.SystemPolicyRole:
		return mesh.ValidateTargetRef(*targetRef, &mesh.ValidateTargetRefOpts{
			SupportedKinds: []common_api.TargetRefKind{
				common_api.Mesh,
				common_api.MeshSubset,
				common_api.MeshGateway,
				common_api.MeshService,
				common_api.MeshServiceSubset,
			},
			GatewayListenerTagsAllowed: true,
		})
	default:
		return mesh.ValidateTargetRef(*targetRef, &mesh.ValidateTargetRefOpts{
			SupportedKinds: []common_api.TargetRefKind{
				common_api.Mesh,
				common_api.MeshSubset,
				common_api.MeshService,
				common_api.MeshServiceSubset,
			},
		})
	}
}

func validateToRef(topTargetRef, targetRef common_api.TargetRef) validators.ValidationError {
	switch topTargetRef.Kind {
	case common_api.MeshGateway:
		return mesh.ValidateTargetRef(targetRef, &mesh.ValidateTargetRefOpts{
			SupportedKinds: []common_api.TargetRefKind{
				common_api.Mesh,
			},
		})
	default:
		return mesh.ValidateTargetRef(targetRef, &mesh.ValidateTargetRefOpts{
			SupportedKinds: []common_api.TargetRefKind{
				common_api.MeshService,
				common_api.MeshExternalService,
			},
		})
	}
}

func validateTo(topTargetRef common_api.TargetRef, to []To) validators.ValidationError {
	var verr validators.ValidationError

	for idx, toItem := range to {
		path := validators.RootedAt("to").Index(idx)

		verr.AddErrorAt(path.Field("targetRef"), validateToRef(topTargetRef, toItem.TargetRef))
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
						common_api.MeshExternalService,
					},
				},
			),
		)
	}

	return verr
}
