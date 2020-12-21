package v2

import envoy_api "github.com/envoyproxy/go-control-plane/envoy/api/v2"

type ResetTagsHeaderConfigurer struct {
}

func (r *ResetTagsHeaderConfigurer) Configure(rc *envoy_api.RouteConfiguration) error {
	rc.RequestHeadersToRemove = append(rc.RequestHeadersToRemove, TagsHeaderName)
	return nil
}
