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
		ValidateTagsOpts: ValidateTagsOpts{
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
	err.Add(d.validateHTTPModify(pathBuilder.Field("modify"), http.GetModify(), http.GetMatch()))
	err.Add(d.validateSplitAndDestination(pathBuilder, http.GetSplit(), http.GetDestination()))
	return
}

func (d *TrafficRouteResource) validateHTTPModify(
	pathBuilder validators.PathBuilder,
	modify *mesh_proto.TrafficRoute_Http_Modify,
	match *mesh_proto.TrafficRoute_Http_Match,
) (err validators.ValidationError) {
	if modify.GetPath() != nil {
		err.Add(d.validateModificationPath(pathBuilder.Field("path"), modify.GetPath(), match))
	}
	if modify.GetHost() != nil {
		err.Add(d.validateModificationHost(pathBuilder.Field("host"), modify.GetHost()))
	}
	if modify.GetRequestHeaders() != nil {
		err.Add(d.validateModificationHeaders(pathBuilder.Field("requestHeaders"), modify.GetRequestHeaders()))
	}
	if modify.GetResponseHeaders() != nil {
		err.Add(d.validateModificationHeaders(pathBuilder.Field("responseHeaders"), modify.GetResponseHeaders()))
	}
	return
}

func (d *TrafficRouteResource) validateModificationPath(
	pathBuilder validators.PathBuilder,
	path *mesh_proto.TrafficRoute_Http_Modify_Path,
	match *mesh_proto.TrafficRoute_Http_Match,
) (err validators.ValidationError) {
	switch path.Type.(type) {
	case *mesh_proto.TrafficRoute_Http_Modify_Path_RewritePrefix:
		if path.GetRewritePrefix() == "" {
			err.AddViolationAt(pathBuilder.Field("rewritePrefix"), "cannot be empty")
		}
		if path.GetRewritePrefix() != "" && match.GetPath().GetPrefix() == "" {
			err.AddViolationAt(pathBuilder.Field("rewritePrefix"), "can only be set when .http.match.path.prefix is not empty")
		}
	case *mesh_proto.TrafficRoute_Http_Modify_Path_Regex:
		if path.GetRegex().GetPattern() == "" {
			err.AddViolationAt(pathBuilder.Field("regex").Field("pattern"), "cannot be empty")
		}
		if path.GetRegex().GetSubstitution() == "" {
			err.AddViolationAt(pathBuilder.Field("regex").Field("substitution"), "cannot be empty")
		}
	default:
		err.AddViolationAt(pathBuilder, `either "rewritePrefix" or "regex" has to be set`)
	}
	return
}

func (d *TrafficRouteResource) validateModificationHost(
	pathBuilder validators.PathBuilder,
	host *mesh_proto.TrafficRoute_Http_Modify_Host,
) (err validators.ValidationError) {
	switch host.Type.(type) {
	case *mesh_proto.TrafficRoute_Http_Modify_Host_Value:
		if host.GetValue() == "" {
			err.AddViolationAt(pathBuilder.Field("value"), "cannot be empty")
		}
	case *mesh_proto.TrafficRoute_Http_Modify_Host_FromPath:
		if host.GetFromPath().GetPattern() == "" {
			err.AddViolationAt(pathBuilder.Field("fromPath").Field("pattern"), "cannot be empty")
		}
		if host.GetFromPath().GetSubstitution() == "" {
			err.AddViolationAt(pathBuilder.Field("fromPath").Field("substitution"), "cannot be empty")
		}
	default:
		err.AddViolationAt(pathBuilder, `either "value" or "fromPath" has to be set`)
	}
	return
}

func (d *TrafficRouteResource) validateModificationHeaders(
	pathBuilder validators.PathBuilder,
	headers *mesh_proto.TrafficRoute_Http_Modify_Headers,
) (err validators.ValidationError) {
	for i, add := range headers.GetAdd() {
		err.Add(validateHeaderName(pathBuilder.Field("add").Index(i).Field("name"), add.GetName()))
		if add.GetValue() == "" {
			err.AddViolationAt(pathBuilder.Field("add").Index(i).Field("value"), "cannot be empty")
		}
	}
	for i, remove := range headers.GetRemove() {
		err.Add(validateHeaderName(pathBuilder.Field("remove").Index(i).Field("name"), remove.GetName()))
	}
	return
}

func validateHeaderName(pathBuilder validators.PathBuilder, headerName string) (err validators.ValidationError) {
	if headerName == "" {
		err.AddViolationAt(pathBuilder, "cannot be empty")
		return
	}
	if headerName[0] == ':' || headerName == "host" || headerName == "Host" {
		err.AddViolationAt(pathBuilder, "host header and HTTP/2 pseudo-headers are not allowed to be modified")
	}
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
		err.Add(ValidateSelector(pathBuilder.Index(i).Field("destination"), routeEntry.GetDestination(), ValidateTagsOpts{
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
	return ValidateSelector(pathBuilder, destination, ValidateTagsOpts{
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
