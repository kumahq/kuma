package mesh

import (
	"fmt"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
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
	return ValidateSelectors(validators.RootedAt("sources"), d.Spec.Sources, ValidateSelectorOpts{})
}

func (d *TrafficRouteResource) validateDestinations() (err validators.ValidationError) {
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

func (d *TrafficRouteResource) validateConf() (err validators.ValidationError) {
	root := validators.RootedAt("conf")
	if len(d.Spec.Conf) == 0 {
		err.AddViolationAt(root, "must have at least one element")
	}
	for i, routeEntry := range d.Spec.Conf {
		err.Add(ValidateSelector(root.Index(i).Field("destination"), routeEntry.GetDestination(), ValidateSelectorOpts{}))
	}
	return
}
