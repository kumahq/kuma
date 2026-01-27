package v1alpha1

import (
	"github.com/kumahq/kuma/v2/pkg/core/resources/model"
	"github.com/kumahq/kuma/v2/pkg/core/validators"
)

func (r *MeshServiceResource) validate() error {
	var verr validators.ValidationError

	name := model.GetDisplayName(r.GetMeta())
	verr.Add(validators.ValidateLength(validators.RootedAt("name"), maxNameLength, name))

	// Validate selector mutual exclusivity
	count := 0
	if r.Spec.Selector.DataplaneTags != nil {
		count++
	}
	if r.Spec.Selector.DataplaneRef != nil {
		count++
	}
	if r.Spec.Selector.DataplaneLabels != nil {
		count++
	}

	if count > 1 {
		verr.AddViolationAt(validators.RootedAt("spec").Field("selector"), "must specify only one of: dataplaneTags, dataplaneRef, or dataplaneLabels")
	}

	return verr.OrNil()
}
