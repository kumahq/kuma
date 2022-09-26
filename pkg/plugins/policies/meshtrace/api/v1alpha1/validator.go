package v1alpha1

import (
	common_proto "github.com/kumahq/kuma/api/common/v1alpha1"
	matcher_validators "github.com/kumahq/kuma/pkg/plugins/policies/matchers/validators"
	"github.com/kumahq/kuma/pkg/core/validators"
)

func (r *MeshTraceResource) validate() error {
	var verr validators.ValidationError
	path := validators.RootedAt("spec")
	verr.AddErrorAt(path.Field("targetRef"), validateTop(r.Spec.GetTargetRef()))
	return verr.OrNil()
}
func validateTop(targetRef *common_proto.TargetRef) validators.ValidationError {
	targetRefErr := matcher_validators.ValidateTargetRef(targetRef, &matcher_validators.ValidateTargetRefOpts{
		SupportedKinds: []common_proto.TargetRef_Kind{
			// TODO add supported TargetRef kinds for this policy
		},
	})
	return targetRefErr
}

func validateDefault(conf *MeshTrace_Conf) validators.ValidationError {
	var verr validators.ValidationError
	// TODO add default conf validation
	return verr
}
