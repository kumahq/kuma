package mesh

import (
	"github.com/Kong/kuma/pkg/core/validators"
)

func (d *TrafficRouteResource) Validate() error {
	var err validators.ValidationError
	err.Add(d.validateSources())
	err.Add(d.validateDestinations())
	err.Add(d.validateConf())
	return err.OrNil()
}

func (d *TrafficRouteResource) validateSources() validators.ValidationError {
	return ValidateSelectors(validators.RootedAt("sources"), d.Spec.Sources, ValidateSelectorsOpts{
		RequireAtLeastOneSelector: true,
		ValidateSelectorOpts: ValidateSelectorOpts{
			RequireAtLeastOneTag: true,
			RequireService:       true,
		},
	})
}

func (d *TrafficRouteResource) validateDestinations() (err validators.ValidationError) {
	return ValidateSelectors(validators.RootedAt("destinations"), d.Spec.Destinations, OnlyServiceTagAllowed)
}

func (d *TrafficRouteResource) validateConf() (err validators.ValidationError) {
	root := validators.RootedAt("conf")
	if len(d.Spec.Conf) == 0 {
		err.AddViolationAt(root, "must have at least one element")
	}
	for i, routeEntry := range d.Spec.Conf {
		err.Add(ValidateSelector(root.Index(i).Field("destination"), routeEntry.GetDestination(), ValidateSelectorOpts{
			RequireAtLeastOneTag: true,
			RequireService:       true,
		}))
	}
	return
}
