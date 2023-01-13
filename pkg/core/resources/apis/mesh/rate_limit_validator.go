package mesh

import (
	"fmt"

	"github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/validators"
)

func (d *RateLimitResource) Validate() error {
	var err validators.ValidationError
	err.Add(d.validateSources())
	err.Add(d.validateDestinations())
	err.Add(d.validateConf())
	return err.OrNil()
}

func (d *RateLimitResource) validateSources() validators.ValidationError {
	return ValidateSelectors(validators.RootedAt("sources"), d.Spec.Sources, ValidateSelectorsOpts{
		RequireAtLeastOneSelector: true,
		ValidateTagsOpts: ValidateTagsOpts{
			RequireAtLeastOneTag: true,
		},
	})
}

func (d *RateLimitResource) validateDestinations() validators.ValidationError {
	return ValidateSelectors(validators.RootedAt("destinations"), d.Spec.Destinations, ValidateSelectorsOpts{
		RequireAtLeastOneSelector: true,
		ValidateTagsOpts: ValidateTagsOpts{
			RequireAtLeastOneTag: true,
		},
	})
}

func (d *RateLimitResource) validateConf() validators.ValidationError {
	var err validators.ValidationError
	root := validators.RootedAt("conf")
	if d.Spec.GetConf() == nil {
		err.AddViolationAt(root, "must have conf")
	}

	if d.Spec.GetConf().GetHttp() != nil {
		err.Add(d.validateHttp(root.Field("http"), d.Spec.GetConf().GetHttp()))
	}

	return err
}

func (d *RateLimitResource) validateHttp(path validators.PathBuilder, http *v1alpha1.RateLimit_Conf_Http) validators.ValidationError {
	var err validators.ValidationError
	if http.GetRequests() == 0 {
		err.AddViolationAt(path.Field("requests"), "requests must be set")
	}

	if http.GetInterval() == nil {
		err.AddViolationAt(path.Field("interval"), "interval must be set")
	}

	if http.GetOnRateLimit() != nil {
		err.Add(d.validateOnRateLimit(path.Field("onRateLimit"), http.GetOnRateLimit()))
	}

	return err
}

func (d *RateLimitResource) validateOnRateLimit(path validators.PathBuilder, onRateLimit *v1alpha1.RateLimit_Conf_Http_OnRateLimit) validators.ValidationError {
	var err validators.ValidationError
	for i, h := range onRateLimit.GetHeaders() {
		if h.Key == "" {
			err.AddViolationAt(path.Field("header").Key(fmt.Sprintf("%d", i)), "key must be set")
		}
		if h.Value == "" {
			err.AddViolationAt(path.Field("header").Key(fmt.Sprintf("%d", i)), "value must be set")
		}
	}
	return err
}
