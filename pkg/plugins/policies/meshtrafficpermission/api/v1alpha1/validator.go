package v1alpha1

import (
	common_proto "github.com/kumahq/kuma/api/common/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/validators"
	matcher_validators "github.com/kumahq/kuma/pkg/plugins/policies/matchers/validators"
)

func (r *MeshTrafficPermissionResource) validate() error {
	var verr validators.ValidationError
	path := validators.RootedAt("spec")
	verr.AddErrorAt(path.Field("targetRef"), validateTop(r.Spec.TargetRef))
	if len(r.Spec.From) == 0 {
		verr.AddViolationAt(path.Field("from"), "needs at least one item")
	}
	verr.AddErrorAt(path, validateFrom(r.Spec.From))
	return verr.OrNil()
}
func validateTop(targetRef common_proto.TargetRef) validators.ValidationError {
	targetRefErr := matcher_validators.ValidateTargetRef(targetRef, &matcher_validators.ValidateTargetRefOpts{
		SupportedKinds: []common_proto.TargetRefKind{
			common_proto.Mesh,
			common_proto.MeshSubset,
			common_proto.MeshService,
			common_proto.MeshServiceSubset,
			common_proto.MeshGatewayRoute,
			common_proto.MeshHTTPRoute,
		},
	})
	return targetRefErr
}
func validateFrom(from []*From) validators.ValidationError {
	var verr validators.ValidationError
	for idx, fromItem := range from {
		path := validators.RootedAt("from").Index(idx)
		verr.AddErrorAt(path.Field("targetRef"), matcher_validators.ValidateTargetRef(fromItem.TargetRef, &matcher_validators.ValidateTargetRefOpts{
			SupportedKinds: []common_proto.TargetRefKind{
				common_proto.Mesh,
				common_proto.MeshSubset,
				common_proto.MeshService,
				common_proto.MeshServiceSubset,
			},
		}))
	}
	return verr
}
