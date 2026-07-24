package v1alpha1

import (
	common_api "github.com/kumahq/kuma/v3/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/core/validators"
	"github.com/kumahq/kuma/v3/pkg/util/pointer"
)

func (r *MeshTCPRouteResource) validate() error {
	var verr validators.ValidationError

	path := validators.RootedAt("spec")

	verr.AddErrorAt(path.Field("targetRef"), r.validateTop(r.Spec.TargetRef))
	verr.AddErrorAt(path, validateTo(pointer.Deref(r.Spec.To)))

	return verr.OrNil()
}

func (r *MeshTCPRouteResource) validateTop(targetRef *common_api.TopLevelTargetRef) validators.ValidationError {
	if targetRef == nil {
		return validators.ValidationError{}
	}
	switch core_model.PolicyRole(r.GetMeta()) {
	case mesh_proto.SystemPolicyRole:
		return mesh.ValidateTargetRef(targetRef.ToTargetRef(), &mesh.ValidateTargetRefOpts{
			SupportedKinds: []common_api.TargetRefKind{
				common_api.Mesh,
				common_api.Dataplane,
			},
			GatewayListenerTagsAllowed: true,
		})
	default:
		return mesh.ValidateTargetRef(targetRef.ToTargetRef(), &mesh.ValidateTargetRefOpts{
			SupportedKinds: []common_api.TargetRefKind{
				common_api.Mesh,
				common_api.Dataplane,
			},
		})
	}
}

func validateToRef(targetRef common_api.OutboundTargetRef) validators.ValidationError {
	return mesh.ValidateTargetRef(targetRef.ToTargetRef(), &mesh.ValidateTargetRefOpts{
		SupportedKinds: []common_api.TargetRefKind{
			common_api.MeshService,
			common_api.MeshExternalService,
			common_api.MeshMultiZoneService,
		},
	})
}

func validateTo(to []To) validators.ValidationError {
	var verr validators.ValidationError

	for idx, toItem := range to {
		path := validators.RootedAt("to").Index(idx)

		verr.AddErrorAt(path.Field("targetRef"), validateToRef(toItem.TargetRef))
		verr.AddErrorAt(path.Field("rules"), validateRules(toItem.Rules))
	}

	return verr
}

func validateRules(rules []Rule) validators.ValidationError {
	var verr validators.ValidationError

	for i, rule := range rules {
		path := validators.Root().Index(i)

		if len(pointer.Deref(rule.Default.BackendRefs)) == 0 {
			verr.AddViolationAt(path.Field("default").Field("backendRefs"), validators.MustBeDefined)
		} else {
			verr.AddErrorAt(path.Field("default").Field("backendRefs"),
				validateBackendRefs(pointer.Deref(rule.Default.BackendRefs)),
			)
		}
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
						common_api.LegacyMeshServiceSubsetKind(),
						common_api.MeshExternalService,
						common_api.MeshMultiZoneService,
					},
				},
			),
		)
		verr.AddErrorAt(
			validators.Root().Index(i),
			validators.ValidateBackendRef(backendRef),
		)
	}

	return verr
}
