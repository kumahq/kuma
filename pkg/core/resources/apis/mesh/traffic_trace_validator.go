package mesh

import (
	"github.com/kumahq/kuma/pkg/core/validators"
)

func (d *TrafficTraceResource) Validate() error {
	var err validators.ValidationError
	err.Add(d.validateSelectors())
	// d.Spec.Conf and d.Spec.Conf.DefaultBackend can be empty, then default backend of the mesh is chosen.
	return err.OrNil()
}

func (d *TrafficTraceResource) validateSelectors() validators.ValidationError {
	return ValidateSelectors(validators.RootedAt("selectors"), d.Spec.GetSelectors(), ValidateSelectorsOpts{
		RequireAtLeastOneSelector: true,
		ValidateTagsOpts: ValidateTagsOpts{
			RequireAtLeastOneTag: true,
		},
	})
}
