package v1alpha1

import (
	common_proto "github.com/kumahq/kuma/api/common/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/validators"
	matcher_validators "github.com/kumahq/kuma/pkg/plugins/policies/matchers/validators"
)

func (r *MeshTrafficPermissionResource) validate() error {
	var verr validators.ValidationError
	path := validators.RootedAt("spec")

	targetRefErr := matcher_validators.ValidateTargetRef(path.Field("targetRef"), r.Spec.GetTargetRef(), &matcher_validators.ValidateTargetRefOpts{
		SupportedKinds: []common_proto.TargetRef_Kind{
			common_proto.TargetRef_Mesh,
			common_proto.TargetRef_MeshSubset,
			common_proto.TargetRef_MeshService,
			common_proto.TargetRef_MeshServiceSubset,
			common_proto.TargetRef_MeshGatewayRoute,
			common_proto.TargetRef_MeshHTTPRoute,
		},
	})
	verr.AddError("", targetRefErr)

	from := path.Field("from")
	if len(r.Spec.GetFrom()) == 0 {
		verr.AddViolationAt(from, "cannot be empty")
	} else {
		for idx, fromItem := range r.Spec.GetFrom() {
			targetRefErr := matcher_validators.ValidateTargetRef(from.Index(idx).Field("targetRef"), fromItem.GetTargetRef(), &matcher_validators.ValidateTargetRefOpts{
				SupportedKinds: []common_proto.TargetRef_Kind{
					common_proto.TargetRef_Mesh,
					common_proto.TargetRef_MeshSubset,
					common_proto.TargetRef_MeshService,
					common_proto.TargetRef_MeshServiceSubset,
				},
			})
			verr.AddError("", targetRefErr)

			if fromItem.GetDefault() == nil {
				verr.AddViolationAt(from.Index(idx).Field("default"), "cannot be nil")
			}
		}
	}

	return verr.OrNil()
}
