package routes

import envoy_api_v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"

// ResetTagsHeader adds x-kuma-tags header to the RequestHeadersToRemove list. x-kuma-tags header is planned to be used
// internally, so we don't want to expose it to the destination application.
func ResetTagsHeader() RouteConfigurationBuilderOpt {
	return RouteConfigurationBuilderOptFunc(func(config *RouteConfigurationBuilderConfig) {
		config.Add(&ResetTagsHeaderConfigurer{})
	})
}

type ResetTagsHeaderConfigurer struct {
}

func (r *ResetTagsHeaderConfigurer) Configure(rc *envoy_api_v2.RouteConfiguration) error {
	rc.RequestHeadersToRemove = append(rc.RequestHeadersToRemove, TagsHeaderName)
	return nil
}
