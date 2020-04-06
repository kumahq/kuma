package routes

import (
	envoy_api_v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"

	"github.com/Kong/kuma/pkg/xds/envoy/tags"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
)

const TagsHeaderName = "x-kuma-tags"

func TagsHeader(tags mesh_proto.MultiValueTagSet) RouteConfigurationBuilderOpt {
	return RouteConfigurationBuilderOptFunc(func(config *RouteConfigurationBuilderConfig) {
		config.Add(&TagsHeaderConfigurer{
			tags: tags,
		})
	})
}

type TagsHeaderConfigurer struct {
	tags mesh_proto.MultiValueTagSet
}

func (t *TagsHeaderConfigurer) Configure(rc *envoy_api_v2.RouteConfiguration) error {
	if len(t.tags) == 0 {
		return nil
	}
	rc.RequestHeadersToAdd = append(rc.RequestHeadersToAdd, &envoy_core.HeaderValueOption{
		Header: &envoy_core.HeaderValue{Key: TagsHeaderName, Value: tags.Serialize(t.tags)},
	})
	return nil
}
