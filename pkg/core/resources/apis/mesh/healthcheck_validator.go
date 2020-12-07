package mesh

import (
	"github.com/kumahq/kuma/pkg/core/validators"
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

func (d *HealthCheckResource) validateConfHttpPath(
	path validators.PathBuilder,
) (err validators.ValidationError) {
	if d.Spec.Conf.Http.Path == nil {
		err.AddViolationAt(path, "has to be defined")
	} else if d.Spec.Conf.Http.Path.Value == "" {
		err.AddViolationAt(path, "cannot be empty")
	}

	return
}

func (d *HealthCheckResource) validateConfHttpExpectedStatuses(
	path validators.PathBuilder,
) (err validators.ValidationError) {
	if d.Spec.Conf.Http.ExpectedStatuses != nil &&
		len(d.Spec.Conf.Http.ExpectedStatuses) == 0 {
		err.AddViolationAt(path, "cannot be empty")
		return
	}

	for i, status := range d.Spec.Conf.Http.ExpectedStatuses {
		if status.Value < 100 || status.Value >= 600 {
			err.AddViolationAt(
				path.Index(i),
				"must be in range [100, 600)",
			)
		}
	}

	return
}

func (d *HealthCheckResource) validateConfHttp(
	path validators.PathBuilder,
) (err validators.ValidationError) {
	err.Add(d.validateConfHttpPath(path.Field("path")))
	err.Add(d.validateConfHttpExpectedStatuses(path.Field("expected_statuses")))
	return
}

func (d *HealthCheckResource) validateConf() (err validators.ValidationError) {
	path := validators.RootedAt("conf")
	if d.Spec.GetConf() == nil {
		err.AddViolationAt(path, "has to be defined")
		return
	}
	if d.Spec.Conf.Tcp != nil && d.Spec.Conf.Http != nil {
		err.AddViolationAt(path.Field("[tcp|http]"), "only one allowed")
	}
	err.Add(ValidateDuration(path.Field("interval"), d.Spec.Conf.Interval))
	err.Add(ValidateDuration(path.Field("timeout"), d.Spec.Conf.Timeout))
	err.Add(ValidateThreshold(path.Field("unhealthyThreshold"), d.Spec.Conf.UnhealthyThreshold))
	err.Add(ValidateThreshold(path.Field("healthyThreshold"), d.Spec.Conf.HealthyThreshold))
	if d.Spec.Conf.Http != nil {
		err.Add(d.validateConfHttp(path.Field("http")))
	}
	return
}
