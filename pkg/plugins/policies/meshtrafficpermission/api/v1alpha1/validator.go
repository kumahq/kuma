package v1alpha1

import (
	common_proto "github.com/kumahq/kuma/api/common/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/validators"
	matcher_validators "github.com/kumahq/kuma/pkg/plugins/policies/matchers/validators"
)

func (r *MeshTrafficPermissionResource) validate() error {
	var verr validators.ValidationError
	path := validators.RootedAt("spec")
	verr.AddErrorAt(path.Field("targetRef"), validateTop(r.Spec.GetTargetRef()))
	if len(r.Spec.GetFrom()) == 0 {
		verr.AddViolationAt(path.Field("from"), "needs at least one item")
	}
	verr.AddErrorAt(path, validateFrom(r.Spec.GetFrom()))
	return verr.OrNil()
}
func validateTop(targetRef *common_proto.TargetRef) validators.ValidationError {
	targetRefErr := matcher_validators.ValidateTargetRef(targetRef, &matcher_validators.ValidateTargetRefOpts{
		SupportedKinds: []common_proto.TargetRef_Kind{
			common_proto.TargetRef_Mesh,
			common_proto.TargetRef_MeshSubset,
			common_proto.TargetRef_MeshService,
			common_proto.TargetRef_MeshServiceSubset,
			common_proto.TargetRef_MeshGatewayRoute,
			common_proto.TargetRef_MeshHTTPRoute,
		},
	})
	return targetRefErr
}
func validateFrom(from []*MeshTrafficPermission_From) validators.ValidationError {
	var verr validators.ValidationError
	for idx, fromItem := range from {
		path := validators.RootedAt("from").Index(idx)
		verr.AddErrorAt(path.Field("targetRef"), matcher_validators.ValidateTargetRef(fromItem.GetTargetRef(), &matcher_validators.ValidateTargetRefOpts{
			SupportedKinds: []common_proto.TargetRef_Kind{
				common_proto.TargetRef_Mesh,
				common_proto.TargetRef_MeshSubset,
				common_proto.TargetRef_MeshService,
				common_proto.TargetRef_MeshServiceSubset,
			},
		}))

		defaultField := path.Field("default")
		if fromItem.GetDefault() == nil {
			verr.AddViolationAt(defaultField, "must be defined")
		} else {
			verr.AddErrorAt(defaultField, validateDefault(fromItem.Default))
		}
	}
	return verr
}

func validateDefault(conf *MeshTrafficPermission_Conf) validators.ValidationError {
	var verr validators.ValidationError
	return verr
}
