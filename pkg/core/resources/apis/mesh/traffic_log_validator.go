package mesh

import (
	"github.com/Kong/kuma/pkg/core/validators"
)

func (d *TrafficLogResource) Validate() error {
	var err validators.ValidationError
	err.Add(d.validateSources())
	err.Add(d.validateDestinations())
	// d.Spec.Conf and d.Spec.Conf.DefaultBackend can be empty, then default backend of the mesh is chosen.
	return err.OrNil()
}

func (d *TrafficLogResource) validateSources() validators.ValidationError {
	return ValidateSelectors(validators.RootedAt("sources"), d.Spec.Sources, ValidateSelectorOpts{})
}

func (d *TrafficLogResource) validateDestinations() (err validators.ValidationError) {
	return ValidateSelectors(validators.RootedAt("destinations"), d.Spec.Destinations, OnlyServiceTagAllowed)
}
