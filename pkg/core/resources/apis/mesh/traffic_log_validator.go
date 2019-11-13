package mesh

import (
	"fmt"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	"github.com/Kong/kuma/pkg/core/validators"
)

func (d *TrafficLogResource) Validate() error {
	var err validators.ValidationError
	err.Add(d.validateSources())
	err.Add(d.validateDestinations())
	return err.OrNil()
}

func (d *TrafficLogResource) validateSources() validators.ValidationError {
	return ValidateSelectors(validators.RootedAt("sources"), d.Spec.Sources, ValidateSelectorOpts{})
}

func (d *TrafficLogResource) validateDestinations() (err validators.ValidationError) {
	return ValidateSelectors(validators.RootedAt("destinations"), d.Spec.Destinations, ValidateSelectorOpts{
		SkipRequireAtLeastOneTag: true,
		ExtraSelectorValidators: []SelectorValidatorFunc{
			func(path validators.PathBuilder, selector map[string]string) (err validators.ValidationError) {
				_, defined := selector[mesh_proto.ServiceTag]
				if len(selector) != 1 || !defined {
					err.AddViolationAt(path, fmt.Sprintf("must consist of exactly one tag %q", mesh_proto.ServiceTag))
				}
				return
			},
		},
		ExtraTagKeyValidators: []TagKeyValidatorFunc{
			func(path validators.PathBuilder, key string) (err validators.ValidationError) {
				if key != mesh_proto.ServiceTag {
					err.AddViolationAt(path.Key(key), fmt.Sprintf("tag %q is not allowed", key))
				}
				return
			},
		},
	})
}
