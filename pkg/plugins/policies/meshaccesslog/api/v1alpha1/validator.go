package v1alpha1

import (
	common_proto "github.com/kumahq/kuma/api/common/v1alpha1"
	matcher_validators "github.com/kumahq/kuma/pkg/plugins/policies/matchers/validators"
	"github.com/kumahq/kuma/pkg/core/validators"
)

func (r *MeshAccessLogResource) validate() error {
	var verr validators.ValidationError
	path := validators.RootedAt("spec")

	targetRefErr := matcher_validators.ValidateTargetRef(path.Field("targetRef"), r.Spec.GetTargetRef(), &matcher_validators.ValidateTargetRefOpts{
		SupportedKinds: []common_proto.TargetRef_Kind{
			// TODO add supported TargetRef kinds for this policy
		},
	})
	verr.AddError("", targetRefErr)

	return verr.OrNil()
}
