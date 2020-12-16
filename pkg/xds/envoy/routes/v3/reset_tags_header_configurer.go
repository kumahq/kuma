package v3

import envoy_route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"

type ResetTagsHeaderConfigurer struct {
}

func (r *ResetTagsHeaderConfigurer) Configure(rc *envoy_route.RouteConfiguration) error {
	rc.RequestHeadersToRemove = append(rc.RequestHeadersToRemove, TagsHeaderName)
	return nil
}
