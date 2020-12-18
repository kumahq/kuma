package v2

import (
	envoy_api "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"

	"github.com/kumahq/kuma/pkg/xds/envoy/tags"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
)

const TagsHeaderName = "x-kuma-tags"

type TagsHeaderConfigurer struct {
	Tags mesh_proto.MultiValueTagSet
}

func (t *TagsHeaderConfigurer) Configure(rc *envoy_api.RouteConfiguration) error {
	if len(t.Tags) == 0 {
		return nil
	}
	rc.RequestHeadersToAdd = append(rc.RequestHeadersToAdd, &envoy_core.HeaderValueOption{
		Header: &envoy_core.HeaderValue{Key: TagsHeaderName, Value: tags.Serialize(t.Tags)},
	})
	return nil
}
