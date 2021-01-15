package mesh

import (
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/validators"
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
	if d.Spec.GetConf().GetSplit() == nil {
		err.AddViolationAt(root, "must have split")
	}

	root = validators.RootedAt("conf.split")
	if len(d.Spec.GetConf().GetSplit()) == 0 {
		err.AddViolationAt(root, "must have at least one element")
	}

	for i, routeEntry := range d.Spec.GetConf().GetSplit() {
		err.Add(ValidateSelector(root.Index(i).Field("destination"), routeEntry.GetDestination(), ValidateSelectorOpts{
			RequireAtLeastOneTag: true,
			RequireService:       true,
		}))
	}
	err.Add(d.validateLb())

	return
}

func (d *TrafficRouteResource) validateLb() (err validators.ValidationError) {
	lb := d.Spec.GetConf().GetLoadBalancer()
	if lb == nil {
		return
	}

	switch lb.LbType.(type) {
	case *mesh_proto.TrafficRoute_LoadBalancer_LeastRequest_:

	case *mesh_proto.TrafficRoute_LoadBalancer_RingHash_:
		lbConfig := lb.GetRingHash()
		switch lbConfig.HashFunction {
		case "XX_HASH", "MURMUR_HASH_2":
		default:
			root := validators.RootedAt("conf.loadBalancer.ringHash.hashFunction")
			err.AddViolationAt(root, "must have a valid hash function")
		}
	}
	return
}
