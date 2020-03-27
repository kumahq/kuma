package routes

import envoy_api_v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"

func ResetTags() RouteConfigurationBuilderOpt {
	return RouteConfigurationBuilderOptFunc(func(config *RouteConfigurationBuilderConfig) {
		config.Add(&ResetTagsConfigurer{})
	})
}

type ResetTagsConfigurer struct {
}

func (r *ResetTagsConfigurer) Configure(rc *envoy_api_v2.RouteConfiguration) error {
	rc.RequestHeadersToRemove = append(rc.RequestHeadersToRemove, TagsHeader)
	return nil
}
