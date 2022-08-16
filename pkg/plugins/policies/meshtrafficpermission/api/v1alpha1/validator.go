package v1alpha1

import (
	common_proto "github.com/kumahq/kuma/api/common/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/common"
	"github.com/kumahq/kuma/pkg/core/validators"
)

func (t *MeshTrafficPermissionResource) validate() error {
	verr := validators.ValidationError{}
	path := validators.RootedAt("spec")

	if targetRef := t.Spec.GetTargetRef(); targetRef != nil {
		common.ValidateTargetRef(path.Field("targetRef"), targetRef, &common.ValidateTargetRefOpts{
			SupportedKinds: []common_proto.TargetRef_Kind{
				common_proto.TargetRef_Mesh,
				common_proto.TargetRef_MeshSubset,
				common_proto.TargetRef_MeshService,
				common_proto.TargetRef_MeshServiceSubset,
				common_proto.TargetRef_MeshGatewayRoute,
				common_proto.TargetRef_MeshHTTPRoute,
			},
		})
	}

	from := path.Field("from")
	if len(t.Spec.GetFrom()) == 0 {
		verr.AddViolationAt(from, "cannot be empty")
	} else {
		for idx, fromItem := range t.Spec.GetFrom() {
			if targetRef := fromItem.GetTargetRef(); targetRef != nil {
				targetRefErr := common.ValidateTargetRef(from.Index(idx).Field("targetRef"), targetRef, &common.ValidateTargetRefOpts{
					SupportedKinds: []common_proto.TargetRef_Kind{
						common_proto.TargetRef_Mesh,
						common_proto.TargetRef_MeshSubset,
						common_proto.TargetRef_MeshService,
						common_proto.TargetRef_MeshServiceSubset,
					},
				})
				verr.AddError("", targetRefErr)
			}
			if fromItem.GetDefault() == nil {
				verr.AddViolationAt(from.Index(idx).Field("default"), "cannot be nil")
			}
		}
	}

	return verr.OrNil()
}
