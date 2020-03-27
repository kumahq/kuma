package routes

import (
	"github.com/Kong/kuma/api/mesh/v1alpha1"
	envoy_api_v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
)

const TagsHeader = "x-kuma-tags"

func Tags(tags v1alpha1.MultiValueTagSet) RouteConfigurationBuilderOpt {
	return RouteConfigurationBuilderOptFunc(func(config *RouteConfigurationBuilderConfig) {
		config.Add(&TagsConfigurer{
			tags: tags,
		})
	})
}

type TagsConfigurer struct {
	tags v1alpha1.MultiValueTagSet
}

func (t *TagsConfigurer) Configure(rc *envoy_api_v2.RouteConfiguration) error {
	if len(t.tags) == 0 {
		return nil
	}
	rc.RequestHeadersToAdd = append(rc.RequestHeadersToAdd, &envoy_core.HeaderValueOption{
		Header: &envoy_core.HeaderValue{Key: TagsHeader, Value: t.tags.String()},
	})
	return nil
}
