package v3

import (
	envoy_route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"

	"github.com/kumahq/kuma/pkg/xds/envoy/tags"
)

type ResetTagsHeaderConfigurer struct{}

func (r *ResetTagsHeaderConfigurer) Configure(rc *envoy_route.RouteConfiguration) error {
	rc.RequestHeadersToRemove = append(rc.RequestHeadersToRemove, tags.TagsHeaderName)
	return nil
}
