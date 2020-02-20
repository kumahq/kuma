package mesh

import (
	"github.com/Kong/kuma/pkg/core/validators"
)

func (d *TrafficLogResource) Validate() error {
	var err validators.ValidationError
	err.Add(d.validateSelectors())
	// d.Spec.Conf and d.Spec.Conf.DefaultBackend can be empty, then default backend of the mesh is chosen.
	return err.OrNil()
}

func (d *TrafficLogResource) validateSelectors() validators.ValidationError {
	return ValidateSelectors(validators.RootedAt("selectors"), d.Spec.Selectors, ValidateSelectorsOpts{
		RequireAtLeastOneSelector: true,
		ValidateSelectorOpts: ValidateSelectorOpts{
			RequireAtLeastOneTag: true,
		},
	})
}
