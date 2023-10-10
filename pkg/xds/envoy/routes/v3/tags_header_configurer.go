package v3

import (
	envoy_config_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_config_route_v3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/xds/envoy/tags"
)

type TagsHeaderConfigurer struct {
	Tags mesh_proto.MultiValueTagSet
}

func (t *TagsHeaderConfigurer) Configure(rc *envoy_config_route_v3.RouteConfiguration) error {
	if len(t.Tags) == 0 {
		return nil
	}
	rc.RequestHeadersToAdd = append(rc.RequestHeadersToAdd, &envoy_config_core_v3.HeaderValueOption{
		Header: &envoy_config_core_v3.HeaderValue{Key: tags.TagsHeaderName, Value: tags.Serialize(t.Tags)},
	})
	return nil
}
