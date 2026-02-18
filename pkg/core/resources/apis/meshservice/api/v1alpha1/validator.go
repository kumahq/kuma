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
	var setSelectors []bool
	if r.Spec.Selector.DataplaneTags != nil {
		setSelectors = append(setSelectors, true)
	}
	if r.Spec.Selector.DataplaneRef != nil {
		setSelectors = append(setSelectors, true)
	}
	if r.Spec.Selector.DataplaneLabels != nil {
		setSelectors = append(setSelectors, true)
	}

	if len(setSelectors) > 1 {
		verr.AddViolationAt(validators.RootedAt("spec").Field("selector"), "must specify only one of: dataplaneTags, dataplaneRef, or dataplaneLabels")
	}

	return verr.OrNil()
}
