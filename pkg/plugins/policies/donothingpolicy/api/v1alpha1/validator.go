package v1alpha1

import (
	"github.com/kumahq/kuma/pkg/core/validators"
	matcher_validators "github.com/kumahq/kuma/pkg/plugins/policies/matchers/validators"
)

func (r *DoNothingPolicyResource) validate() error {
	var err validators.ValidationError
	matcher_validators.Validate(err, r)
	err.AddViolationAt(validators.RootedAt(""), "Not implemented!")
	return err.OrNil()
}
