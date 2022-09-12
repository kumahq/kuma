package v1alpha1

import (
	common_proto "github.com/kumahq/kuma/api/common/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/validators"
	matcher_validators "github.com/kumahq/kuma/pkg/plugins/policies/matchers/validators"
)

func (r *DoNothingPolicyResource) validate() error {
	var verr validators.ValidationError
	path := validators.RootedAt("spec")

	targetRefErr := matcher_validators.ValidateTargetRef(path.Field("targetRef"), r.Spec.GetTargetRef(), &matcher_validators.ValidateTargetRefOpts{
		SupportedKinds: []common_proto.TargetRef_Kind{
			// TODO add supported TargetRef kinds for this policy
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
					// TODO add supported TargetRef for 'from' item
				},
			})
			verr.AddError("", targetRefErr)

			defaultField := from.Index(idx).Field("default")
			if fromItem.GetDefault() == nil {
				verr.AddViolationAt(defaultField, "cannot be nil")
			} else {
				// TODO add default conf validation
				verr.AddViolationAt(defaultField, "")
			}
		}
	}

	to := path.Field("to")
	if len(r.Spec.GetTo()) == 0 {
		verr.AddViolationAt(to, "cannot be empty")
	} else {
		for idx, toItem := range r.Spec.GetTo() {
			targetRefErr := matcher_validators.ValidateTargetRef(from.Index(idx).Field("targetRef"), toItem.GetTargetRef(), &matcher_validators.ValidateTargetRefOpts{
				SupportedKinds: []common_proto.TargetRef_Kind{
					// TODO add supported TargetRef for 'to' item
				},
			})
			verr.AddError("", targetRefErr)

			defaultField := to.Index(idx).Field("default")
			if toItem.GetDefault() == nil {
				verr.AddViolationAt(defaultField, "cannot be nil")
			} else {
				// TODO add default conf validation
				verr.AddViolationAt(defaultField, "")
			}
		}
	}

	return verr.OrNil()
}
