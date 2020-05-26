package mesh

import (
	"github.com/golang/protobuf/ptypes/wrappers"

	"github.com/Kong/kuma/pkg/core/validators"
)

func (c *CircuitBreakerResource) HasDetectors() bool {
	return c.Spec.Conf.GetDetectors().GetTotalErrors() != nil ||
		c.Spec.Conf.GetDetectors().GetGatewayErrors() != nil ||
		c.Spec.Conf.GetDetectors().GetLocalErrors() != nil ||
		c.Spec.Conf.GetDetectors().GetStandardDeviation() != nil ||
		c.Spec.Conf.GetDetectors().GetFailure() != nil
}

func (c *CircuitBreakerResource) Validate() error {
	var err validators.ValidationError
	err.Add(c.validateSources())
	err.Add(c.validateDestinations())
	err.Add(c.validateConf())
	return err.OrNil()
}

func (c *CircuitBreakerResource) validateSources() validators.ValidationError {
	return ValidateSelectors(validators.RootedAt("sources"), c.Spec.Sources, ValidateSelectorsOpts{
		RequireAtLeastOneSelector: true,
		ValidateSelectorOpts: ValidateSelectorOpts{
			RequireAtLeastOneTag: true,
			RequireService:       true,
		},
	})
}

func (c *CircuitBreakerResource) validateDestinations() validators.ValidationError {
	return ValidateSelectors(validators.RootedAt("destinations"), c.Spec.Destinations, OnlyServiceTagAllowed)
}

func (c *CircuitBreakerResource) validateConf() (err validators.ValidationError) {
	root := validators.RootedAt("conf")
	if !c.HasDetectors() {
		err.AddViolationAt(root, "must have at least one of the detectors configured")
	}
	err.Add(c.validatePercentage(root.Field("maxEjectionPercent"), c.Spec.GetConf().GetMaxEjectionPercent()))

	path := root.Field("detectors")
	if failure := c.Spec.Conf.GetDetectors().GetFailure(); failure != nil {
		err.Add(c.validatePercentage(path.Field("failure").Field("threshold"), failure.GetThreshold()))
	}
	return
}

func (c *CircuitBreakerResource) validatePercentage(path validators.PathBuilder, value *wrappers.UInt32Value) (err validators.ValidationError) {
	if value.GetValue() < 0.0 || value.GetValue() > 100.0 {
		err.AddViolationAt(path, "has to be in [0.0 - 100.0] range")
	}
	return
}
