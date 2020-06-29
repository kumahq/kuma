package mesh

import (
	"github.com/Kong/kuma/pkg/core/validators"
)

func (d *HealthCheckResource) Validate() error {
	var err validators.ValidationError
	err.Add(d.validateSources())
	err.Add(d.validateDestinations())
	err.Add(d.validateConf())
	return err.OrNil()
}

func (d *HealthCheckResource) validateSources() validators.ValidationError {
	return ValidateSelectors(validators.RootedAt("sources"), d.Spec.Sources, ValidateSelectorsOpts{
		RequireAtLeastOneSelector: true,
		ValidateSelectorOpts: ValidateSelectorOpts{
			RequireAtLeastOneTag: true,
			RequireService:       true,
		},
	})
}

func (d *HealthCheckResource) validateDestinations() (err validators.ValidationError) {
	return ValidateSelectors(validators.RootedAt("destinations"), d.Spec.Destinations, OnlyServiceTagAllowed)
}

func (d *HealthCheckResource) validateConf() (err validators.ValidationError) {
	path := validators.RootedAt("conf")
	if d.Spec.GetConf() == nil {
		err.AddViolationAt(path, "has to be defined")
		return
	}
	err.Add(ValidateDuration(path.Field("interval"), d.Spec.Conf.Interval))
	err.Add(ValidateDuration(path.Field("timeout"), d.Spec.Conf.Timeout))
	err.Add(ValidateThreshold(path.Field("unhealthyThreshold"), d.Spec.Conf.UnhealthyThreshold))
	err.Add(ValidateThreshold(path.Field("healthyThreshold"), d.Spec.Conf.HealthyThreshold))
	return
}
