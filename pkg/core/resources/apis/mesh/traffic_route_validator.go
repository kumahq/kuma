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
	if d.Spec.GetConf() == nil {
		err.AddViolationAt(root, "cannot be empty")
		return
	}

	err.Add(d.validateSplitAndDestination(root, d.Spec.GetConf().GetSplit(), d.Spec.GetConf().GetDestination()))
	for i, http := range d.Spec.GetConf().GetHttp() {
		err.Add(d.validateHTTP(root.Field("http").Index(i), http))
	}
	err.Add(d.validateLb())
	return
}

func (d *TrafficRouteResource) validateHTTP(pathBuilder validators.PathBuilder, http *mesh_proto.TrafficRoute_Http) (err validators.ValidationError) {
	err.Add(d.validateHTTPMatch(pathBuilder.Field("match"), http.GetMatch()))
	err.Add(d.validateSplitAndDestination(pathBuilder, http.GetSplit(), http.GetDestination()))
	return
}

func (d *TrafficRouteResource) validateSplitAndDestination(pathBuilder validators.PathBuilder, split []*mesh_proto.TrafficRoute_Split, destination map[string]string) (err validators.ValidationError) {
	if split == nil && destination == nil {
		err.AddViolationAt(pathBuilder, `requires either "destination" or "split"`)
	}
	if split != nil && destination != nil {
		err.AddViolationAt(pathBuilder, `"split" cannot be defined at the same time as "destination"`)
	}

	if split != nil {
		err.Add(d.validateSplit(pathBuilder.Field("split"), split))
	}
	if destination != nil {
		err.Add(d.validateDestination(pathBuilder.Field("destination"), destination))
	}
	return
}

func (d *TrafficRouteResource) validateHTTPMatch(pathBuilder validators.PathBuilder, match *mesh_proto.TrafficRoute_Http_Match) (err validators.ValidationError) {
	if match.GetPath() == nil && match.GetMethod() == nil && match.GetHeaders() == nil {
		err.AddViolationAt(pathBuilder, `must be present and contain at least one of the elements: "method", "path" or "headers"`)
		return
	}
	if match.GetMethod() != nil {
		err.Add(d.validateStringMatcher(pathBuilder.Field("method"), match.GetMethod()))
	}
	if match.GetPath() != nil {
		err.Add(d.validateStringMatcher(pathBuilder.Field("path"), match.GetPath()))
	}
	if match.GetHeaders() != nil && len(match.GetHeaders()) == 0 {
		err.AddViolationAt(pathBuilder.Field("headers"), "must contain at least one element")
	}
	for key, matcher := range match.GetHeaders() {
		path := pathBuilder.Field("headers").Key(key)
		if len(key) == 0 {
			err.AddViolationAt(path, "cannot be empty")
		}
		err.Add(d.validateStringMatcher(path, matcher))
	}
	return
}

func (d *TrafficRouteResource) validateStringMatcher(pathBuilder validators.PathBuilder, matcher *mesh_proto.TrafficRoute_Http_Match_StringMatcher) (err validators.ValidationError) {
	switch matcher.GetMatcherType().(type) {
	case *mesh_proto.TrafficRoute_Http_Match_StringMatcher_Exact:
	case *mesh_proto.TrafficRoute_Http_Match_StringMatcher_Prefix:
		if len(matcher.GetPrefix()) == 0 {
			err.AddViolationAt(pathBuilder.Field("prefix"), `cannot be empty`)
		}
	case *mesh_proto.TrafficRoute_Http_Match_StringMatcher_Regex:
		if len(matcher.GetRegex()) == 0 {
			err.AddViolationAt(pathBuilder.Field("regex"), `cannot be empty`)
		}
	default:
		err.AddViolationAt(pathBuilder, `cannot be empty. Available options: "exact", "split" or "regex"`)
	}
	return
}

func (d *TrafficRouteResource) validateSplit(pathBuilder validators.PathBuilder, split []*mesh_proto.TrafficRoute_Split) (err validators.ValidationError) {
	if len(split) == 0 {
		err.AddViolationAt(pathBuilder, "must have at least one element")
		return
	}
	var totalWeight uint32
	for i, routeEntry := range split {
		if routeEntry.GetWeight() == nil {
			err.AddViolationAt(pathBuilder.Index(i).Field("weight"), "needs to be defined")
		}
		totalWeight += routeEntry.GetWeight().GetValue()
		err.Add(ValidateSelector(pathBuilder.Index(i).Field("destination"), routeEntry.GetDestination(), ValidateSelectorOpts{
			RequireAtLeastOneTag: true,
			RequireService:       true,
		}))
	}
	if totalWeight == 0 {
		err.AddViolationAt(pathBuilder, "there must be at least one split entry with weight above 0")
	}
	return
}

func (d *TrafficRouteResource) validateDestination(pathBuilder validators.PathBuilder, destination map[string]string) validators.ValidationError {
	return ValidateSelector(pathBuilder, destination, ValidateSelectorOpts{
		RequireAtLeastOneTag: true,
		RequireService:       true,
	})
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
