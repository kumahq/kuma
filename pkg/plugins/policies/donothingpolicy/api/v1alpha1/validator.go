package v1alpha1

import (
	"github.com/kumahq/kuma/pkg/core/validators"
)

func (r *DoNothingPolicyResource) validate() error {
	var err validators.ValidationError
	err.AddViolationAt(validators.RootedAt(""), "Not implemented!")
	return err.OrNil()
}
