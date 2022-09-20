package v1alpha1

import (
	common_proto "github.com/kumahq/kuma/api/common/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/validators"
	matcher_validators "github.com/kumahq/kuma/pkg/plugins/policies/matchers/validators"
)

func (r *DoNothingPolicyResource) validate() error {
	var verr validators.ValidationError
	path := validators.RootedAt("spec")
	verr.AddErrorAt(path.Field("targetRef"), validateTop(r.Spec.GetTargetRef()))
	if len(r.Spec.GetTo()) == 0 && len(r.Spec.GetFrom()) == 0 {
		verr.AddViolationAt(path, "at least one of 'from', 'to' has to be defined")
	}
	verr.AddErrorAt(path, validateTo(r.Spec.GetTo()))
	verr.AddErrorAt(path, validateFrom(r.Spec.GetFrom()))
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
func validateFrom(from []*DoNothingPolicy_From) validators.ValidationError {
	var verr validators.ValidationError
	for idx, fromItem := range from {
		path := validators.RootedAt("from").Index(idx)
		verr.AddErrorAt(path.Field("targetRef"), matcher_validators.ValidateTargetRef(fromItem.GetTargetRef(), &matcher_validators.ValidateTargetRefOpts{
			SupportedKinds: []common_proto.TargetRef_Kind{
				// TODO add supported TargetRef for 'from' item
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
func validateTo(to []*DoNothingPolicy_To) validators.ValidationError {
	var verr validators.ValidationError
	for idx, toItem := range to {
		path := validators.RootedAt("to").Index(idx)
		verr.AddErrorAt(path.Field("targetRef"), matcher_validators.ValidateTargetRef(toItem.GetTargetRef(), &matcher_validators.ValidateTargetRefOpts{
			SupportedKinds: []common_proto.TargetRef_Kind{
				// TODO add supported TargetRef for 'to' item
			},
		}))

		defaultField := path.Field("default")
		if toItem.GetDefault() == nil {
			verr.AddViolationAt(defaultField, "must be defined")
		} else {
			verr.AddErrorAt(defaultField, validateDefault(toItem.Default))
		}
	}
	return verr
}

func validateDefault(conf *DoNothingPolicy_Conf) validators.ValidationError {
	var verr validators.ValidationError
	// TODO add default conf validation
	return verr
}
