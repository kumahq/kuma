package v3

import (
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/xds/envoy/tags"
)

type TagsHeaderConfigurer struct {
	Tags mesh_proto.MultiValueTagSet
}

func (t *TagsHeaderConfigurer) Configure(rc *envoy_route.RouteConfiguration) error {
	if len(t.Tags) == 0 {
		return nil
	}
	rc.RequestHeadersToAdd = append(rc.RequestHeadersToAdd, &envoy_core.HeaderValueOption{
		Header: &envoy_core.HeaderValue{Key: tags.TagsHeaderName, Value: tags.Serialize(t.Tags)},
	})
	return nil
}
